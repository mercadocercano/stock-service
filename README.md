# saas-mt-stock-service

Este proyecto fue extraído del monorepo SaaS Marketplace como parte de la migración a repositorios independientes.

## Estado en producción (Mar 2026)

| Aspecto | Estado |
|---------|--------|
| **K8s** | ✅ Desplegado en `k8s/stock/` |
| **Kong** | Ruta `/stock/` |
| **DB** | `stock_db` |
| **Endpoints clave** | `POST /api/v1/sale`, `POST /api/v1/stock-entries`, `GET /api/v1/availability` |
| **Build** | `go build -o stock-service .` (desde `main.go`, no `cmd/api`) |

Ver: [`docs/`](docs/README.md) para arquitectura, ADRs y guías.

---

## Descripción

Servicio stock del ecosistema SaaS Marketplace.

## Tecnología

- **Tipo**: go
- **Lenguaje**: Go
- **Framework**: Gin/Fiber
- **Base de datos**: PostgreSQL

## Desarrollo

### Prerrequisitos

- Go 1.21+
- PostgreSQL 15+
- Docker (opcional)

### Instalación

```bash
# Clonar el repositorio
git clone https://github.com/trinityweb/saas-mt-stock-service.git
cd saas-mt-stock-service

# Instalar dependencias
go mod download

# Ejecutar
go run main.go
```

### Docker

```bash
# Construir imagen
docker build -t saas-mt-stock-service .

# Ejecutar contenedor
docker run -p 8080:8080 saas-mt-stock-service
```

## Configuración

Copia `.env.example` a `.env` y configura las variables necesarias.

## Documentación

Ver [`docs/`](docs/README.md) para arquitectura, ADRs, guías operativas y setup.

- [Documentación de API (OpenAPI)](./api-docs/openapi.yml)
- [Índice de documentación](docs/README.md)

## Migración desde Monorepo

Este proyecto fue extraído del monorepo original manteniendo todo su historial de git.

**Repositorio original**: https://github.com/trinityweb/saas-marketplace.git

## Contribución

1. Fork el proyecto
2. Crea una rama para tu feature
3. Commit tus cambios
4. Push a la rama
5. Abre un Pull Request

## Licencia

[Licencia del proyecto]
