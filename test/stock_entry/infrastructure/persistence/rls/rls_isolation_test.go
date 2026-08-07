//go:build integration

// Test de aislamiento cross-tenant adversarial (PLAT-E30 T-STK7).
//
// Este es el criterio REAL de fail-closed de la épica: no que la migración 009 corra, sino que
// un tenant jamás vea ni escriba filas de otro bajo RLS, en las 5 tablas de stock:
// locations, warehouses, stock_locations, stock_entries y stock_availability.
//
// Levanta un Postgres efímero (testcontainers), aplica TODAS las migraciones de esquema reales
// del servicio (001, 004, 005, 006, 007, 008, 009, 010 — el subdir migrations/roles/ queda
// afuera vía e.IsDir()) y conecta bajo un rol sin BYPASSRLS que replica stock_app (creado en
// lab-postgres por PLAT-E30 T-STK3). Un superuser SIEMPRE bypasea RLS aunque la tabla tenga
// FORCE ROW LEVEL SECURITY, así que probar contra el usuario `postgres` del contenedor daría
// falsos verdes (gotcha probado empíricamente en E24). TODAS las aserciones de persistencia
// corren bajo stock_app_test (NOBYPASSRLS) → las policies tenant_isolation de la 009 (forma
// unificada `::uuid` tras 010) se ejercen de verdad.
//
// Cobertura (T-STK7 "Hecho cuando", patrón E25–E29 adaptado a las 5 tablas de stock-service):
//
//	(a) Save/Insert de A bajo RLSContext{A} pasa el WITH CHECK.
//	(b) los datos de A NO son visibles bajo RLSContext{B}: la consulta cruda por PK devuelve 0.
//	(c) INSERT con tenant_id=A bajo sesión B → rechazado por WITH CHECK.
//	(d) UPDATE cruda cross-tenant de B sobre filas de A → 0 filas, fila intacta.
//	(e) fail-closed sin contexto (M1): SELECT/INSERT/UPDATE directos a las 5 tablas sin
//	    app.tenant_id fijado erran o devuelven 0 filas — lecturas Y escrituras.
//	(f) reinterpretada (caveat T-STK3): no hay triggers ni funciones trigger vivas en stock_db
//	    (la migración 008 las dropeó y movió el recálculo a Go `recalcAvailability`). Se prueba
//	    que `Save` de un StockEntry corre `recalcAvailability` dentro de la misma tx RLS bajo
//	    stock_app NOBYPASSRLS y produce una fila de stock_availability aislada para el tenant
//	    del caller — la lógica de negocio viva que C2 pedía ejercer, hoy en el repositorio.
//	(g) las 3 tablas normalizadas en T-STK2b (locations, warehouses, stock_locations) se
//	    cubren con datos reales persistentes (no seed transaccional revertido) — condición del
//	    sign-off de T-STK2: hoy están vacías y su aislamiento nunca se probó de verdad.
//	(h) un JWT válido de tenant A con X-Tenant-ID: B → 403 y cero filas de B — prueba que D1
//	    (claim JWT verificado > header crudo) cierra la forja de header sin depender de Kong
//	    (contrapartida del retiro de C-E, mudada a PLAT-E37). Ejercida con un router Gin mínimo
//	    + el middleware real go-shared `TenantValidation` (RejectMissingTenant: true, igual que
//	    main.go:75), sin levantar el servicio completo: (h) es verificación de middleware, no
//	    de RLS de base — la aserción de aislamiento en DB la cubren (a)-(g).
//
// NOTA sobre los repositorios CRUD de location/warehouse/stock_location: el schema real de
// las 3 tablas (verificado contra lab-postgres y las migraciones 001/010) NO tiene las columnas
// `type`/`address`/`city`/`description`/`priority` que sus entidades Go y sus `repo.Save`
// asumen e insertan. Eso es deuda preexistente del servicio (los endpoints CRUD estaban
// permanentemente rotos por el bug del camelCase `ctx.Get("tenantID")`, cerrado en T-STK4b, así
// que `Save`/`FindByID` nunca se ejercitaron en runtime y el desalineamiento quedó latente).
// No es deuda introducida ni del alcance del retrofit RLS (PLAT-E30). Para que este test
// aísle RLS y no ruido de ese desalineamiento, las 3 tablas CRUD se ejercen con SQL crudo
// bajo RLSContext (respetando el schema real), no vía sus repositorios. stock_entries y
// stock_availability SÍ usan sus repositorios: esos sí están alineados y son los que la
// lógica de negocio viva (recalcAvailability) ejerce.
//
// El contenedor se levanta UNA sola vez para todo el binario vía TestMain (no por TestXxx).
//
// Para correrlo:
//
//	GOWORK=off go test -tags=integration ./test/stock_entry/infrastructure/persistence/rls/... -count=1 -v
package rls_test

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/hornosg/go-shared/infrastructure/middleware"
	"github.com/hornosg/go-shared/infrastructure/postgres"

	stockEntryRepo "stock/src/stock_entry/infrastructure/persistence"
	stockEntryEntity "stock/src/stock_entry/domain/entity"
	stockEntryException "stock/src/stock_entry/domain/exception"
)

// containerDB es el nombre de la base del contenedor efímero. El rol de app necesita CONNECT
// sobre ESTA base (stock_app.sql hardcodea `GRANT CONNECT ON DATABASE stock_db`, el nombre
// real de lab-postgres, que no existe acá; por eso ese script NO se aplica y el rol se
// replica abajo).
const containerDB = "stock_test"

