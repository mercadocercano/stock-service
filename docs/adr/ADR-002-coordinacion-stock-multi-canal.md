# ADR-002: Coordinación de stock multi-canal con dominio puro y stock físico único

**Estado**: Aceptado
**Fecha**: 2026-06-10
**Contexto**: El servicio debe vender el mismo producto por múltiples canales (POS, Marketplace) sin sobrevender. Había que definir cómo se coordina el stock entre canales y dónde vive esa lógica, evitando acoplar entidades o mezclar reglas con la persistencia.

## Decisión

Adoptamos un stock físico único compartido por todos los canales, gobernado por un invariante del sistema: **si un producto está habilitado para Marketplace, todos los canales deben respetar el stock físico**. La coordinación entre canales se resuelve con un Policy Object dedicado, evitando acoplamiento entre entidades. El canal se modela como enum (`POS` | `MARKETPLACE`, extensible) y la quota es manual y fija (no porcentaje, no dinámica). Toda la validación vive en el dominio en Go (dominio puro, sin tocar DB ni triggers), con el repositorio expuesto como port para testear con mocks. El módulo se construyó dominio-primero, con 49 tests unitarios antes de la capa de persistencia.

## Alternativas consideradas

| Opción | Por qué no |
|--------|-----------|
| Stock separado por canal | Genera desperdicio de inventario |
| Quota como porcentaje | Complejidad innecesaria |
| Triggers para validar quota | Rompe el principio de dominio puro |
| Validación solo en el request | Deja el dominio sin reglas |
| Repository concreto primero | Dificulta el testing con mocks |

## Consecuencias

**Positivas**: Imposible sobrevender con canales mixtos activos; el dominio es testeable de forma aislada; coordinación entre canales sin acoplar entidades.
**Negativas / trade-offs**: La quota fija manual no se ajusta dinámicamente; agregar un canal nuevo requiere extender el enum y revisar la policy.
**Neutral**: La infraestructura de persistencia del módulo quedó como Fase 2 (no implementada en este hito).
