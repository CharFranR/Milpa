# Milpa API Reference

Milpa is a marketplace API that connects agricultural producers (MIPYMEs) with buyers: users publish **offerings**, buyers open **inquiries** and **conversations**, suppliers publish **liquidations** (open purchase requests), and admins moderate **reports**. The service is a Go HTTP API built on `chi/v5` with hexagonal architecture. Every route below is served under the version prefix **`/api/v1`**; for local development the base URL is:

```
http://localhost:8080/api/v1
```

This document is derived from the authoritative route table in `server/infrastructure/adapters/primary/api/router.go` (44 registrations / 43 distinct routes).

---

## Quick start

Three calls take you from zero to an authenticated request.

### 1. Register

```bash
curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{
    "email": "ana@example.com",
    "first_name": "Ana",
    "last_name": "Lopez",
    "role": 1,
    "password": "secret123",
    "confirm_password": "secret123",
    "address": "Barrio San Jose, Matagalpa",
    "phone_number": "+50588887777"
  }'
```

`201 Created`:

```json
{
  "data": {
    "id": "9f1c2a7e-2f4e-4a4a-9c6a-1b2c3d4e5f60",
    "email": "ana@example.com",
    "first_name": "Ana",
    "last_name": "Lopez",
    "address": "Barrio San Jose, Matagalpa",
    "phone_number": "+50588887777",
    "role": 1,
    "created_at": "2026-09-25T14:03:11Z",
    "updated_at": "2026-09-25T14:03:11Z"
  }
}
```

`role` must be `1` (MIPYME) or `2` (Provider); `0` and `3` are rejected with `400 invalid input`.

### 2. Log in

```bash
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email": "ana@example.com", "password": "secret123"}'
```

`200 OK`:

```json
{
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 86400,
    "user": { "id": "9f1c2a7e-2f4e-4a4a-9c6a-1b2c3d4e5f60", "role": 1 }
  }
}
```

The token is an HS256 JWT valid for **24 h** (`expires_in: 86400`).

### 3. Call an authenticated endpoint

```bash
TOKEN="<access_token from step 2>"

curl -s -X POST http://localhost:8080/api/v1/companies/ \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name": "Finca El Progreso", "address": "Matagalpa"}'
```

`201 Created` returns the created company (`{"data": { ...CompanyDTO... }}`). A missing/invalid token returns `401 {"error":"missing or invalid authorization header"}`.

---

## Authentication

| Mechanism | Where it applies | How to send it |
|---|---|---|
| Bearer JWT | Every route marked **Auth** in the reference | `Authorization: Bearer <access_token>` |
| WebSocket subprotocol | `GET /api/v1/ws/{conversationID}` only | `Sec-WebSocket-Protocol: milpa.chat.v1, bearer.<access_token>` |

**Obtaining a token:** only `POST /auth/login` issues one (`data.access_token`). `POST /auth/register` creates the user but does **not** return a token.

**Header format:** exactly `Bearer ` (uppercase `B`, single space) followed by the raw JWT. Any other scheme results in `401 {"error":"missing or invalid authorization header"}`. An expired/tampered JWT results in `401 {"error":"invalid or expired token"}`.

**Public vs. protected:** routes marked *Public* in the tables need no header. Protected routes run two middlewares in order:

1. `Authenticate` — parses the header, validates the JWT, and puts the principal (`user_id` + `role`) into the request context.
2. `CheckSuspension` — reloads the user; if the account is suspended it short-circuits with `403` and the body `{"error":"your account has been suspended"}` (note: written through `http.Error`, so the `Content-Type` is `text/plain; charset=utf-8`, not JSON).

**Role checks live in the use cases, not the router.** Several routes accept any authenticated user but return `403 {"error":"forbidden"}` for the wrong role (admin-only) or the wrong ownership/participant relationship.

### WebSocket authentication

Browsers cannot set headers on a WebSocket handshake, so the token travels in the `Sec-WebSocket-Protocol` header (`server/.../middleware/auth.go`, `AuthenticateWebSocket`):

1. The client requests **two** subprotocols: the protocol name `milpa.chat.v1` and an entry prefixed with `bearer.`:

   ```js
   const ws = new WebSocket(
     "ws://localhost:8080/api/v1/ws/<conversationID>",
     ["milpa.chat.v1", "bearer." + token]
   );
   ```

