# Milpa — Eco-Mercado Digital

Plataforma digital de mercado agropecuario que conecta directamente a pequeños y medianos agro-productores de frutas y cítricos con compradores minoristas y mayoristas en Nicaragua. El sistema busca eliminar la cadena de intermediación que encarece los productos y reduce el beneficio del productor, permitiendo un canal directo de venta que disminuya el desperdicio de cosechas y ofrezca precios más justos para ambas partes.

Proyecto desarrollado para el **Hackathon Nicaragua 2026** (Categoría Agropecuario / Medio Ambiente).

## Estado actual

Backend implementado con arquitectura hexagonal. Lo que el código soporta hoy:

- **Autenticación JWT** (registro e inicio de sesión) con bcrypt y middleware de autorización.
- **Catálogo**: categorías, empresas y publicaciones de productos/servicios (CRUD completo).
- **Reseñas**: calificación de 1 a 5 estrellas con comentario sobre la empresa.
- **Consultas**: mensajes de un comprador sobre una publicación, con estados `pending`, `read`, `replied`, `closed`.
- **Caché Redis** sobre los casos de uso, implementado con patrón decorator (TTL 5 min e invalidación).
- **Migraciones** de base de datos automáticas al arrancar, validación de request bodies y CORS.

## Roadmap

Funcionalidades definidas en la especificación ([docs/brief.typ](docs/brief.typ)) y pendientes de implementar:

| Funcionalidad | Descripción | Estado |
|---|---|---|
| Chat directo (RF-08) | Canal de chat en tiempo real (WebSocket) entre comprador minorista y agricultor, sin necesidad de match previo. | Pendiente |
| Solicitudes de abastecimiento | Publicación de necesidades por compradores mayoristas (producto, cantidad, unidad, fecha límite) y ofertas de los agricultores. | Pendiente |
| Sistema de match | Asignación tipo Tinder entre mayorista y agricultor basada en cercanía, precio, tiempo de entrega y reputación. Al aceptar se genera el match, se rechazan las demás ofertas y se habilita el chat. | Pendiente |
| Chat post-match (RF-13) | Chat en tiempo real (WebSocket) entre mayorista y agricultor, habilitado solo después de formarse el match. | Pendiente |
| Ofertas de liquidación | Publicación de lotes completos a precio preferencial para evitar pérdidas por excedente, con opción de restringir a mayoristas y asignación FCFS o selección. | Pendiente |
| Caducidad de publicaciones | Fecha de expiración por publicación según temporalidad del cultivo — diferenciador frente a marketplaces genéricos. | Pendiente |

## Arquitectura

Backend en **Go** con **arquitectura hexagonal** (puertos y adaptadores), que desacopla el dominio de las dependencias externas:

```
┌────────────────────────────────────────────────────┐
│                    HTTP API (chi)                  │
│  handlers → middleware (auth, CORS) → router       │
└────────────────────────┬───────────────────────────┘
                         │ primary ports
┌────────────────────────▼───────────────────────────┐
│                 Casos de uso (aplication)          │
│           + wrappers de caché (decorator)          │
└────────────────────────┬───────────────────────────┘
                         │ secondary ports
┌────────────────────────▼───────────────────────────┐
│            Adaptadores secundarios                 │
│  PostgreSQL (pgx) · Redis · JWT · bcrypt · clock   │
└────────────────────────────────────────────────────┘
```

- **domain**: entidades de negocio, reglas y puertos (primary/secondary).
- **aplication**: casos de uso y DTOs.
- **infrastructure**: adaptadores primarios (API) y secundarios (repositorios, caché, auth, tiempo).
- **cmd**: binarios de la aplicación (`api`) y de migraciones (`migrate`).

## Stack tecnológico

| Capa | Tecnología |
|---|---|
| Lenguaje | Go 1.26 |
| HTTP | chi/v5 + go-chi/cors |
| Base de datos | PostgreSQL 16 (pgx/v5) |
| Migraciones | golang-migrate/v4 (automáticas al arrancar) |
| Caché | Redis 7 (go-redis/v9) |
| Autenticación | JWT (golang-jwt/v5) + bcrypt (x/crypto) |
| Configuración | godotenv |
| Tests | testify |
| Contenedores | Docker Compose |

