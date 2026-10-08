# Milpa — Eco-Mercado Digital

Plataforma digital de mercado agropecuario que conecta directamente a pequeños y medianos agro-productores de frutas y cítricos con compradores minoristas y mayoristas en Nicaragua. El sistema busca eliminar la cadena de intermediación que encarece los productos y reduce el beneficio del productor, permitiendo un canal directo de venta que disminuya el desperdicio de cosechas y ofrezca precios más justos para ambas partes.

Proyecto desarrollado para el **Hackathon Nicaragua 2026** (Categoría Agropecuario / Medio Ambiente).

## Clientes

| Cliente | Repositorio / ubicación | Stack | Descripción |
|---|---|---|---|
| Web | `frontend/` (este repositorio) | React 19 + Vite + Tailwind CSS v4 | Portal de marketplace con panel de comprador y panel de productor. |
| Móvil | [CharFranR/milpaClient](https://github.com/CharFranR/milpaClient) | Flutter | Cliente móvil con flujos de comprador y agricultor. |

Ambos clientes consumen la misma API (`/api/v1`).

**Cliente web:**

```bash
cd frontend
npm install
npm run dev
```

Instrucciones completas en [frontend/README.md](frontend/README.md).

**Cliente móvil:** clonar [CharFranR/milpaClient](https://github.com/CharFranR/milpaClient) y seguir las instrucciones de ese repositorio.

## Estado actual

Backend implementado en **Go** con arquitectura hexagonal. Lo que el código soporta hoy:

- **Autenticación JWT** (registro e inicio de sesión) con bcrypt y middleware de autorización. Roles: `1` agricultor, `2`–`4` comprador (minorista, mayorista detallista, mayorista corporativo), `5` admin y `6` auditor (solo lectura).
- **Moderación de cuentas**: suspensión de usuarios y guardias de acceso por rol y estado.
- **Catálogo**: categorías y unidades de medida (CRUD administrativo) y publicaciones de productos con imágenes, caducidad configurable, renovación y desactivación.
- **Búsqueda** con filtros, orden y paginación, indexada en Elasticsearch.
- **Empresas y consultas**: registro de empresas agrícolas y mensajes de compradores sobre publicaciones.
- **Chat**: conversaciones, mensajes y canal WebSocket autenticado (`/ws/{conversationID}`).
- **Match**: like/pass entre comprador mayorista y agricultor, listado priorizado de candidatos y creación automática de la conversación al formarse el match.
- **Solicitudes de abastecimiento**: publicación por compradores mayoristas, actualización de montos y plazos, cancelación y expiración.
- **Ofertas a solicitudes**: ofertas de agricultores, listados por solicitud y retiro.
- **Liquidaciones**: lotes completos a precio preferencial con expresión de interés y asignación.
- **Transacciones**: confirmación de inicio y de entrega, cancelación y consultas por match o por solicitud.
- **Inventario y recomendaciones**: inventario por productor y disponibilidad recomendada.
- **Reseñas**: calificación de 1 a 5 estrellas con comentario.
- **Reportes y administración**: reportes con resolución por admin y panel administrativo (usuarios, roles, auditoría, categorías, unidades de medida y estadísticas).
- **Infraestructura**: caché Redis sobre los casos de uso (TTL 5 min con invalidación), migraciones automáticas al arrancar, carga de imágenes, CORS y validación de request bodies.

## Roadmap

Mejoras identificadas en la especificación ([docs/brief.typ](docs/brief.typ)) y aún no implementadas:

| Mejora | Descripción |
|---|---|
| Verificación SMS con badge | Verificación de agricultores individuales con insignia (RF-03, opcional para personas naturales). |
| Notificaciones | Avisos de nueva oferta, match, mensaje y recordatorio de entrega. |
| Favoritos | Relación de favorito entre un comprador y un agricultor. |
| Bloqueo automático | Suspensión automática tras múltiples reportes y bloqueo de publicaciones con información incompleta. |

## Arquitectura

Backend en **Go** con **arquitectura hexagonal** (puertos y adaptadores), que desacopla el dominio de las dependencias externas:

```
┌────────────────────────────────────────────────────┐
│          HTTP API (chi) + WebSocket (chat)         │
│   handlers → middleware (auth, roles, CORS) → router│
└────────────────────────┬───────────────────────────┘
                         │ primary ports
┌────────────────────────▼───────────────────────────┐
│                 Casos de uso (aplication)          │
│           + wrappers de caché (decorator)          │
└────────────────────────┬───────────────────────────┘
                         │ secondary ports
┌────────────────────────▼───────────────────────────┐
│            Adaptadores secundarios                 │
│  PostgreSQL (pgx) · Redis · Elasticsearch · JWT ·  │
│                 bcrypt · clock                     │
└────────────────────────────────────────────────────┘
```

- **domain**: entidades de negocio, reglas y puertos (primary/secondary).
- **aplication**: casos de uso y DTOs.
- **infrastructure**: adaptadores primarios (API HTTP, WebSocket) y secundarios (repositorios, caché, búsqueda, auth, tiempo).
- **cmd**: binarios de la aplicación (`api`), migraciones (`migrate`), datos de prueba (`seed`) y reindexado de búsqueda (`reindex`).

## Stack tecnológico

| Capa | Tecnología |
|---|---|
| Lenguaje | Go 1.26 |
| HTTP | chi/v5 + go-chi/cors |
| Tiempo real | WebSocket (gorilla/websocket) |
| Base de datos | PostgreSQL 16 (pgx/v5) |
| Migraciones | golang-migrate/v4 (automáticas al arrancar) |
| Búsqueda | Elasticsearch 8.12 (go-elasticsearch/v8) |
| Caché | Redis 7 (go-redis/v9) |
| Autenticación | JWT (golang-jwt/v5) + bcrypt (x/crypto) |
| Configuración | godotenv |
| Tests | testify + testcontainers-go |
| Contenedores | Docker Compose |
| Cliente web | React 19 + Vite + Tailwind CSS v4 |
| Cliente móvil | Flutter |

## Inicio rápido

Requisitos: **Docker** y **Docker Compose**.

```bash
# 1. Configurar variables de entorno
cp server/.env.example server/.env

# 2. Levantar el stack (api, db, redis, elasticsearch, adminer, migrate)
docker compose up -d

# 3. Verificar
curl http://localhost:8080/api/v1/categories
```

El stack expone:

| Servicio | Puerto |
|---|---|
| API | `8080` |
| PostgreSQL | `5432` |
| Redis | `6379` |
| Elasticsearch | `9200` |
| Adminer | `8081` |

Las migraciones de base de datos se aplican automáticamente al iniciar la API, o de forma explícita con el servicio `migrate` (binario `./migrate`). Para cargar datos de prueba: `make seed`.

## Ejecución

Con el stack levantado, cada componente del sistema se ejecuta así:

| Componente | Comando | Acceso |
|---|---|---|
| Backend (API) | `docker compose up -d` (contenedor `api`) | `http://localhost:8080/api/v1` |
| Cliente web | `cd frontend && npm run dev` (requiere Node.js ≥ 18) | `http://localhost:5173` (puerto por defecto de Vite) |
| Adminer (UI de la base de datos) | `docker compose up -d` | `http://localhost:8081` |
| Cliente móvil | `flutter pub get && flutter run` dentro de `flutter_application_1/` en [milpaClient](https://github.com/CharFranR/milpaClient) (requiere Flutter SDK) | Emulador o dispositivo |

## Variables de entorno

| Variable | Descripción |
|---|---|
| `POSTGRES_HOST` / `POSTGRES_PORT` | Host y puerto de PostgreSQL |
| `POSTGRES_USER` / `POSTGRES_PASSWORD` | Credenciales de la base de datos |
| `POSTGRES_DB` | Nombre de la base de datos |
| `DB_SSLMODE` | Modo SSL de la conexión (ej. `disable`, `require`) |
| `JWT_SECRET` | Secreto para firmar tokens JWT |
| `SERVER_PORT` | Puerto de escucha de la API |
| `REDIS_HOST` / `REDIS_PORT` / `REDIS_PASSWORD` | Conexión a Redis |
| `ESCLIENT_USER` / `ESCLIENT_PASSWORD` | Credenciales de Elasticsearch |
| `ESCLIENT_ENDPOINT1` / `ESCLIENT_ENDPOINT2` | Endpoints de Elasticsearch |
| `ESCLIENT_MAXID` | Máximo de conexiones idle por host hacia Elasticsearch |
| `ESCLIENT_INDEX` | Índice de búsqueda (por defecto `milpa-offerings`) |

## API

Referencia completa: **[API.md](API.md)** (80 registros de ruta / 79 rutas distintas). Base: `http://localhost:8080/api/v1`.

| Área | Rutas | Descripción |
|---|---|---|
| Auth | `/auth/register`, `/auth/login` | Registro e inicio de sesión (JWT) |
| Usuarios | `/users/{id}`, `/users/{id}/photo` | Perfil público/privado y foto |
| Catálogo | `/categories`, `/units-of-measure`, `/offerings`, `/images` | Categorías, unidades, publicaciones e imágenes |
| Búsqueda | `/search` | Búsqueda con filtros, orden y paginación |
| Empresas y consultas | `/companies`, `/inquiries` | Empresas agrícolas y consultas de compradores |
| Reseñas | `/reviews`, `/reviews/average` | Calificaciones de 1 a 5 estrellas |
| Chat | `/conversations`, `/messages`, `/ws/{conversationID}` | Conversaciones, mensajes y WebSocket |
| Match | `/matches` | Like/pass y candidatos priorizados |
| Abastecimiento | `/supply-requests`, `/supply-offers` | Solicitudes mayoristas y ofertas de agricultores |
| Liquidaciones | `/liquidations` | Lotes a precio preferencial, interés y asignación |
| Transacciones | `/transactions` | Confirmación de inicio/entrega y cancelación |
| Inventario | `/inventory`, `/recommendations` | Inventario por productor y disponibilidad |
| Reportes | `/reports` | Reportes y moderación |
| Admin | `/admin` | Usuarios, roles, auditoría, categorías, unidades y estadísticas |

## Comandos útiles

```bash
make build          # compilar la API (bin/api)
make run            # ejecutar la API localmente
make test           # go test ./... -v -race -count=1
make lint           # go vet ./...
make up             # docker compose up -d
make down           # docker compose down
make down-clean     # docker compose down -v (borra volúmenes)
make restart-api    # reiniciar el contenedor de la API
make seed           # cargar datos de prueba (perfil tools)
make logs-api       # logs de la API
make logs-db        # logs de PostgreSQL
```

## Estructura del repositorio

```
├── docker-compose.yml
├── render.yaml               # Despliegue de la API (Render)
├── API.md                    # Referencia completa de la API
├── frontend/                 # Cliente web (React + Vite)
├── docs/                     # Documentación del proyecto
│   ├── brief.typ / brief.pdf
│   ├── arquitectura.typ / arquitectura.pdf
│   ├── modelos.pdf
│   └── security-best-practices.typ / .pdf
└── server/
    ├── cmd/
    │   ├── api/              # Punto de entrada de la API
    │   ├── migrate/          # Runner de migraciones
    │   ├── seed/             # Datos de prueba
    │   └── reindex/          # Reindexado de búsqueda
    ├── domain/
    │   ├── entities/         # Entidades y reglas de negocio
    │   └── port/             # Puertos primary/secondary
    ├── aplication/
    │   ├── dto/              # Objetos de transferencia
    │   └── use-cases/        # Casos de uso (+ wrappers de caché)
    ├── infrastructure/
    │   ├── adapters/
    │   │   ├── primary/api/  # Handlers, middleware, router
    │   │   └── secondary/    # Repositorios, caché, búsqueda, auth, clock
    │   ├── config/           # Carga de configuración
    │   └── database/         # Conexión y migraciones
    ├── internal/             # Auth context y validación
    ├── Makefile
    ├── Dockerfile
    └── go.mod
```

## Documentación

- [API.md](API.md) — referencia completa de la API (auth, roles, flujos y ejemplo end-to-end).
- `docs/brief.typ` / `docs/brief.pdf` — especificación de requisitos completa: 17 requisitos funcionales, 7 no funcionales, flujos comerciales y modelo de dominio.
- `docs/arquitectura.typ` / `docs/arquitectura.pdf` — arquitectura del sistema.
- `docs/modelos.pdf` — modelos de diseño del dominio.
- `docs/security-best-practices.typ` / `.pdf` — mejores prácticas de seguridad.

## Licencia

GPL-3.0 — ver [LICENSE](LICENSE).