2. The middleware scans the requested subprotocol list in order, takes the **first** entry starting with `bearer.`, and uses everything after the prefix as the JWT.
3. The server's upgrader only negotiates `milpa.chat.v1`, so that entry must also be present — otherwise the handshake fails.
4. **Fallback:** if no `bearer.` entry is found, the middleware falls back to the `Authorization: Bearer` header (usable from non-browser clients such as `curl` or `websocat`). If neither is present: `401 {"error":"missing or invalid websocket credentials"}`.

---

## Conventions

### Content types

- All JSON endpoints: request `Content-Type: application/json`, response `Content-Type: application/json`.
- Exception: `POST /api/v1/offerings/create2/` takes `multipart/form-data`; `GET /api/v1/images/{filename}` returns an image content type with `Cache-Control: public, max-age=86400`.
- Exception: the suspension block returns `text/plain; charset=utf-8` with a JSON-looking body.

### Success envelope

```json
{ "data": <payload> }
```

- Lists and objects are wrapped as shown above.
- Endpoints that return no payload answer `200`/`201` with an **empty object** `{}` (the `data` key is omitted when `null`).
- `GET /images/{filename}` is the only route that returns a raw body instead of the envelope.

### Error envelope

```json
{ "error": "human-readable message" }
```

Exactly one key. Malformed JSON always yields `{"error":"invalid request body"}`.

### Status-code mapping

`httpx.StatusCode` maps domain errors to HTTP status (`server/.../httpx/httpx.go`):

| Condition | HTTP | Example body |
|---|---|---|
| Missing/malformed `Authorization` header | 401 | `{"error":"missing or invalid authorization header"}` |
| Invalid or expired JWT | 401 | `{"error":"invalid or expired token"}` |
| Missing WS credentials | 401 | `{"error":"missing or invalid websocket credentials"}` |
| `domain.ErrUnauthorized` (wrong password) | 401 | `{"error":"unauthorized"}` |
| Suspended account (suspension middleware) | 403 | `{"error":"your account has been suspended"}` |
| `domain.ErrForbidden` (role/ownership/participant) | 403 | `{"error":"forbidden"}` |
| `domain.ErrNotFound` | 404 | `{"error":"resource not found"}` |
| `domain.ErrDuplicate` / `ErrEmailTaken` | 409 | `{"error":"email already registered"}` |
| Validation errors (`httpx.IsValidationError` list) | 400 | `{"error":"rating must be between 1 and 5"}` |
| Handler-level input checks (bad UUID, blank field, unparsable body) | 400 | `{"error":"invalid user id"}`, `{"error":"email: cannot be blank"}` |
| Everything else (unknown/`fmt.Errorf` errors) | 500 | `{"error":"internal server error"}` (message logged server-side) |

> **Gotcha:** some business-rule failures are plain `fmt.Errorf` errors and therefore fall through to `500 internal server error` instead of `400` — e.g. report `reason` length outside 10–500, an invalid report `action`, an invalid suspend `action`, liquidation `quantity <= 0`, or a liquidation that is not open. See the notes at the end.

### CORS

Configured in `router.go`:

| Setting | Value |
|---|---|
| `AllowedOrigins` | `*` |
| `AllowedMethods` | `GET`, `POST`, `PATCH`, `DELETE`, `OPTIONS` |
| `AllowedHeaders` | `Accept`, `Authorization`, `Content-Type` |
| `ExposedHeaders` | `Link` |
| `AllowCredentials` | `false` |
| Preflight `MaxAge` | `300` s |

`PUT`/`HEAD` are not in `AllowedMethods`, so browsers will block them cross-origin.

### Enumerations (serialized as **integers**, except reports)

| Field | Values |
|---|---|
| `role` | `0` pending, `1` MIPYME, `2` Provider, `3` admin (register accepts `1`/`2` only) |
| `type` (offering) | `0` product, `1` service |
| `status` (inquiry) | `0` pending, `1` read, `2` replied, `3` closed |
| `status` (liquidation) | `0` open, `1` closed, `2` expired, `3` assigned |
| `allocation_method` | `0` manual |
| `status` / `target_type` / `action` (reports & moderation) | strings: `pending`\|`approved`\|`rejected`, `offering`\|`user`, `approve`\|`reject`, `suspend`\|`reactivate` |

---

## Endpoint reference

Auth column: **Public** = no token; **Bearer** = `Authorization` header + suspension check; **Bearer (WS)** = subprotocol/header token + suspension check.

### Auth

