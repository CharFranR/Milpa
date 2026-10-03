# Milpa — Frontend

Conecta productores agropecuarios con compradores en Nicaragua mediante consultas y contacto directo. No es una tienda en línea: el comprador envía una consulta al productor y coordinan la compra por WhatsApp. Construido para el **Hackathon Nicaragua 2026**.

---

## Stack tecnológico

| Capa | Tecnología |
|------|------------|
| Framework | React 19 + Vite 8 |
| Estilos | Tailwind CSS v4 (tokens: brand `#075809`, accent `#d2e749`) |
| Tipografía | Outfit (variable) |
| Iconos | Material Symbols Rounded |
| Rutas | Hash routing manual (`#/login`, `#/marketplace`, etc.) — sin librería externa |
| Autenticación | JWT en `localStorage`/`sessionStorage` |
| Linting | oxlint |

---

## Inicio rápido

**Requisitos:** Node.js ≥ 18, npm ≥ 9

```bash
# 1. Instalar dependencias
cd frontend
npm install

# 2. Configurar variable de entorno (opcional, tiene valor por defecto)
echo "VITE_API_URL=http://localhost:8080/api/v1" > .env

# 3. Correr en desarrollo
npm run dev

# 4. Build de producción
npm run build
```

**El backend (Go) debe estar corriendo por separado** con `docker compose up` en la raíz del proyecto para que la app funcione con datos reales. Sin backend, las pantallas de login, marketplace y dashboard mostrarán errores de conexión.

---

## Mapa de pantallas y roles

| Ruta | Pantalla | Descripción | Rol |
|------|----------|-------------|-----|
| `#/` | Landing | Página principal con hero, pasos de uso, CTA | Público |
| `#/login` | Login | Formulario de inicio de sesión | Público |
| `#/register` | Register | Formulario de registro (comprador o productor) | Público |
| `#/marketplace` | Marketplace | Catálogo de productos con filtros y búsqueda | Comprador |
| `#/product/:id` | ProductDetail | Detalle de producto + formulario de consulta | Comprador |
| `#/dashboard` | BuyerDashboard | Inicio, favoritos, mensajes, perfil del comprador | Comprador |
| `#/producer` | ProducerDashboard | Resumen, productos, solicitudes, negocio del productor | Productor |
| `#/admin` | AdminDashboard | Gestión de usuarios, productos, reportes | Administrador |

---

## Integración con backend

Todas las llamadas pasan por `src/services/api.js`. El wrapper `request()` agrega `Authorization: Bearer <jwt>`, maneja errores 401 (redirige a login), y desenrolla el envelope `{ data: ... }`.

### Autenticación

| Función | Endpoint | Método | Pantalla |
|---------|----------|--------|----------|
| `auth.login()` | `/auth/login` | POST | Login |
| `auth.register()` | `/auth/register` | POST | Register |

### Usuarios

| Función | Endpoint | Método | Pantalla |
|---------|----------|--------|----------|
| `users.getById()` | `/users/{id}` | GET | BuyerProfile, ProducerHome |
| `users.update()` | `/users/{id}` | PATCH | BuyerProfile |

### Empresas

| Función | Endpoint | Método | Pantalla |
|---------|----------|--------|----------|
| `companies.getByOwner()` | `/companies?owner_id=` | GET | ProducerBusiness |
| `companies.getById()` | `/companies/{id}` | GET | ProducerBusiness |
| `companies.create()` | `/companies` | POST | ProducerBusiness |
| `companies.update()` | `/companies/{id}` | PATCH | ProducerBusiness |

### Ofertas / Productos

| Función | Endpoint | Método | Pantalla |
|---------|----------|--------|----------|
| `offerings.getByCompany()` | `/offerings?company_id=` | GET | ProducerProducts |
| `offerings.getById()` | `/offerings/{id}` | GET | ProductDetail |
| `offerings.create()` | `/offerings` | POST | ProducerProducts |
| `offerings.update()` | `/offerings/{id}` | PATCH | ProducerProducts |
| `offerings.getFeatured()` | `/offerings?company_id=` (×N) | GET | MarketplaceCatalog |