// appRoleName/appRolePassword replican en el Postgres efímero el rol stock_app creado en
// lab-postgres por PLAT-E30 T-STK3 (mismos grants sobre stock_db que stock_app.sql).
const (
	appRoleName     = "stock_app_test"
	appRolePassword = "stock_app_test"
)

// appDB es la única conexión de persistencia que usan los TestXxx — bajo el rol sin
// BYPASSRLS, la que de verdad queda sujeta a las policies tenant_isolation de la 009.
var appDB *sql.DB

// TestMain levanta el Postgres efímero, aplica las migraciones de esquema reales y crea el
// rol UNA sola vez para todo el binario — se comparte entre todas las TestXxx en vez de
// pagar el arranque del contenedor por cada una.
func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase(containerDB),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
		),
	)
	if err != nil {
		log.Fatalf("error starting postgres container: %v", err)
	}

	superConnStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("error getting connection string: %v", err)
	}

	superDB, err := sql.Open("postgres", superConnStr)
	if err != nil {
		log.Fatalf("error opening database: %v", err)
	}
	if err := superDB.PingContext(ctx); err != nil {
		log.Fatalf("error pinging database: %v", err)
	}

	if err := applyMigrations(superDB); err != nil {
		log.Fatalf("error applying migrations: %v", err)
	}
	if err := createRoles(superDB); err != nil {
		log.Fatalf("error creating roles: %v", err)
	}

	appConnStr, err := withCredentials(superConnStr, appRoleName, appRolePassword)
	if err != nil {
		log.Fatalf("error building app connection string: %v", err)
	}
	appDB, err = sql.Open("postgres", appConnStr)
	if err != nil {
		log.Fatalf("error opening app-role database: %v", err)
	}
	if err := appDB.PingContext(ctx); err != nil {
		log.Fatalf("error pinging database as %s: %v", appRoleName, err)
	}

	code := m.Run()

	_ = appDB.Close()
	_ = superDB.Close()
	_ = container.Terminate(ctx)

	os.Exit(code)
}