| Route | Auth | Request | Success | Notable statuses |
|---|---|---|---|---|
| `POST /api/v1/auth/register` | Public | JSON: `email`, `first_name`, `last_name`, `password`, `confirm_password`, `role` (1\|2); optional `address`, `phone_number` | `201` → `UserDTO` | `400` blank field / bad role / password mismatch / bad body; `409` email taken |
| `POST /api/v1/auth/login` | Public | JSON: `email`, `password` | `200` → `{access_token, expires_in, user}` | `400` blank field/bad body; `404` unknown email; `401` wrong password |

### Users

| Route | Auth | Params / body | Success | Notable statuses |
|---|---|---|---|---|
| `GET /api/v1/users/{id}` | Public | path `id` (uuid) | `200` → `UserDTO` | `400` invalid uuid; `404` |
| `PATCH /api/v1/users/{id}` | Bearer | path `id`; JSON (all optional): `email`, `first_name`, `last_name`, `address`, `phone_number` | `200` `{}` | `400`; `401`; `403` if `{id}` ≠ token user; `404` |

### Categories

| Route | Auth | Params | Success | Notable statuses |
|---|---|---|---|---|
| `GET /api/v1/categories` | Public | — | `200` → `[CategoryDTO]` | `500` |

### Companies

| Route | Auth | Params / body | Success | Notable statuses |
|---|---|---|---|---|
| `GET /api/v1/companies/{id}` | Public | path `id` (uuid) | `200` → `CompanyDTO` | `400`; `404` |
| `GET /api/v1/companies/` | Public | query `owner_id` (uuid, required) | `200` → `[CompanyDTO]` | `400` invalid/missing `owner_id` |
| `POST /api/v1/companies/` | Bearer | JSON: `name` (required); optional `category_id`, `address`, `description`, `phone_number`, `email`, `website`. Owner = token user | `201` → `CompanyDTO` | `400` blank name/bad body; `401`; `403` suspended; `404` unknown `category_id` |
| `PATCH /api/v1/companies/{id}` | Bearer | path `id`; JSON (all optional): `name`, `address`, `description`, `phone_number`, `email`, `website` | `200` `{}` | `403` if not the owner; `404` |

Collection routes are registered with a trailing slash; chi's mount also answers the same request without the trailing slash, but the form shown in the table is the registered one.

### Offerings

| Route | Auth | Params / body | Success | Notable statuses |
|---|---|---|---|---|
| `GET /api/v1/offerings/{id}` | Public | path `id` (uuid) | `200` → `OfferingDTO` | `400`; `404` |
| `GET /api/v1/offerings/` | Public | query `user_id` (uuid, required) | `200` → `[OfferingDTO]` | `400` invalid `user_id` |
| `POST /api/v1/offerings/` | Bearer | JSON: `user_id` (must equal token user), `name`, `type` (0\|1), `price`; optional `description`, `image_url` | `201` → `OfferingDTO` | `400` blank `user_id`/`name`; `403` `user_id` ≠ token user; `404` unknown user |
| `POST /api/v1/offerings/create2/` | Bearer | `multipart/form-data`: `user_id`, `type`, `name`, `price`, optional `description`, optional file `image_url` (≤ 10 MB) | `201` → `OfferingDTO` | `400` unparsable `user_id`/`type`/`price`, upload failure |
| `PATCH /api/v1/offerings/{id}` | Bearer | path `id` only — **this path is registered twice and the second handler (`DeleteOffering`) wins** | `200` `{}` (offering deleted) | `400` invalid uuid (handler keeps writing after the error); `404` |

> The intended update handler is registered first and is unreachable — see the notes.

### Reviews

| Route | Auth | Params / body | Success | Notable statuses |
|---|---|---|---|---|
| `GET /api/v1/reviews/` | Public | query `company_id` **or** `user_id` (uuid) | `200` → `[ReviewDTO]` | `400` missing both / invalid uuid |
| `POST /api/v1/reviews/` | Bearer | JSON: `company_id` (required), `rating` (1–5), `comment`. `user_id` = token user | `201` → `ReviewDTO` | `400` blank `company_id` / rating out of range |

### Inquiries

| Route | Auth | Params / body | Success | Notable statuses |
|---|---|---|---|---|
| `GET /api/v1/inquiries/company/{company_id}` | Public | path `company_id` (uuid) | `200` → `[InquiryDTO]` | `400`; `404` |
| `GET /api/v1/inquiries/{id}` | Public | path `id` (uuid) | `200` → `InquiryDTO` | `400`; `404` |
| `GET /api/v1/inquiries/` | Public | query `user_id` (uuid, required) | `200` → `[InquiryDTO]` | `400` invalid `user_id` |
| `POST /api/v1/inquiries/` | Bearer | JSON: `offering_id` (uuid), `message`. `user_id` = token user | `201` → `InquiryDTO` | `400` blank field; `401`; `403` |
| `PATCH /api/v1/inquiries/{id}` | Bearer | path `id`; JSON: `status` (1, 2 or 3) | `200` `{}` | `400` invalid uuid/body or `status: 0`; `404` |

