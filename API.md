# Milpa API Reference

Milpa is a marketplace API that connects agricultural producers (MIPYMEs) with buyers: users publish **offerings**, buyers open **inquiries** and **conversations**, suppliers publish **liquidations** (open purchase requests), and admins moderate **reports**. The service is a Go HTTP API built on `chi/v5` with hexagonal architecture. Every route below is served under the version prefix **`/api/v1`**; for local development the base URL is:

```
http://localhost:8080/api/v1
```

This document is derived from the authoritative route table in `server/infrastructure/adapters/primary/api/router.go` (70 registrations / 69 distinct routes; 65 registered inside `NewRouter` and 5 by the separate `RegisterTransactionRoutes` entry point).

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

### The supply chain, end to end

The supply-chain surface is a multi-step sequence between two accounts, so it is worth walking through once. `BUYER` and `SUPPLIER` below are two different tokens; the roles do not gate the routes, ownership does.

**1. The buyer opens a request** — `POST /api/v1/supply-requests/`, as `BUYER`:

```bash
curl -s -X POST http://localhost:8080/api/v1/supply-requests/ \
  -H "Authorization: Bearer $BUYER" \
  -H 'Content-Type: application/json' \
  -d '{
    "product_name": "Cafe",
    "total_amount": 500,
    "amount_unit": 0,
    "unit_of_measure": 0,
    "description": "Cosecha 2026",
    "address": {
      "Department": "Matagalpa",
      "Municipality": "Matagalpa",
      "AddressLine": "Barrio San Jose"
    },
    "request_deadline": "2026-10-01T00:00:00Z",
    "delivery_deadline": "2026-11-01T00:00:00Z",
    "multiple_providers": true,
    "min_amount_per_provider": 100
  }'
```

`201 Created` returns the `SupplyRequestDTO`; note its `id` and that `actual_amount` comes back equal to `total_amount`.

**2. The supplier offers** — `POST /api/v1/supply-offers/`, as `SUPPLIER`. First it can find open work with `GET /api/v1/supply-requests/available`, which hides the caller's own requests and anything already offered on:

```bash
curl -s -X POST http://localhost:8080/api/v1/supply-offers/ \
  -H "Authorization: Bearer $SUPPLIER" \
  -H 'Content-Type: application/json' \
  -d '{
    "supply_request_id": "<request id from step 1>",
    "total_amount": 500,
    "measurement": 0,
    "delivery_day": "2026-10-20T00:00:00Z",
    "delivery_available": true
  }'
```

`201 Created` returns the `SupplyOfferDTO`; note its `id`.

**3. The buyer likes the offer** — `POST /api/v1/matches/like/{offerID}`, as `BUYER`. This is the single point where the match and the transaction are born, so the `201` carries both:

```bash
curl -s -X POST "http://localhost:8080/api/v1/matches/like/$OFFER_ID" \
  -H "Authorization: Bearer $BUYER"
```

`201 Created` → `MatchCreatedDTO`. Take `data.match.id` and `data.transaction.id` from the response. The offer moves to `status: 1` and the request's `actual_amount` drops by the matched amount.

> `GET /api/v1/matches/requests/{requestID}/prioritized` is the buyer's ranked shortlist of the remaining offers — worth calling before step 3, not after.

**4. Both confirm the start** — `POST /api/v1/transactions/{transaction_id}/confirm-start`, called once by each side. The first call leaves the transaction at `status: 0`; the second moves it to `1`:

```bash
curl -s -X POST "http://localhost:8080/api/v1/transactions/$TX_ID/confirm-start" \
  -H "Authorization: Bearer $BUYER"
curl -s -X POST "http://localhost:8080/api/v1/transactions/$TX_ID/confirm-start" \
  -H "Authorization: Bearer $SUPPLIER"
```

**5. Both confirm the delivery** — same route with `/confirm-delivery`, again once per side. The second call completes the transaction (`status: 2`):

```bash
curl -s -X POST "http://localhost:8080/api/v1/transactions/$TX_ID/confirm-delivery" \
  -H "Authorization: Bearer $BUYER"
curl -s -X POST "http://localhost:8080/api/v1/transactions/$TX_ID/confirm-delivery" \
  -H "Authorization: Bearer $SUPPLIER"
```