### Consultas (Inquiries)

| Función | Endpoint | Método | Pantalla |
|---------|----------|--------|----------|
| `inquiries.create()` | `/inquiries` | POST | ProductDetail |
| `inquiries.getByUser()` | `/inquiries?user_id=` | GET | BuyerMessages |
| `inquiries.getByCompany()` | `/inquiries/company/{id}` | GET | ProducerRequests |
| `inquiries.updateStatus()` | `/inquiries/{id}` | PATCH | ProducerRequests |

### Categorías

| Función | Endpoint | Método | Pantalla |
|---------|----------|--------|----------|
| `categories.getAll()` | `/categories` | GET | FiltersSidebar |

---

## Limitaciones conocidas / pendientes

| Funcionalidad | Estado | Motivo |
|---------------|--------|--------|
| Admin Dashboard | Mock | El backend no tiene endpoints de administración (listar todos los usuarios, moderar productos, reportes agregados) |
| Chat en tiempo real | No implementado | No hay WebSocket en el backend. "Mensajes" muestra una lista de inquiries con estados, no un chat en vivo |
| Favoritos del comprador | Mock | No hay endpoint `GET /favorites` ni tabla de favoritos en el backend |
| Valoraciones / Reviews | Mock | El backend tiene tabla `reviews` pero el frontend aún no lo consume |
| Imágenes de productos | Local | Las imágenes se almacenan como base64 en el campo `description` del offering (no en un almacenamiento externo). El campo `image_url` tiene VARCHAR(2048) en la BD |
| Reseñas de productor | Mock | `ProducerHome` muestra valoración "4.8" hardcodeada |
| Notificaciones | No implementado | No hay sistema de notificaciones push ni email |
| Búsqueda de producto | Client-side | El filtro de búsqueda en Marketplace filtra en el array local, no hace query al backend |

---

## Convenciones de código

### Estructura de carpetas

```
src/
├── components/        # Componentes reutilizables
│   ├── dashboard/     # Sidebars de dashboards
│   ├── layout/        # Navbar, Footer
│   ├── product/       # ProductCardGrid, ProductCardList, ProductImage
│   └── ui/            # Badge, Button, Icon (baja complejidad)
├── config/            # Constantes (FEATURED_COMPANY_IDS)
├── hooks/             # Custom hooks (useOfferings, useUserProfile, etc.)
├── lib/               # Utilidades puras (session, format, user, productImages)
├── mocks/             # Datos mock (pendientes de conectar a backend)
├── pages/             # Páginas/rutas (Auth, Marketplace, BuyerDashboard, etc.)
│   ├── buyer/         # Subcomponentes del dashboard de comprador
│   └── producer/      # Subcomponentes del dashboard de productor
└── services/          # api.js (wrapper de fetch al backend)
```

### Paleta de colores (Tailwind v4)

| Token | Valor | Uso |
|-------|-------|-----|
| `brand` | `#075809` | Verde principal (botones, links, acentos) |
| `brand-dark` | `#05460a` | Hover de brand |
| `brand-soft` | `#e9f2e7` | Fondos suaves, badges |
| `accent` | `#d2e749` | Acentos amarillo-verde |
| `night` | `#1d1d1b` | Texto oscuro |

### Convenciones generales

- **Sin librería de rutas**: se usa hash routing manual en `App.jsx` con `resolveRoute()`
- **Auth**: el token JWT se guarda en `localStorage` y `sessionStorage`. `session.js` expone `getUser()`, `setUser()`, `hasRole()`, `getCompanyId()`
- **Nombres de usuario**: el backend usa `first_name`/`last_name` (no `name`). Para mostrar: `lib/user.js` → `getDisplayName(user)`
- **Localización**: C$ NIO, formato `es-NI`, teléfonos `+505`
- **Imágenes**: se embeden en `description` con formato `ImageBase64:<data_url>` para persistir sin modificar el esquema de la BD