### Liquidations

| Route | Auth | Params / body | Success | Notable statuses |
|---|---|---|---|---|
| `GET /api/v1/liquidations/open` | Public | — (matches before `{id}`) | `200` → `[LiquidationDTO]` (status open) | `500` |
| `GET /api/v1/liquidations/` | Public | query `supplier_id` (uuid, required) | `200` → `[LiquidationDTO]` | `400` invalid `supplier_id` |
| `GET /api/v1/liquidations/{id}` | Public | path `id` (uuid) | `200` → `LiquidationDTO` | `400`; `404` |
| `POST /api/v1/liquidations/` | Bearer | JSON: `product_name`, `quantity`, `unit_of_measure`, `total_price`, `unit_price` (all required); optional `delivery_time`, `location_id`, `visibility`, `expires_at`. Supplier = token user | `201` → `LiquidationDTO` | `400` blank field; `401`; `403` |
| `PATCH /api/v1/liquidations/{id}` | Bearer | path `id`; JSON (all optional): `product_name`, `quantity`, `unit_of_measure`, `total_price`, `unit_price`, `delivery_time`, `location_id`, `visibility`, `expires_at` | `200` `{}` | `403` if not supplier; `404`; `500` if not open / invalid quantity |
| `DELETE /api/v1/liquidations/{id}` | Bearer | path `id` | `200` `{}` | `403` if not supplier; `404` |

### Reports

| Route | Auth | Params / body | Success | Notable statuses |
|---|---|---|---|---|
| `POST /api/v1/reports/` | Bearer | JSON: `target_type` (`offering`\|`user`), `target_id` (uuid string), `reason` (10–500 chars). Reporter = token user | `201` → `ReportResponse` | `400` invalid target type / blank or duplicate-pending / self-report; `500` reason length violation |
| `GET /api/v1/reports/` | Bearer, **admin** | query: `status`, `target_type`, `page`, `page_size` (≤ 100, default 20) | `200` → `{items, total, page, size}` | `403` non-admin; `400` invalid `status`/`target_type` |
| `PATCH /api/v1/reports/{id}/action` | Bearer, **admin** | path `id`; JSON: `action` (`approve`\|`reject`). Approve suspends a reported user or deletes a reported offering | `200` → `ReportResponse` | `400` already resolved; `403` non-admin; `404`; `500` bad `action` |

### Conversations

| Route | Auth | Params / body | Success | Notable statuses |
|---|---|---|---|---|
| `GET /api/v1/conversations/` | Bearer | — (returns conversations where the token user is farmer or buyer) | `200` → `[ConversationDTO]` | `401`; `403` |
| `POST /api/v1/conversations/` | Bearer | JSON: `farmer_id` (uuid), `offering_id` (uuid). Buyer = token user | `201` → `ConversationDTO` | `400` if `farmer_id` == token user or missing offering/user; `404` |
| `GET /api/v1/conversations/{id}` | Bearer | path `id` | `200` → `ConversationDTO` | `403` not a participant; `404` |
| `DELETE /api/v1/conversations/{id}` | Bearer | path `id` | `200` `{}` | `403` not a participant; `404` |
| `GET /api/v1/conversations/{id}/messages` | Bearer | path `id` | `200` → `[MessageDTO]` | `403` not a participant; `404` |

### Messages

| Route | Auth | Params / body | Success | Notable statuses |
|---|---|---|---|---|
| `POST /api/v1/messages/` | Bearer | JSON: `conversation_id` (uuid), `content`. Sender = token user | `201` `{}` | `400` blank content; `403` not a participant; `404` unknown conversation |
| `DELETE /api/v1/messages/{id}` | Bearer | path `id` (message id) | `200` `{}` | `403` if sender ≠ token user; `404` |

### WebSocket

| Route | Auth | Params | Success | Notable statuses |
|---|---|---|---|---|
| `GET /api/v1/ws/{conversationID}` | Bearer (WS) | path `conversationID` (uuid) | `101 Switching Protocols` | `401` no/invalid credentials; `400` unparsable id; `403` suspended or not a participant; `404` unknown conversation |