**6. The request closes itself** — no call does this. Inside the second delivery confirmation, the completed transactions of the request are summed and, once they cover `total_amount`, the request is completed in the same unit of work: `GET /api/v1/supply-requests/{id}` now answers `status: 2` with `actual_amount: 0`.

If the deal falls apart instead, `POST /api/v1/transactions/{transaction_id}/cancel` with a non-blank `{"reason": "..."}` releases everything atomically: transaction, match, the request's `actual_amount`, and the offer (which goes back to `status: 0` so it can be offered or matched again).

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
| `domain.ErrInvalidRequestStatus` / `ErrInvalidOfferStatus` / `ErrInvalidMatchStatus` | 409 | `{"error":"invalid supply request status"}` |
| `domain.ErrInsufficientAmount` | 409 | `{"error":"insufficient amount: ..."}` |
| `primary.ErrActiveMatch` | 409 | `{"error":"supply request already has an active match"}` |
| `domain.ErrInvalidTransactionTransition` / `ErrAlreadyConfirmed` / `ErrTerminalState` | 409 | `{"error":"participant has already confirmed"}` |
| Validation errors (`httpx.IsValidationError` list) | 400 | `{"error":"rating must be between 1 and 5"}` |
| Handler-level input checks (bad UUID, blank field, unparsable body) | 400 | `{"error":"invalid user id"}`, `{"error":"email: cannot be blank"}` |
| Everything else (unknown/`fmt.Errorf` errors) | 500 | `{"error":"internal server error"}` (message logged server-side) |

> The supply-chain handlers do not rely on that table alone: `handleSupplyError` and `handleMatchError` pre-empt `409` for the sentinels above before delegating, and `handleTransactionError` has its own 409 set. That is why an out-of-order state machine step is a `409` and not a `500`. Two consequences worth knowing: the transaction handler's 409 set does **not** include `ErrInsufficientAmount`, and `repository.ErrAmountConstraint` is in no list at all, so both fall through to `500` if they ever reach the boundary. The messages in the 409 rows are the sentinel's own text; the use cases usually wrap them with detail (`"insufficient amount: offer total amount exceeds the remaining amount of the supply request"`).

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
| `amount_unit` / `unit_of_measure` / `measurement` (offer) | `0` Kg, `1` Lb, `2` Tn — one `MeasurementOptions` enum reused by all three keys |
| `status` (supply request) | `0` open, `1` cancelled, `2` completed, `3` expired |
| `status` (supply offer) | `0` active, `1` matched, `2` rejected, `3` withdrawn |
| `status` (match) | `0` active, `1` cancelled |
| `status` (transaction) | `0` matched, `1` in_progress, `2` completed, `3` cancelled |
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

### Supply requests

Buyer-owned purchase requests. Every route runs the `Authenticate` + `CheckSuspension` chain, so the Auth column is `Bearer` throughout; ownership is then checked in the use case, not the router.

| Route | Auth | Params / body | Success | Notable statuses |
|---|---|---|---|---|
| `GET /api/v1/supply-requests/` | Bearer | — (requests whose `buyer_id` is the token user, all statuses) | `200` → `[SupplyRequestDTO]` | `401`; `403` suspended |
| `POST /api/v1/supply-requests/` | Bearer | JSON: `product_name` (required, non-blank), `total_amount` (required, > 0); optional `amount_unit`, `number_of_units`, `amount_per_unit`, `unit_of_measure`, `address`, `request_deadline`, `delivery_deadline`, `description`, `multiple_providers`, `min_amount_per_provider`. Buyer = token user | `201` → `SupplyRequestDTO` | `400` blank `product_name`, `total_amount <= 0`, `request_deadline` after `delivery_deadline`, bad body; `401`; `403` suspended |
| `GET /api/v1/supply-requests/available` | Bearer | — (open requests from *other* buyers with `actual_amount > 0`, minus any the caller already offered on) | `200` → `[SupplyRequestDTO]` | `401`; `403` suspended |
| `GET /api/v1/supply-requests/{id}` | Bearer | path `id` (uuid) — any authenticated user, no ownership check | `200` → `SupplyRequestDTO` | `400` invalid uuid; `404` |
| `PATCH /api/v1/supply-requests/{id}` | Bearer | path `id`; JSON: `SupplyGeneralUpdateDTO` — `product_name` and `total_amount` are **required** (the handler overwrites every other field, so omitting one zeroes it) | `200` `{}` | `400` validation; `403` not the buyer; `404`; `409` request not open |
| `PATCH /api/v1/supply-requests/{id}/amounts` | Bearer | path `id`; JSON: `SupplyUpdateAmountsDTO` — `total_amount` (required, > 0), `actual_amount` (required, 0…`total_amount`) | `200` `{}` | `400` `actual_amount` out of range; `403` not the buyer; `404`; `409` request not open |
| `PATCH /api/v1/supply-requests/{id}/deadlines` | Bearer | path `id`; JSON: `SupplyUpdateTimeDTO` — `request_deadline`, `delivery_deadline` | `200` `{}` | `400` `request_deadline` after `delivery_deadline`; `403` not the buyer; `404`; `409` request not open |
| `POST /api/v1/supply-requests/{id}/cancel` | Bearer | path `id`; no body | `200` `{}` | `403` not the buyer; `404`; `409` not open, or an active match exists |
| `POST /api/v1/supply-requests/{id}/expire` | Bearer | path `id`; no body | `200` `{}` | `403` not the buyer; `404`; `409` not open, or an active match exists |