## Inicio rápido

Requisitos: **Docker** y **Docker Compose**.

```bash
# 1. Configurar variables de entorno
cp server/.env.example server/.env

# 2. Levantar el stack (api, db, redis, adminer, migrate)
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
| Adminer | `8081` |

Las migraciones de base de datos se aplican automáticamente al iniciar la API, o de forma explícita con el servicio `migrate` (binario `./migrate`).

## Variables de entorno

| Variable | Descripción |
|---|---|
| `POSTGRES_HOST` / `POSTGRES_PORT` | Host y puerto de PostgreSQL |
| `POSTGRES_USER` / `POSTGRES_PASSWORD` | Credenciales de la base de datos |
| `POSTGRES_DB` | Nombre de la base de datos |
| `DB_SSLMODE` | Modo SSL de la conexión (ej. `disable`) |
| `JWT_SECRET` | Secreto para firmar tokens JWT |
| `SERVER_PORT` | Puerto de escucha de la API |
| `REDIS_HOST` / `REDIS_PORT` / `REDIS_PASSWORD` | Conexión a Redis |

## API

Base: `http://localhost:8080/api/v1`

| Método | Ruta | Autenticación | Descripción |
|---|---|---|---|
| POST | `/auth/register` | — | Registro de usuario |
| POST | `/auth/login` | — | Inicio de sesión (JWT) |
| GET | `/categories` | — | Lista de categorías |
| GET | `/users/{id}` | — | Detalle de usuario |
| PATCH | `/users/{id}` | ✅ | Actualización de perfil |
| GET | `/companies/{id}` | — | Detalle de empresa |
| GET | `/companies/` | — | Empresas por propietario |
| POST | `/companies/` | ✅ | Crear empresa |
| PATCH | `/companies/{id}` | ✅ | Actualizar empresa |
| GET | `/offerings/{id}` | — | Detalle de oferta |
| GET | `/offerings/` | — | Ofertas por empresa |
| POST | `/offerings/` | ✅ | Crear oferta |
| PATCH | `/offerings/{id}` | ✅ | Actualizar oferta |
| GET | `/reviews/` | — | Lista de reseñas |
| POST | `/reviews/` | ✅ | Crear reseña |
| GET | `/inquiries/{id}` | — | Detalle de solicitud |
| GET | `/inquiries/` | — | Solicitudes por usuario |
| POST | `/inquiries/` | ✅ | Crear solicitud |
| PATCH | `/inquiries/{id}` | ✅ | Actualizar solicitud |

## Comandos útiles

```bash
make build          # compilar la API (bin/api)
make run            # ejecutar la API localmente
make test           # go test ./... -v -race -count=1
make lint           # go vet ./...
make up             # docker compose up -d
make down           # docker compose down
make down-clean     # docker compose down -v (borra volúmenes)
make logs-api       # logs de la API
make logs-db        # logs de PostgreSQL
```

## Estructura del repositorio

```
├── docker-compose.yml
├── docs/                     # Documentación del proyecto
│   ├── brief.typ / brief.pdf # Especificación de requisitos
│   └── modelos.pdf
└── server/
    ├── cmd/
    │   ├── api/              # Punto de entrada de la API
    │   └── migrate/          # Runner de migraciones
    ├── domain/
    │   ├── entities/         # Entidades y reglas de negocio
    │   └── port/             # Puertos primary/secondary
    ├── aplication/
    │   ├── dto/              # Objetos de transferencia
    │   └── use-cases/        # Casos de uso (+ wrappers de caché)
    ├── infrastructure/
    │   ├── adapters/
    │   │   ├── primary/api/  # Handlers, middleware, router
    │   │   └── secondary/    # Repositorios, caché, auth, clock
    │   ├── config/           # Carga de configuración
    │   └── database/         # Conexión y migraciones
    ├── internal/             # Auth context y validación
    ├── Makefile
    ├── Dockerfile
    └── go.mod
```

## Documentación

- `docs/brief.typ` / `docs/brief.pdf` — especificación de requisitos completa: 17 requisitos funcionales, 7 no funcionales, flujos comerciales y modelo de dominio.
- `docs/modelos.pdf` — modelos de diseño del dominio.

## Licencia

GPL-3.0 — ver [LICENSE](LICENSE).