Full handshake and message details below.

### Images

| Route | Auth | Params | Success | Notable statuses |
|---|---|---|---|---|
| `GET /api/v1/images/{filename}` | Public | path `filename` (single path segment, resolved under `./uploads`) | `200` raw bytes with the detected `Content-Type` + `Cache-Control: public, max-age=86400` | `404` `{"error":"could not find image"}` |

Images are uploaded through `POST /api/v1/offerings/create2/` (field `image_url`); the stored value is the on-disk path and the public read path is `/api/v1/images/<filename>`.

### Search

| Route | Auth | Query params | Success | Notable statuses |
|---|---|---|---|---|
| `GET /api/v1/search` | Public | `term`, `type`, `department`, `municipality`, `price_min`, `price_max`, `farmer_id`, `sort` (`relevance`\|`price_asc`\|`price_desc`\|`proximity`), `lat`, `lng`, `page`, `page_size` | `200` → `{results, total_hits, page, page_size, total_pages}` | `400` `{"error":"Search error"}` on any backend failure |

Unparsable numeric params are silently ignored (they are skipped, not rejected).

### Admin / Moderation

| Route | Auth | Params / body | Success | Notable statuses |
|---|---|---|---|---|
| `PATCH /api/v1/admin/users/{id}/suspend` | Bearer, **admin** | path `id`; JSON: `action` (`suspend`\|`reactivate`) | `200` `{}` | `403` non-admin; `404`; `400` self-suspend / suspend-an-admin; `500` bad `action` |
| `DELETE /api/v1/admin/offerings/{id}` | Bearer, **admin** | path `id` | `204` (body written as `{}`) | `403` non-admin; `404` |
| `GET /api/v1/admin/audit-logs` | Bearer, **admin** | query: `action`, `actor_id`, `target_type`, `page`, `page_size` (≤ 100, default 20) | `200` → `{items, total, page, size}` of `AuditLogResponse` | `403` non-admin |

Known audit `action` values: `report_created`, `report_approved`, `report_rejected`, `user_suspended`, `user_reactivated`, `offering_deleted`.

---

## WebSocket

**Connect**

```
ws://localhost:8080/api/v1/ws/{conversationID}
```

**Subprotocol auth (browser):**