> Both amount-changing routes are additionally clamped against the amount already committed by active matches: `total_amount` may not drop below it, and `actual_amount` may not reach into it (`409 insufficient amount`). The three `PATCH` routes are not a full replacement — `Update` rewrites the whole row, `UpdateAmounts` and `UpdateDeadlines` touch only their own columns.

### Supply offers

| Route | Auth | Params / body | Success | Notable statuses |
|---|---|---|---|---|
| `GET /api/v1/supply-offers/` | Bearer | — (offers whose `supplier_id` is the token user, all statuses) | `200` → `[SupplyOfferDTO]` | `401`; `403` suspended |
| `POST /api/v1/supply-offers/` | Bearer | JSON: `supply_request_id` (required, uuid), `total_amount` (required, > 0); optional `measurement`, `delivery_day`, `delivery_available`. Supplier = token user | `201` → `SupplyOfferDTO` | `400` missing request id / `total_amount <= 0` / below `min_amount_per_provider`; `401`; `403` offering on your own request; `404` unknown request; `409` request not open, duplicate offer from this supplier, active match on a single-provider request, or offer above the remaining amount |
| `GET /api/v1/supply-offers/requests/{request_id}` | Bearer | path `request_id` (uuid) — **rejected offers are filtered out** | `200` → `[SupplyOfferDTO]` | `400` invalid uuid; `403` not the request's buyer; `404` |
| `GET /api/v1/supply-offers/{id}` | Bearer | path `id` (uuid) — visible to the offer's supplier **or** the buyer of the underlying request | `200` → `SupplyOfferDTO` | `400` invalid uuid; `403` neither party; `404` |
| `PATCH /api/v1/supply-offers/{id}` | Bearer | path `id`; JSON: `SupplyOfferUpdateDTO` — `total_amount` (required, > 0), `measurement`, `delivery_day`, `delivery_available` | `200` `{}` | `400` `total_amount <= 0` / below `min_amount_per_provider`; `403` not the supplier; `404`; `409` offer not active, or amount above the request's remaining amount |
| `POST /api/v1/supply-offers/{id}/withdraw` | Bearer | path `id`; no body | `200` `{}` | `403` not the supplier; `404`; `409` offer not active |

> One offer per supplier per request: a second `POST` on the same pair is `409 resource already exists`. The JSON keys here are the crossed ones — `measurement` (not `amount_unit`) and `delivery_day` (not `proposed_delivery_day`) — unlike the supply request DTOs, which were corrected. See the notes.

### Matches