// applyMigrations corre en orden TODAS las migraciones .up.sql de ESQUEMA de stock-service
// (001, 004, 005, 006, 007, 008, 009, 010 — golang-migrate versiona por número de archivo, no
// por secuencia contigua; los huecos 002/003 son benignos) — el mismo esquema que corre en
// lab-postgres/stock_db tras T-STK2/T-STK2b. El subdirectorio migrations/roles/ (stock_app.sql)
// queda afuera: os.ReadDir lo devuelve como dir (e.IsDir() → skip) y además hardcodea `stock_db`
// + grants que no aplican en el contenedor efímero. El rol de app se replica acá vía
// createRoles() con sus propios grants.
//
// Estado final verificado idéntico a lab-postgres: 008 dropea trigger_update_stock_availability
// y las funciones update_stock_availability[_v2], dejando 0 triggers vivos; el recálculo de
// availability vive en Go (recalcAvailability), no en triggers (caveat T-STK3 / C2).
func applyMigrations(db *sql.DB) error {
	dir := "../../../../../migrations"
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	var files []string
	for _, e := range dirEntries {
		name := e.Name()
		if e.IsDir() || filepath.Ext(name) != ".sql" || strings.HasSuffix(name, ".down.sql") {
			continue
		}
		files = append(files, filepath.Join(dir, name))
	}
	sort.Strings(files)
	if len(files) == 0 {
		return errors.New("no se encontraron migraciones .up.sql en " + dir)
	}

	for _, f := range files {
		content, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		if _, err := db.Exec(string(content)); err != nil {
			return errors.New("migración " + f + ": " + err.Error())
		}
	}

	// schema_migrations NO la crea ninguna migración .up.sql: la crea golang-migrate v4 en el
	// primer arranque del runtime (no participa del //go:embed de migraciones de esquema, es
	// bookkeeping del migrador). Como este test aplica las migraciones con db.Exec directo (sin
	// golang-migrate), la tabla no existe — pero stock_app.sql le da SELECT, y createRoles
	// replica ese grant, así que hay que crearla. Se replica el estado estacionario real de
	// lab-postgres (v10/dirty=false, verificado T-STK6: las 001..010 aplicadas) para que el
	// schema del contenedor efímero sea fiel al del runtime, no un stand-in sin bookkeeping.
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (version bigint NOT NULL, dirty boolean NOT NULL)`); err != nil {
		return errors.New("crear schema_migrations: " + err.Error())
	}
	if _, err := db.Exec(`TRUNCATE schema_migrations`); err != nil {
		return errors.New("truncate schema_migrations: " + err.Error())
	}
	if _, err := db.Exec(`INSERT INTO schema_migrations (version, dirty) VALUES (10, false)`); err != nil {
		return errors.New("seed schema_migrations: " + err.Error())
	}
	return nil
}

// createRoles crea (con el superuser) el rol de aplicación sin DDL ni BYPASSRLS con los mismos
// grants efectivos que stock_app.sql le da a stock_app en lab-postgres sobre stock_db —
// réplica del rol real, no un stand-in:
//   - SELECT/INSERT/UPDATE/DELETE sobre locations, warehouses, stock_locations (CRUD completo).
//   - SELECT/INSERT/UPDATE (sin DELETE) sobre stock_entries y stock_availability (los repos
//     nunca DELETEan de estas dos; recalcAvailability es upsert + update).
//   - SELECT sobre schema_migrations (el runtime sólo chequea versión al arrancar).
//
// Sin secuencias (IDs UUID), sin eventbus (stock no publica eventos — una sola conexión).
func createRoles(superDB *sql.DB) error {
	stmts := []string{
		`CREATE ROLE ` + appRoleName + ` WITH LOGIN PASSWORD '` + appRolePassword + `' NOSUPERUSER NOCREATEDB NOCREATEROLE NOBYPASSRLS`,
		`GRANT CONNECT ON DATABASE ` + containerDB + ` TO ` + appRoleName,
		`GRANT USAGE ON SCHEMA public TO ` + appRoleName,
		`REVOKE ALL ON ALL TABLES IN SCHEMA public FROM PUBLIC`,
		// CRUD completo (Create/Get/Update/Delete en los repositorios).
		`GRANT SELECT, INSERT, UPDATE, DELETE ON locations      TO ` + appRoleName,
		`GRANT SELECT, INSERT, UPDATE, DELETE ON warehouses      TO ` + appRoleName,
		`GRANT SELECT, INSERT, UPDATE, DELETE ON stock_locations TO ` + appRoleName,
		// Sin DELETE (los repos nunca borran filas de estas dos tablas).
		`GRANT SELECT, INSERT, UPDATE ON stock_entries      TO ` + appRoleName,
		`GRANT SELECT, INSERT, UPDATE ON stock_availability TO ` + appRoleName,
		// El runtime sólo chequea la versión al arrancar (SELECT alcanza).
		`GRANT SELECT ON schema_migrations TO ` + appRoleName,
	}
	for _, stmt := range stmts {
		if _, err := superDB.Exec(stmt); err != nil {
			return errors.New("stmt " + stmt + ": " + err.Error())
		}
	}
	return nil
}

// withCredentials reconstruye la connection string del contenedor reemplazando usuario y
// contraseña — evita depender de las APIs internas de testcontainers para host/puerto.
func withCredentials(connStr, user, password string) (string, error) {
	u, err := url.Parse(connStr)
	if err != nil {
		return "", err
	}
	u.User = url.UserPassword(user, password)
	return u.String(), nil
}

// countRaw cuenta filas bajo un RLSContext dado — usado para las aserciones de invisibilidad
// cross-tenant (nunca vía el WHERE tenant_id=... del repository, para no confundir aislamiento
// por RLS con un filtro de la query).
func countRaw(t *testing.T, rc postgres.RLSContext, query string, args ...interface{}) int {
	t.Helper()
	var count int
	err := postgres.WithRLSInTransaction(context.Background(), appDB, rc, func(ctx context.Context, tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, query, args...).Scan(&count)
	})
	require.NoError(t, err)
	return count
}

// execRaw corre un stmt bajo un RLSContext y devuelve RowsAffected — para (a) siembra, (c)/(d).
func execRaw(t *testing.T, rc postgres.RLSContext, query string, args ...interface{}) (int64, error) {
	t.Helper()
	var affected int64
	err := postgres.WithRLSInTransaction(context.Background(), appDB, rc, func(ctx context.Context, tx *sql.Tx) error {
		res, e := tx.ExecContext(ctx, query, args...)
		if e != nil {
			return e
		}
		affected, e = res.RowsAffected()
		return e
	})
	return affected, err
}

// seedLocation inserta una location bajo RLSContext{tenant} respetando el schema REAL
// (id, name, code, tenant_id, active, created_at, updated_at — sin type/address, que no
// existen en la tabla). Devuelve el id generado.
func seedLocation(t *testing.T, tenant uuid.UUID, name, code string) string {
	t.Helper()
	id := uuid.New().String()
	_, err := execRaw(t, postgres.RLSContext{TenantID: tenant.String()},
		`INSERT INTO locations (id, name, code, tenant_id, active, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, true, NOW(), NOW())`,
		id, name, code, tenant)
	require.NoError(t, err, "la siembra de location del propio tenant debe pasar el WITH CHECK")
	return id
}

// seedWarehouse inserta un warehouse bajo RLSContext{tenant} referenciando locationID (FK real
// del mismo tenant). Schema real: id, name, code, tenant_id, location_id, active, created_at,
// updated_at.
func seedWarehouse(t *testing.T, tenant uuid.UUID, locationID, name, code string) string {
	t.Helper()
	id := uuid.New().String()
	_, err := execRaw(t, postgres.RLSContext{TenantID: tenant.String()},
		`INSERT INTO warehouses (id, name, code, tenant_id, location_id, active, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, true, NOW(), NOW())`,
		id, name, code, tenant, locationID)
	require.NoError(t, err, "la siembra de warehouse del propio tenant debe pasar el WITH CHECK")
	return id
}

// seedStockLocation inserta un stock_location bajo RLSContext{tenant} referenciando
// warehouseID (FK real del mismo tenant). Schema real: id, name, code, tenant_id, warehouse_id,
// parent_id, path, active, created_at, updated_at.
func seedStockLocation(t *testing.T, tenant uuid.UUID, warehouseID, name, code string) string {
	t.Helper()
	id := uuid.New().String()
	_, err := execRaw(t, postgres.RLSContext{TenantID: tenant.String()},
		`INSERT INTO stock_locations (id, name, code, tenant_id, warehouse_id, path, active, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, true, NOW(), NOW())`,
		id, name, code, tenant, warehouseID, id)
	require.NoError(t, err, "la siembra de stock_location del propio tenant debe pasar el WITH CHECK")
	return id
}

// ----------------------------------------------------------------------------
// (g) locations, warehouses, stock_locations — datos reales persistentes + (a)-(d)
// Las 3 tablas se ejercen con SQL crudo bajo RLSContext (ver NOTA del header del archivo):
// sus repositorios Go están desalineados con el schema real (deuda preexistente, no del
// retrofit RLS), así que usarlos mezclaría ruido de columna inexistente con la aserción de
// aislamiento. SQL crudo aísla exactamente la policy tenant_isolation.
// ----------------------------------------------------------------------------

func TestLocation_CrossTenantIsolation(t *testing.T) {
	tenantA := uuid.New()
	tenantB := uuid.New()

	locA := seedLocation(t, tenantA, "Depósito A", "LOC-A")

	t.Run("(a) A lee su location por PK", func(t *testing.T) {
		require.Equal(t, 1, countRaw(t, postgres.RLSContext{TenantID: tenantA.String()},
			`SELECT count(*) FROM locations WHERE id = $1`, locA))
	})

	t.Run("(b) B NO ve la location de A por consulta cruda", func(t *testing.T) {
		require.Equal(t, 0, countRaw(t, postgres.RLSContext{TenantID: tenantB.String()},
			`SELECT count(*) FROM locations WHERE id = $1`, locA), "la fila de A es visible desde B")
	})

	t.Run("(c) INSERT con tenant_id=A bajo sesión B viola WITH CHECK", func(t *testing.T) {
		_, err := execRaw(t, postgres.RLSContext{TenantID: tenantB.String()},
			`INSERT INTO locations (id, name, code, tenant_id, active, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, true, NOW(), NOW())`,
			uuid.New(), "intruso", "LOC-INTRUSO-1", tenantA)
		require.Error(t, err, "esperaba que el INSERT con tenant_id=A bajo sesión B fallara por WITH CHECK")
	})

	t.Run("(d) UPDATE cruda cross-tenant de B sobre la fila de A es no-op", func(t *testing.T) {
		affected, err := execRaw(t, postgres.RLSContext{TenantID: tenantB.String()},
			`UPDATE locations SET name = 'hackeado' WHERE id = $1`, locA)
		require.NoError(t, err)
		require.Equal(t, int64(0), affected, "la policy USING oculta la fila de A a B: el UPDATE no puede tocarla")

		require.Equal(t, 1, countRaw(t, postgres.RLSContext{TenantID: tenantA.String()},
			`SELECT count(*) FROM locations WHERE id = $1 AND name = 'Depósito A'`, locA),
			"el name de A no debe haber cambiado por el UPDATE de B")
	})
}

func TestWarehouse_CrossTenantIsolation(t *testing.T) {
	tenantA := uuid.New()
	tenantB := uuid.New()

	locA := seedLocation(t, tenantA, "Sede A", "LOC-W")
	whA := seedWarehouse(t, tenantA, locA, "Galpón A", "WH-A")

	t.Run("(a) A lee su warehouse por PK", func(t *testing.T) {
		require.Equal(t, 1, countRaw(t, postgres.RLSContext{TenantID: tenantA.String()},
			`SELECT count(*) FROM warehouses WHERE id = $1`, whA))
	})

	t.Run("(b) B NO ve el warehouse de A por consulta cruda", func(t *testing.T) {
		require.Equal(t, 0, countRaw(t, postgres.RLSContext{TenantID: tenantB.String()},
			`SELECT count(*) FROM warehouses WHERE id = $1`, whA), "la fila de A es visible desde B")
	})

	t.Run("(c) INSERT con tenant_id=A bajo sesión B viola WITH CHECK (FK pasa, RLS no)", func(t *testing.T) {
		// location_id referencia una location de A real: la FK se valida a nivel storage (RLS no
		// aplica a FK checks), así que la aserción aísla RLS y no ruido de FK violation.
		_, err := execRaw(t, postgres.RLSContext{TenantID: tenantB.String()},
			`INSERT INTO warehouses (id, name, code, tenant_id, location_id, active, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, true, NOW(), NOW())`,
			uuid.New(), "intruso", "WH-INTRUSO-1", tenantA, locA)
		require.Error(t, err, "esperaba que el INSERT con tenant_id=A bajo sesión B fallara por WITH CHECK (no por FK)")
	})

	t.Run("(d) UPDATE cruda cross-tenant de B sobre la fila de A es no-op", func(t *testing.T) {
		affected, err := execRaw(t, postgres.RLSContext{TenantID: tenantB.String()},
			`UPDATE warehouses SET name = 'hackeado' WHERE id = $1`, whA)
		require.NoError(t, err)
		require.Equal(t, int64(0), affected, "la policy USING oculta la fila de A a B: el UPDATE no puede tocarla")

		require.Equal(t, 1, countRaw(t, postgres.RLSContext{TenantID: tenantA.String()},
			`SELECT count(*) FROM warehouses WHERE id = $1 AND name = 'Galpón A'`, whA),
			"el name de A no debe haber cambiado por el UPDATE de B")
	})
}

func TestStockLocation_CrossTenantIsolation(t *testing.T) {
	tenantA := uuid.New()
	tenantB := uuid.New()

	locA := seedLocation(t, tenantA, "Nodo A", "LOC-SL")
	whA := seedWarehouse(t, tenantA, locA, "Galpón A2", "WH-A2")
	slA := seedStockLocation(t, tenantA, whA, "Estante A", "SL-A")

	t.Run("(a) A lee su stock_location por PK", func(t *testing.T) {
		require.Equal(t, 1, countRaw(t, postgres.RLSContext{TenantID: tenantA.String()},
			`SELECT count(*) FROM stock_locations WHERE id = $1`, slA))
	})

	t.Run("(b) B NO ve el stock_location de A por consulta cruda", func(t *testing.T) {
		require.Equal(t, 0, countRaw(t, postgres.RLSContext{TenantID: tenantB.String()},
			`SELECT count(*) FROM stock_locations WHERE id = $1`, slA), "la fila de A es visible desde B")
	})

	t.Run("(c) INSERT con tenant_id=A bajo sesión B viola WITH CHECK (FK pasa, RLS no)", func(t *testing.T) {
		_, err := execRaw(t, postgres.RLSContext{TenantID: tenantB.String()},
			`INSERT INTO stock_locations (id, name, code, tenant_id, warehouse_id, path, active, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, true, NOW(), NOW())`,
			uuid.New(), "intruso", "SL-INTRUSO-1", tenantA, whA, "path-intruso")
		require.Error(t, err, "esperaba que el INSERT con tenant_id=A bajo sesión B fallara por WITH CHECK (no por FK)")
	})

	t.Run("(d) UPDATE cruda cross-tenant de B sobre la fila de A es no-op", func(t *testing.T) {
		affected, err := execRaw(t, postgres.RLSContext{TenantID: tenantB.String()},
			`UPDATE stock_locations SET name = 'hackeado' WHERE id = $1`, slA)
		require.NoError(t, err)
		require.Equal(t, int64(0), affected, "la policy USING oculta la fila de A a B: el UPDATE no puede tocarla")

		require.Equal(t, 1, countRaw(t, postgres.RLSContext{TenantID: tenantA.String()},
			`SELECT count(*) FROM stock_locations WHERE id = $1 AND name = 'Estante A'`, slA),
			"el name de A no debe haber cambiado por el UPDATE de B")
	})
}

// ----------------------------------------------------------------------------
// stock_entries + (f) recalcAvailability reinterpretado
// stock_entries y stock_availability SÍ usan sus repositorios (alineados con el schema).
// ----------------------------------------------------------------------------

// TestStockEntry_CrossTenantIsolation cubre (a)-(d) sobre stock_entries.
func TestStockEntry_CrossTenantIsolation(t *testing.T) {
	repo := stockEntryRepo.NewPostgresStockEntryRepository(appDB)
	ctx := context.Background()

	tenantA := uuid.New()
	tenantB := uuid.New()

	entryA, err := stockEntryEntity.NewStockEntry(tenantA, "SKU-A-1", stockEntryEntity.EntryTypeInitialStock, 100)
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, entryA), "el Save del propio tenant debe pasar el WITH CHECK")

	t.Run("(a) A lee su stock_entry por FindByTenantAndSKU", func(t *testing.T) {
		list, err := repo.FindByTenantAndSKU(ctx, tenantA, "SKU-A-1")
		require.NoError(t, err)
		require.Len(t, list, 1)
		require.Equal(t, entryA.ID, list[0].ID)
		require.Equal(t, 100.0, list[0].Quantity)
	})

	t.Run("(b) B NO ve el stock_entry de A ni por el repo ni por consulta cruda", func(t *testing.T) {
		list, err := repo.FindByTenantAndSKU(ctx, tenantB, "SKU-A-1")
		require.NoError(t, err)
		require.Empty(t, list, "el stock_entry de A es visible desde B")

		require.Equal(t, 0, countRaw(t, postgres.RLSContext{TenantID: tenantB.String()},
			`SELECT count(*) FROM stock_entries WHERE id = $1`, entryA.ID), "la fila de A es visible por consulta cruda desde B")
	})

	t.Run("(c) INSERT con tenant_id=A bajo sesión B viola WITH CHECK", func(t *testing.T) {
		_, err := execRaw(t, postgres.RLSContext{TenantID: tenantB.String()},
			`INSERT INTO stock_entries (id, tenant_id, variant_sku, product_sku, entry_type, quantity, status, is_active, created_at, updated_at)
			 VALUES ($1, $2, $3, $3, 'initial_stock', 10, 'confirmed', true, NOW(), NOW())`,
			uuid.New(), tenantA, "SKU-INTRUSO-1")
		require.Error(t, err, "esperaba que el INSERT con tenant_id=A bajo sesión B fallara por WITH CHECK")
	})

	t.Run("(d) UPDATE cruda cross-tenant de B sobre la fila de A es no-op", func(t *testing.T) {
		affected, err := execRaw(t, postgres.RLSContext{TenantID: tenantB.String()},
			`UPDATE stock_entries SET quantity = 999 WHERE id = $1`, entryA.ID)
		require.NoError(t, err)
		require.Equal(t, int64(0), affected, "la policy USING oculta la fila de A a B: el UPDATE no puede tocarla")

		list, err := repo.FindByTenantAndSKU(ctx, tenantA, "SKU-A-1")
		require.NoError(t, err)
		require.Len(t, list, 1)
		require.Equal(t, 100.0, list[0].Quantity, "la quantity de A no debe haber cambiado por el UPDATE de B")
	})
}

// TestRecalcAvailability_UnderRLS cubre (f) reinterpretada (caveat T-STK3 / C2): la migración
// 008 dropeó trigger_update_stock_availability y movió el recálculo de availability al código
// Go (recalcAvailability). Este test prueba esa lógica de negocio viva bajo stock_app
// NOBYPASSRLS: Save de un StockEntry corre recalcAvailability dentro de la misma tx RLS del
// tenant del caller, y produce una fila de stock_availability aislada.
func TestRecalcAvailability_UnderRLS(t *testing.T) {
	entryRepo := stockEntryRepo.NewPostgresStockEntryRepository(appDB)
	availRepo := stockEntryRepo.NewPostgresStockAvailabilityRepository(appDB)
	ctx := context.Background()

	tenantA := uuid.New()
	tenantB := uuid.New()

	// (f.1) Save de un initial_stock dispara recalcAvailability → availability A creada y visible.
	entry1, err := stockEntryEntity.NewStockEntry(tenantA, "SKU-RECALC", stockEntryEntity.EntryTypeInitialStock, 100)
	require.NoError(t, err)
	require.NoError(t, entryRepo.Save(ctx, entry1), "Save (INSERT + recalcAvailability) debe pasar bajo RLS del caller")

	availA, err := availRepo.FindByTenantAndSKU(ctx, tenantA, "SKU-RECALC")
	require.NoError(t, err)
	require.Equal(t, 100.0, availA.TotalQuantity, "recalcAvailability debe consolidar 100 desde el initial_stock")
	require.Equal(t, 100.0, availA.AvailableQuantity)

	// (f.2) B NO ve la availability de A — recalcAvailability escribió bajo el RLSContext de A.
	gotB, err := availRepo.FindByTenantAndSKU(ctx, tenantB, "SKU-RECALC")
	require.ErrorIs(t, err, stockEntryException.ErrStockAvailabilityNotFound, "B no debe encontrar la availability de A")
	require.Nil(t, gotB)

	require.Equal(t, 0, countRaw(t, postgres.RLSContext{TenantID: tenantB.String()},
		`SELECT count(*) FROM stock_availability WHERE variant_sku = $1`, "SKU-RECALC"),
		"la fila de availability de A es visible por consulta cruda desde B")

	// (f.3) Un segundo movimiento (sale) recalcula availability dentro de la tx RLS — la lógica
	// de negocio (sumar initial, restar sale) corre correctamente bajo stock_app NOBYPASSRLS.
	entry2, err := stockEntryEntity.NewStockEntry(tenantA, "SKU-RECALC", stockEntryEntity.EntryTypeSale, 30)
	require.NoError(t, err)
	require.NoError(t, entryRepo.Save(ctx, entry2))

	availA2, err := availRepo.FindByTenantAndSKU(ctx, tenantA, "SKU-RECALC")
	require.NoError(t, err)
	require.Equal(t, 70.0, availA2.TotalQuantity, "recalcAvailability debe restar la sale: 100 - 30 = 70")
	require.Equal(t, 70.0, availA2.AvailableQuantity)

	// (f.4) El upsert de recalcAvailability respeta el índice parcial único
	// (tenant_id, variant_sku) WHERE location_id IS NULL — una sola fila por tenant+sku, no
	// duplicados por cada Save.
	require.Equal(t, 1, countRaw(t, postgres.RLSContext{TenantID: tenantA.String()},
		`SELECT count(*) FROM stock_availability WHERE variant_sku = $1 AND location_id IS NULL`, "SKU-RECALC"),
		"recalcAvailability debe upsertar (1 fila), no acumular una por Save")
}

// ----------------------------------------------------------------------------
// stock_availability (lectura) — (a)-(d)
// ----------------------------------------------------------------------------

// TestStockAvailability_CrossTenantIsolation cubre (a)-(d) sobre stock_availability. La
// escritura la hace recalcAvailability (probada arriba); acá se ejerce el aislamiento de
// lectura y los intentos cross-tenant crudos.
func TestStockAvailability_CrossTenantIsolation(t *testing.T) {
	entryRepo := stockEntryRepo.NewPostgresStockEntryRepository(appDB)
	repo := stockEntryRepo.NewPostgresStockAvailabilityRepository(appDB)
	ctx := context.Background()

	tenantA := uuid.New()
	tenantB := uuid.New()

	// Sembrar availability de A vía recalcAvailability (la forma legítima de escribirla).
	entry, err := stockEntryEntity.NewStockEntry(tenantA, "SKU-AVAIL", stockEntryEntity.EntryTypeInitialStock, 50)
	require.NoError(t, err)
	require.NoError(t, entryRepo.Save(ctx, entry))

	t.Run("(a) A lee su availability por FindByTenantAndSKU", func(t *testing.T) {
		got, err := repo.FindByTenantAndSKU(ctx, tenantA, "SKU-AVAIL")
		require.NoError(t, err)
		require.Equal(t, 50.0, got.TotalQuantity)
	})

	t.Run("(b) B NO ve la availability de A ni por el repo ni por consulta cruda", func(t *testing.T) {
		got, err := repo.FindByTenantAndSKU(ctx, tenantB, "SKU-AVAIL")
		require.ErrorIs(t, err, stockEntryException.ErrStockAvailabilityNotFound, "B no debe encontrar la availability de A")
		require.Nil(t, got)

		require.Equal(t, 0, countRaw(t, postgres.RLSContext{TenantID: tenantB.String()},
			`SELECT count(*) FROM stock_availability WHERE variant_sku = $1`, "SKU-AVAIL"),
			"la fila de A es visible por consulta cruda desde B")
	})

	t.Run("(c) INSERT con tenant_id=A bajo sesión B viola WITH CHECK", func(t *testing.T) {
		_, err := execRaw(t, postgres.RLSContext{TenantID: tenantB.String()},
			`INSERT INTO stock_availability (id, tenant_id, variant_sku, product_sku, available_quantity, reserved_quantity, total_quantity, updated_at)
			 VALUES ($1, $2, $3, $3, 1, 0, 1, NOW())`,
			uuid.New(), tenantA, "SKU-AVAIL-INTRUSO")
		require.Error(t, err, "esperaba que el INSERT con tenant_id=A bajo sesión B fallara por WITH CHECK")
	})

	t.Run("(d) UPDATE cruda cross-tenant de B sobre la fila de A es no-op", func(t *testing.T) {
		affected, err := execRaw(t, postgres.RLSContext{TenantID: tenantB.String()},
			`UPDATE stock_availability SET total_quantity = 999 WHERE variant_sku = $1`, "SKU-AVAIL")
		require.NoError(t, err)
		require.Equal(t, int64(0), affected, "la policy USING oculta la fila de A a B: el UPDATE no puede tocarla")

		got, err := repo.FindByTenantAndSKU(ctx, tenantA, "SKU-AVAIL")
		require.NoError(t, err)
		require.Equal(t, 50.0, got.TotalQuantity, "el total_quantity de A no debe haber cambiado por el UPDATE de B")
	})
}

// ----------------------------------------------------------------------------
// (e) fail-closed sin contexto
// ----------------------------------------------------------------------------

// TestFailClosed_WithoutTenantContext cubre (e) / M1: lecturas Y escrituras directas al pool
// (fuera de WithRLSInTransaction → sin SET LOCAL app.tenant_id) erran o devuelven 0 filas en
// LAS 5 TABLAS — nunca todas. Corre reutilizando el pool appDB cuyas conexiones físicas YA
// tuvieron SET LOCAL fijado (y reseteado al commitear) — el escenario real de un pool
// reutilizado entre requests de distintos tenants.
func TestFailClosed_WithoutTenantContext(t *testing.T) {
	ctx := context.Background()

	// Garantizar al menos una fila en cada tabla para que "0 visibles" sea significativo.
	tenant := uuid.New()
	locID := seedLocation(t, tenant, "FC Loc", "FC-LOC")
	whID := seedWarehouse(t, tenant, locID, "FC WH", "FC-WH")
	seedStockLocation(t, tenant, whID, "FC SL", "FC-SL")
	entryRepo := stockEntryRepo.NewPostgresStockEntryRepository(appDB)
	entry, err := stockEntryEntity.NewStockEntry(tenant, "SKU-FC", stockEntryEntity.EntryTypeInitialStock, 5)
	require.NoError(t, err)
	require.NoError(t, entryRepo.Save(ctx, entry)) // también siembra stock_availability vía recalcAvailability

	tables := []string{"locations", "warehouses", "stock_locations", "stock_entries", "stock_availability"}

	t.Run("(e) lecturas sin contexto: 0 filas o error, nunca todas", func(t *testing.T) {
		for _, table := range tables {
			var count int
			err := appDB.QueryRowContext(ctx, `SELECT count(*) FROM `+table).Scan(&count)
			if err != nil {
				continue // fail-closed vía error (current_setting sin valor): comportamiento esperado
			}
			require.Equal(t, 0, count, table+" devolvió filas sin contexto de tenant — RLS no está aislando")
		}
	})

	t.Run("(e) escrituras sin contexto fallan en las 5 tablas", func(t *testing.T) {
		// locations
		_, errInsLoc := appDB.ExecContext(ctx,
			`INSERT INTO locations (id, name, code, tenant_id, active, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, true, NOW(), NOW())`,
			uuid.New(), "x", "FC-INS-1", tenant)
		require.Error(t, errInsLoc, "INSERT en locations sin app.tenant_id debe fallar")

		_, errUpdLoc := appDB.ExecContext(ctx,
			`UPDATE locations SET active = false WHERE tenant_id = $1`, tenant)
		require.Error(t, errUpdLoc, "UPDATE en locations sin app.tenant_id debe fallar")

		// warehouses
		_, errInsWh := appDB.ExecContext(ctx,
			`INSERT INTO warehouses (id, name, code, tenant_id, location_id, active, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, true, NOW(), NOW())`,
			uuid.New(), "x", "FC-INS-2", tenant, locID)
		require.Error(t, errInsWh, "INSERT en warehouses sin app.tenant_id debe fallar")

		_, errUpdWh := appDB.ExecContext(ctx,
			`UPDATE warehouses SET active = false WHERE tenant_id = $1`, tenant)
		require.Error(t, errUpdWh, "UPDATE en warehouses sin app.tenant_id debe fallar")

		// stock_locations
		_, errInsSL := appDB.ExecContext(ctx,
			`INSERT INTO stock_locations (id, name, code, tenant_id, warehouse_id, path, active, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, true, NOW(), NOW())`,
			uuid.New(), "x", "FC-INS-3", tenant, whID, "p")
		require.Error(t, errInsSL, "INSERT en stock_locations sin app.tenant_id debe fallar")

		_, errUpdSL := appDB.ExecContext(ctx,
			`UPDATE stock_locations SET active = false WHERE tenant_id = $1`, tenant)
		require.Error(t, errUpdSL, "UPDATE en stock_locations sin app.tenant_id debe fallar")

		// stock_entries
		_, errInsEntry := appDB.ExecContext(ctx,
			`INSERT INTO stock_entries (id, tenant_id, variant_sku, product_sku, entry_type, quantity, status, is_active, created_at, updated_at)
			 VALUES ($1, $2, $3, $3, 'initial_stock', 1, 'confirmed', true, NOW(), NOW())`,
			uuid.New(), tenant, "SKU-FC-INS")
		require.Error(t, errInsEntry, "INSERT en stock_entries sin app.tenant_id debe fallar")

		_, errUpdEntry := appDB.ExecContext(ctx,
			`UPDATE stock_entries SET quantity = 1 WHERE tenant_id = $1`, tenant)
		require.Error(t, errUpdEntry, "UPDATE en stock_entries sin app.tenant_id debe fallar")

		// stock_availability
		_, errInsAvail := appDB.ExecContext(ctx,
			`INSERT INTO stock_availability (id, tenant_id, variant_sku, product_sku, available_quantity, reserved_quantity, total_quantity, updated_at)
			 VALUES ($1, $2, $3, $3, 1, 0, 1, NOW())`,
			uuid.New(), tenant, "SKU-FC-AVAIL-INS")
		require.Error(t, errInsAvail, "INSERT en stock_availability sin app.tenant_id debe fallar")

		_, errUpdAvail := appDB.ExecContext(ctx,
			`UPDATE stock_availability SET total_quantity = 1 WHERE tenant_id = $1`, tenant)
		require.Error(t, errUpdAvail, "UPDATE en stock_availability sin app.tenant_id debe fallar")
	})
}

// ----------------------------------------------------------------------------
// (h) forja de header — JWT A + X-Tenant-ID: B → 403 (D1, sin depender de Kong)
// ----------------------------------------------------------------------------

// jwtSecret es el secret de firma del contenedor efímero de middleware. No es el del lab: (h)
// prueba la semántica de TenantValidation (claim > header) aisladamente, no la integración con
// el JWT_SECRET real (que ya verificó T-STK6 en vivo contra el servicio).
const jwtSecret = "rls-test-secret-T-STK7"

// mintTenantJWT firma un JWT HS256 con un claim tenant_id (y namespace opcional).
func mintTenantJWT(t *testing.T, tenantID, namespace string) string {
	t.Helper()
	claims := jwt.MapClaims{
		"tenant_id": tenantID,
		"user_id":   "u-test",
		"exp":       time.Now().Add(time.Hour).Unix(),
	}
	if namespace != "" {
		claims["namespace"] = namespace
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(jwtSecret))
	require.NoError(t, err)
	return signed
}

// newTenantProbeRouter monta un router Gin mínimo con el middleware real go-shared
// TenantValidation (RejectMissingTenant: true, igual que main.go:75) y un handler stub que
// devuelve el tenant_id que el middleware fijó desde el claim. (h) es verificación de
// middleware, no de RLS de base — por eso no necesita la DB.
func newTenantProbeRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.TenantValidation(middleware.TenantValidationConfig{
		JWTSecret:           jwtSecret,
		RejectMissingTenant: true,
	}))
	r.GET("/probe", func(c *gin.Context) {
		// tenant_id SIEMPRE del claim JWT verificado (D1), nunca del header.
		c.JSON(http.StatusOK, gin.H{"tenant_id": c.GetString("tenant_id")})
	})
	return r
}

// TestHeaderForge_RejectedByMiddleware cubre (h): un JWT válido de tenant A con
// X-Tenant-ID: B → 403 "Tenant mismatch" y cero filas de B. Prueba que D1 (claim verificado >
// header crudo) cierra la forja de header sin depender de Kong (contrapartida del retiro de
// C-E, mudada a PLAT-E37). El handler nunca ve el header forjado: el middleware aborta antes.
func TestHeaderForge_RejectedByMiddleware(t *testing.T) {
	router := newTenantProbeRouter()

	tenantA := uuid.New().String()
	tenantB := uuid.New().String()

	t.Run("(h.1) JWT A + X-Tenant-ID: B → 403 (forja de header rechazada)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/probe", nil)
		req.Header.Set("Authorization", "Bearer "+mintTenantJWT(t, tenantA, ""))
		req.Header.Set("X-Tenant-ID", tenantB)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusForbidden, w.Code, "el middleware debe rechazar la forja de header con 403")
		require.Contains(t, w.Body.String(), "Tenant mismatch")
	})

	t.Run("(h.2) JWT A + X-Tenant-ID: A → 200, handler ve tenant_id del claim (no del header)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/probe", nil)
		req.Header.Set("Authorization", "Bearer "+mintTenantJWT(t, tenantA, ""))
		req.Header.Set("X-Tenant-ID", tenantA)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Body.String(), tenantA, "el handler debe ver el tenant_id del claim JWT")
	})

	t.Run("(h.3) sin Authorization → 401 (fail-closed)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/probe", nil)
		req.Header.Set("X-Tenant-ID", tenantA)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("(h.4) JWT sin tenant_id + RejectMissingTenant → 403 (no bypass)", func(t *testing.T) {
		// Token sin claim tenant_id: el bypass histórico está cerrado (RejectMissingTenant: true).
		claims := jwt.MapClaims{"user_id": "u-notenant", "exp": time.Now().Add(time.Hour).Unix()}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := token.SignedString([]byte(jwtSecret))
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/probe", nil)
		req.Header.Set("Authorization", "Bearer "+signed)
		req.Header.Set("X-Tenant-ID", tenantA)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusForbidden, w.Code, "un token sin tenant_id no debe pasar (RejectMissingTenant)")
	})

	t.Run("(h.5) sin X-Tenant-ID → 400 (el header sigue requerido por el middleware)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/probe", nil)
		req.Header.Set("Authorization", "Bearer "+mintTenantJWT(t, tenantA, ""))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
	})
}