```js
const ws = new WebSocket(
  `ws://localhost:8080/api/v1/ws/${conversationID}`,
  ["milpa.chat.v1", `bearer.${token}`]
);
```

- Requested protocols: `milpa.chat.v1` (the only one the server negotiates) plus `bearer.<JWT>` — the token entry is consumed by the auth middleware and does not need to be negotiated.
- Non-browser clients may instead send `Authorization: Bearer <JWT>`.
- Before the upgrade, the server verifies the conversation exists **and** the caller is a participant; failures answer with a normal JSON error (`401`/`403`/`404`) instead of `101`.

**Client → server:** a raw **text frame** whose payload becomes `content` (max 4096 bytes). The server wraps it as a `MessageDTO` with `conversation_id` from the path and `sender_id` from the token, persists it, and broadcasts it.

**Server → client:** JSON `MessageDTO` frames:

```json
{
  "id": "…uuid…",
  "conversation_id": "…uuid…",
  "sender_id": "…uuid…",
  "content": "hello",
  "visibility": true,
  "created_at": "2026-09-25T14:03:11Z"
}
```

**On connect:** the server immediately broadcasts a sponsor notice to the conversation — `{"content": "Espacio disponible gracias a Hackaton Nicaragua"}` with a fresh `id`, no `sender_id`, and **not persisted**.

**Keep-alive:** server ping control frames every 54 s; read deadline 60 s (refreshed on pong); write deadline 10 s. A frame larger than 4096 bytes closes the connection. Broadcasts are per-conversation: only clients joined to the same `conversationID` receive a message.

---

## DTO reference

Exact JSON shapes (field names as implemented in `server/aplication/dto/`).

| DTO | Fields |
|---|---|
| `UserDTO` | `id`, `email`, `first_name`, `last_name`, `address`, `phone_number`, `role`, `created_at`, `updated_at` |
| `LoginResponse` | `access_token`, `expires_in`, `user` (`UserDTO`) |
| `CompanyDTO` | `id`, `name`, `category_id`, `owner_id`, `address`, `description`, `phone_number`, `email`, `website`, `verified`, `created_at`, `updated_at` |
| `CategoryDTO` | `id`, `name`, `description` |
| `OfferingDTO` | `id`, `user_id`, `type`, `name`, `description`, `price`, `image_url`, `created_at`, `updated_at` |
| `ReviewDTO` | `id`, `user_id`, `company_id`, `rating`, `comment`, `created_at` |
| `InquiryDTO` | `id`, `user_id`, `offering_id`, `offering_name`, `message`, `status`, `created_at` |
| `LiquidationDTO` | `id`, `supplier_id`, `product_name`, `quantity`, `unit_of_measure`, `total_price`, `unit_price`, `delivery_time`, `location_id`, `visibility`, `allocation_method`, `status`, `closed_at?`, `expires_at?`, `created_at`, `updated_at` |
| `ConversationDTO` | `id`, `farmer_id`, `buyer_id`, `offering_id`, `visibility`, `created_at`, `updated_at` |
| `MessageDTO` | `id`, `conversation_id`, `sender_id`, `content`, `visibility`, `created_at` |
| `ReportResponse` | `id`, `reporter` (`{id,name,email}`), `target_type`, `target` (`{id,name}`), `reason`, `status`, `resolved_by?`, `resolved_at?`, `created_at` |
| `AuditLogResponse` | `id`, `actor_id`, `action`, `target_type`, `target_id`, `metadata?`, `created_at` |
| Paginated reports / audit logs | `items`, `total`, `page`, `size` |
| `SearchResponse` | `results` (`{id,name,description,price,type,image_url,farmer_id,farmer_name,farmer_verified,department,municipality,latitude,longitude}`), `total_hits`, `page`, `page_size`, `total_pages` |

---

## Checklist / notes

Derived from `server/infrastructure/adapters/primary/api/router.go`; nothing in this document is a route that is not registered there.

- [ ] **Route count:** `router.go` has **44** route registrations; `chi` resolves them into **43** distinct method+path routes because `PATCH /api/v1/offerings/{id}` is registered twice (`Update`, then `DeleteOffering`).
- [ ] **`PATCH /api/v1/offerings/{id}` behaves as delete.** chi's tree keeps the last handler written for a method+pattern, so `DeleteOffering` wins and `OfferingHandler.Update` is unreachable. There is currently **no working "update an offering" endpoint** despite the handler existing.
- [ ] **No ownership check on offering delete.** `OfferingUseCase.DeleteOffering` never inspects the principal, so any authenticated (non-suspended) user can delete any offering through that PATCH route. The admin route `DELETE /admin/offerings/{id}` is the audited path.
- [ ] **`PATCH /api/v1/inquiries/{id}` has no ownership check** either — any authenticated user can change any inquiry's status.
- [ ] **Handler bugs worth knowing:** `OfferingHandler.DeleteOffering` writes a `400` for an invalid uuid but does not `return`, producing a second response write; `POST /api/v1/messages/` and several other creates answer `201` with an empty `{}` body; `DELETE /admin/offerings/{id}` answers `204` while still writing a `{}` body.
- [ ] **Several domain errors surface as `500`** instead of `400` because they are plain `fmt.Errorf`/unlisted sentinels (report reason length, invalid report/suspend `action`, liquidation quantity/status rules).
- [ ] **Trailing slashes matter** for `POST /api/v1/offerings/create2/` (registered with a trailing slash). The collection routes (`/companies/`, `/offerings/`, `/reviews/`, `/inquiries/`, `/liquidations/`, `/reports/`, `/conversations/`, `/messages/`) also answer without the trailing slash thanks to chi's mount behavior.
- [ ] **Ordering matters** where static and parameterized paths coexist: `GET /liquidations/open` wins over `GET /liquidations/{id}`, and `GET /inquiries/company/{company_id}` wins over `GET /inquiries/{id}` (chi matches static segments first).
- [ ] **Images:** the stored/read filename comes from the multipart upload's original name and is joined onto `./uploads`; treat `GET /images/{filename}` as serving only single-segment names.

### Server facts

| Item | Value |
|---|---|
| Port | `SERVER_PORT`, default `8080` (`server/cmd/api/main.go`) |
| Required env | `JWT_SECRET` (fatal if missing), Postgres (`POSTGRES_*`, `DB_SSLMODE`), Redis (`REDIS_URL`/`REDIS_HOST`+`REDIS_PORT`), Elasticsearch (`ESCLIENT_*`) |
| Timeouts | read 10 s, write 15 s, idle 60 s |
| Migrations | run automatically at startup |
| Image store | local directory `./uploads` |