| Route | Auth | Params / body | Success | Notable statuses |
|---|---|---|---|---|
| `POST /api/v1/matches/like/{offerID}` | Bearer | path `offerID` (uuid) — the buyer of the offer's request, no body | `201` → `MatchCreatedDTO` (`{match, transaction}`) | `400` invalid uuid; `403` not the request's buyer; `404` unknown offer; `409` offer not actionable, request not open, offer already matched, or amount above the remaining amount |
| `POST /api/v1/matches/pass/{offerID}` | Bearer | path `offerID` (uuid) — the buyer, no body | `200` `{}` | `400` invalid uuid; `403` not the request's buyer; `404`; `409` offer not actionable |
| `GET /api/v1/matches/requests/{requestID}` | Bearer | path `requestID` (uuid) — buyer only | `200` → `[MatchDTO]` | `400` invalid uuid; `403` not the buyer; `404` |
| `GET /api/v1/matches/requests/{requestID}/prioritized` | Bearer | path `requestID` (uuid) — buyer only | `200` → `[PrioritizedOfferDTO]` (actionable offers, ranked) | `400` invalid uuid; `403` not the buyer; `404` |
| `GET /api/v1/matches/{matchID}` | Bearer | path `matchID` (uuid) — buyer of the request or the matched supplier | `200` → `MatchDTO` | `400` invalid uuid; `403` neither party; `404` |

> `like` is the only place a match *and* its transaction are created, and it does so as one unit of work: reservation, match, transaction and offer status all become visible or none do. On a `multiple_providers: false` request it is also the last one — the competing offers on that request are rejected in the same transaction. The `transaction` inside `MatchCreatedDTO` is a partial DTO (only `id`, `match_id`, `status`; every timestamp is `null` and `history` is `null`); read the full one from `GET /api/v1/transactions/matches/{match_id}`.

`PrioritizedOfferDTO` is ordered by `score` descending, then by the offer's `created_at` ascending, then by id. The only configured factor is `availability` at weight `1` (`recommendation.go`, `DefaultScoreFactors`), so `score` currently equals `available_quantity`.

### Recommendations

| Route | Auth | Query params | Success | Notable statuses |
|---|---|---|---|---|
| `GET /api/v1/recommendations/availability` | Bearer | `supplier_id` (uuid, **required** and must equal the token user), `product_name` (required, non-blank) | `200` → `AvailabilityDTO` | `400` invalid/missing `supplier_id` or blank `product_name`; `403` `supplier_id` ≠ token user; `404` no inventory row for that supplier+product; `401` |

> `available_quantity` is `inventory.quantity` minus the `matched_amount` of every active match of that supplier on the same product and the same `amount_unit`, floored at `0`. The `404` is deliberate on this route only: the internal helper swallows a missing inventory row and treats it as `0`, but the public read reports it.

### Transactions

Registered by the separate `RegisterTransactionRoutes` entry point, called on the same `chi.Mux` right after `NewRouter` (`cmd/api/main.go`), with the same `Authenticate` + `CheckSuspension` chain. Only the request's **buyer** and the matched offer's **supplier** are participants; anyone else gets `403`.

| Route | Auth | Params / body | Success | Notable statuses |
|---|---|---|---|---|
| `GET /api/v1/transactions/matches/{match_id}` | Bearer | path `match_id` (uuid) — either participant | `200` → `TransactionDTO` | `400` invalid uuid; `403` not a participant; `404` |
| `GET /api/v1/transactions/requests/{request_id}` | Bearer | path `request_id` (uuid) — the buyer sees all; a supplier sees only the transactions of their own matches | `200` → `[TransactionDTO]` | `400` invalid uuid; `403` not a participant, **and** `403` for a supplier whose filtered set comes out empty; `404` |
| `POST /api/v1/transactions/{transaction_id}/confirm-start` | Bearer | path `transaction_id` (uuid); no body. Each participant calls it once | `200` `{}` | `400` invalid uuid; `403` not a participant; `404`; `409` wrong status (not `matched`), or this participant already confirmed |
| `POST /api/v1/transactions/{transaction_id}/confirm-delivery` | Bearer | path `transaction_id` (uuid); no body. Each participant calls it once | `200` `{}` | `400` invalid uuid; `403` not a participant; `404`; `409` wrong status (not `in_progress`), or this participant already confirmed |
| `POST /api/v1/transactions/{transaction_id}/cancel` | Bearer | path `transaction_id` (uuid); JSON: `{"reason": "..."}` — **required and non-blank** | `200` `{}` | `400` blank/missing `reason`; `403` not a participant; `404`; `409` transaction already completed or cancelled |

> The status ladder is `0` matched → `1` in_progress → `2` completed, and it only advances when **both** participants have acted, so a single call is a no-op on the status. `confirm-delivery` runs as one unit of work with an auto-completion check: when the completed transactions of the request cover its `total_amount`, the request is completed and its `actual_amount` zeroed in the same transaction. `cancel` cascades the inverse atomically — transaction, match, the released `actual_amount`, and the matched offer returned to `status: 0`. None of the three takes a body other than `cancel`'s `reason`.

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
| `SupplyRequestDTO` | `id`, `buyer_id`, `product_name`, `total_amount`, `actual_amount`, `amount_unit`, `number_of_units`, `amount_per_unit`, `unit_of_measure`, `address`, `request_deadline`, `delivery_deadline`, `description`, `multiple_providers`, `min_amount_per_provider`, `status`, `created_at`, `updated_at` |
| `SupplyGeneralUpdateDTO` (`PATCH /supply-requests/{id}`) | `id`, `product_name`, `total_amount`, `actual_amount`, `amount_unit`, `number_of_units`, `amount_per_unit`, `unit_of_measure`, `address`, `request_deadline`, `delivery_deadline`, `description`, `multiple_providers`, `min_amount_per_provider` |
| `SupplyUpdateAmountsDTO` (`PATCH /supply-requests/{id}/amounts`) | `id`, `total_amount`, `actual_amount`, `amount_unit`, `amount_per_unit`, `unit_of_measure`, `multiple_providers`, `min_amount_per_provider` |
| `SupplyUpdateTimeDTO` (`PATCH /supply-requests/{id}/deadlines`) | `id`, `request_deadline`, `delivery_deadline` |
| `SupplyOfferDTO` | `id`, `supplier_id`, `supply_request_id`, `total_amount`, `measurement`, `delivery_day`, `delivery_available`, `status`, `created_at`, `updated_at` |
| `SupplyOfferUpdateDTO` (`PATCH /supply-offers/{id}`) | `total_amount`, `measurement`, `delivery_day`, `delivery_available` |
| `MatchDTO` | `id`, `supply_offer`, `supply_request`, `status`, `matched_amount`, `amount_unit`, `created_at`, `updated_at` |
| `MatchCreatedDTO` (`POST /matches/like/{offerID}`) | `match` (`MatchDTO`), `transaction` (`TransactionDTO`, partial — see the Matches notes) |
| `TransactionDTO` | `id`, `match_id`, `status`, `buyer_start_confirmed_at`, `supplier_start_confirmed_at`, `buyer_delivery_confirmed_at`, `supplier_delivery_confirmed_at`, `cancelled_by`, `cancel_reason`, `created_at`, `updated_at`, `history` |
| `TransactionHistoryEntryDTO` (inside `history`) | `status`, `at`, `cancel_reason?`, `cancelled_by?` |
| `TransactionCancelDTO` (`POST /transactions/{id}/cancel`) | `reason` |
| `PrioritizedOfferDTO` | `offer` (`SupplyOfferDTO`), `score`, `available_quantity`, `contributions` |
| `ScoreContributionDTO` (inside `contributions`) | `factor`, `weight`, `score`, `weighted_score` |
| `AvailabilityDTO` | `supplier_id`, `product_name`, `available_quantity` |

### Money, units and the `address` object

- **Amounts are `float64`.** Every amount, score and quantity on this surface is an IEEE-754 double, serialized as a JSON number — not a fixed-point or minor-unit integer. `SupplyRequestDTO` carries both `total_amount` (what the buyer asked for) and `actual_amount` (what is still **un-reserved**); `actual_amount` starts equal to `total_amount`, drops by every `matched_amount`, comes back up on a transaction cancel, and is zeroed when the request completes.
- **`amount_unit` vs `unit_of_measure`.** Both are the same `MeasurementOptions` enum (`0` Kg, `1` Lb, `2` Tn). `amount_unit` is the unit the amounts are counted in; `unit_of_measure` is the unit of the physical goods.
- **`address` has no JSON tags.** `domain.Address` declares no struct tags and no custom marshaler, so the object serializes with its **Go field names**: `{"ID","Department","Municipality","AddressLine","Latitude","Longitude"}` — capitalized, unlike every other key in this document. See the notes.

---

## Checklist / notes

Derived from `server/infrastructure/adapters/primary/api/router.go`; nothing in this document is a route that is not registered there.

- [ ] **Route count:** `router.go` has **70** route registrations; `chi` resolves them into **69** distinct method+path routes because `PATCH /api/v1/offerings/{id}` is registered twice (`Update`, then `DeleteOffering`). The registrations split **65** inside `NewRouter` and **5** inside the separate `RegisterTransactionRoutes`, which is called on the same mux from `cmd/api/main.go:178` — counting only `NewRouter` undercounts the table by five. Sprint 3 added 26 routes (9 supply requests, 6 supply offers, 5 matches, 1 recommendation, 5 transactions) to the 44 the previous revision of this document counted.
- [ ] **`PATCH /api/v1/offerings/{id}` behaves as delete.** chi's tree keeps the last handler written for a method+pattern, so `DeleteOffering` wins and `OfferingHandler.Update` is unreachable. There is currently **no working "update an offering" endpoint** despite the handler existing. Still unfixed: `router.go:72` and `router.go:73`.
- [ ] **No ownership check on offering delete.** `OfferingUseCase.DeleteOffering` never inspects the principal, so any authenticated (non-suspended) user can delete any offering through that PATCH route. The admin route `DELETE /admin/offerings/{id}` is the audited path.
- [ ] **`PATCH /api/v1/inquiries/{id}` has no ownership check** either — any authenticated user can change any inquiry's status.
- [ ] **Handler bugs worth knowing:** `OfferingHandler.DeleteOffering` writes a `400` for an invalid uuid but does not `return`, producing a second response write; `POST /api/v1/messages/` and several other creates answer `201` with an empty `{}` body; `DELETE /admin/offerings/{id}` answers `204` while still writing a `{}` body.
- [ ] **Several domain errors surface as `500`** instead of `400` because they are plain `fmt.Errorf`/unlisted sentinels (report reason length, invalid report/suspend `action`, liquidation quantity/status rules).
- [ ] **Breaking change — the supply request JSON keys were renamed.** `dto.SupplyRequestDTO`, `SupplyGeneralUpdateDTO` and `SupplyUpdateAmountsDTO` previously exposed crossed and misspelled keys, and the crossed pair was a duplicate-key bug: `amount_measure` for the enum, `amount_unit` for the per-unit float, `unit_measure`, plus `numer_units`, `Addrres` and `min_amount_provider`. They are now `amount_unit` (the enum), `amount_per_unit`, `number_of_units`, `unit_of_measure`, `address` and `min_amount_per_provider` — the snake_case of the Go field. Clients written against the old contract must be updated; `git show eae0fd8 -- server/aplication/dto/SupplyRequest.go` is the exact diff.
- [ ] **Known inconsistency — `SupplyOfferDTO` was not renamed with it.** The offer side still carries the crossed keys: `AmountUnit` serializes as `measurement` and `ProposedDeliveryDay` as `delivery_day` (`aplication/dto/SupplyOffer.go:15-16`, and the same pair in `SupplyOfferUpdateDTO` at :25-26). This is documented as-is, not as a bug to expect to be fixed: reading an offer's unit means reading `measurement`, and the request it belongs to uses `amount_unit` for the same enum.
- [ ] **`address` serializes with Go field names.** `domain.Address` carries no JSON tags and no custom marshaler, so `SupplyRequestDTO.address` is `{"ID","Department","Municipality","AddressLine","Latitude","Longitude"}` — capitalized, unlike every other key here. A client that lowercases keys will silently read an all-zero object.
- [ ] **The availability endpoint enforces ownership, and that is deliberate.** `GET /api/v1/recommendations/availability` requires `supplier_id` to equal the token user, so it answers `403` for anyone else (`use-cases/recommendation.go:77-79`). The internal helper `availableQuantity` stays ungated on purpose (comment at :74-76): the buyer of a request is entitled to see a candidate supplier's stock, and that buyer reads it through `GET /api/v1/matches/requests/{requestID}/prioritized`, which calls `RankOffers` with `strictNotFound = false` (:110) instead of the public read's `true` (:81). Two consequences: a supplier with no inventory row gets `404` on the public route but is ranked with `available_quantity: 0` on the buyer's route.
- [ ] **A buyer's offer list no longer shows passed offers.** `GET /api/v1/supply-offers/requests/{request_id}` — the buyer's view of the offers on one of their own requests — filters out `status: 2` (rejected), because re-surfacing a declined supplier would let the buyer pick the same one again on a request they already passed (`use-cases/supply_offer.go:179-184`). A **matched** offer is deliberately still listed, since the buyer must keep seeing the offer they committed to. Note the path: offers hang off `/supply-offers/requests/{request_id}`, **not** off `/supply-requests/{id}/offers`, which is not a registered route.
- [ ] **`Expire` refuses a request with an active match, exactly like `Cancel`.** `POST /api/v1/supply-requests/{id}/expire` returns `primary.ErrActiveMatch` when `ExistsActiveByRequest` is true (`use-cases/supply_request.go:253-259`), the same sentinel `Cancel` returns at :229-235. `handleSupplyError` maps it to **`409 {"error":"supply request already has an active match"}`** — the conflict list is shared, not per-handler. Both routes also return `409 invalid supply request status` when the request is not open.
- [ ] **Money is `float64` in JSON, and that is not a defect.** Every amount, quantity and score on the supply-chain surface is an IEEE-754 double, not a fixed-point value. Two things consumers should know: `SupplyRequestDTO` carries both `total_amount` and `actual_amount`, where `actual_amount` is the **remaining un-reserved** amount (not a delivery total); and completion writes `actual_amount = 0` explicitly rather than leaving the residue of subtracting fractions one at a time, so a completed request reads an exact `0`.
- [ ] **Availability is read-only over the API.** `SupplierInventoryRepositoryImpl` has full `Create`/`Update`/`Delete` (`repository/supplier_inventory.go:37,120,138`) but **no route is registered for any of them** — the only inventory surface is `GET /api/v1/recommendations/availability`. Availability figures therefore reflect whatever was seeded out of band, and there is no way to correct them from the API.
- [ ] **`repository.ErrAmountConstraint` is unmapped.** The `ck_supply_requests_amounts` CHECK violation is a deliberate, diagnosable sentinel (`repository/supply_request.go:25`), but it appears in no handler conflict list and in no `httpx.StatusCode` case, so if it ever reaches the boundary it is logged and answered as `500 internal server error`. The same holds for `ErrInsufficientAmount` on the transaction routes, whose `409` set omits it (`handler/transaction.go:107-118`).
- [ ] **Trailing slashes matter** for `POST /api/v1/offerings/create2/` (registered with a trailing slash). The collection routes (`/companies/`, `/offerings/`, `/reviews/`, `/inquiries/`, `/liquidations/`, `/reports/`, `/conversations/`, `/messages/`, `/supply-requests/`, `/supply-offers/`) also answer without the trailing slash thanks to chi's mount behavior.
- [ ] **Ordering matters** where static and parameterized paths coexist: `GET /liquidations/open` wins over `GET /liquidations/{id}`, `GET /inquiries/company/{company_id}` wins over `GET /inquiries/{id}`, and `GET /supply-requests/available` wins over `GET /supply-requests/{id}` (chi matches static segments first). The supply-offer and match groups avoid the problem by arity: `/supply-offers/requests/{request_id}` and `/matches/requests/{requestID}` are two segments where `/{id}` is one.
- [ ] **Images:** the stored/read filename comes from the multipart upload's original name and is joined onto `./uploads`; treat `GET /images/{filename}` as serving only single-segment names.

### Server facts

| Item | Value |
|---|---|
| Port | `SERVER_PORT`, default `8080` (`server/cmd/api/main.go:81-84`) |
| Required env | `JWT_SECRET` (fatal if missing), Postgres (`POSTGRES_*`, `DB_SSLMODE`), Redis (`REDIS_URL`/`REDIS_HOST`+`REDIS_PORT`), Elasticsearch (`ESCLIENT_*`) |
| Timeouts | read 10 s, write 15 s, idle 60 s (`main.go:183-185`) |
| Migrations | run automatically at startup and **embedded in the binary** — `//go:embed migrations/*.sql` (`repository/migrations.go:19`) read through the `iofs` source driver (`database/migrate.go`, `NewMigrationSource`). No working directory and no on-disk migration path is involved, so the server boots from any directory. |
| Image store | local directory `./uploads` (`main.go:148`) |
| Transactions | registered outside `NewRouter` by `api.RegisterTransactionRoutes(r, transactionHandler, authMW, suspensionMW)` (`main.go:178`) |
