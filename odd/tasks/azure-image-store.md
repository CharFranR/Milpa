# Feature: Azure Blob image storage and real uploads from the product form

Status: WU1–WU2 done, all checks green including a live Azure smoke test and an independent
verification pass. Pending: Render env var and push/PR (user decisions).
Owner: el Gentleman (orchestrator) + delegated writer
Scope: `server/` (storage adapter, config, wiring) + `frontend/` (upload flow) + `render.yaml` + `README.md`
Tracker: user report 2026-10-09 ("no se me suben al azure las imagenes"). No GitHub issue.

## Context

Commit `78764e0` pointed the local-disk image store's base dir at an Azure ARM resource ID
(`/subscriptions/.../storageAccounts/milpaimages`). That is not a filesystem path, so
`os.WriteFile` fails (ENOENT) and `POST /api/v1/images/` returns 400; nothing ever reaches Azure.
There is no Azure adapter — the only `port.ImageStore` implementation writes to local disk — and
the frontend never calls the upload endpoint: it compresses to a base64 data URL, stores it in
`localStorage` and embeds it in the offering description (`frontend/src/lib/productImages.js`).
The last pushed attempt on disk (`image.go` WIP, uncommitted) uses `DefaultAzureCredential` against
the ARM resource ID and does not compile.

## Decisions

1. New `AzureImageStoreImpl` on `github.com/Azure/azure-sdk-for-go/sdk/storage/azblob`,
   authenticated with a connection string (`AZURE_STORAGE_CONNECTION_STRING`) — the deployment
   target is Render, which has no Azure managed identity, so `DefaultAzureCredential` cannot work.
2. Container name from `AZURE_STORAGE_CONTAINER` (default `images`); created at startup if missing.
3. `Upload` returns an absolute public URL served by the existing public route
   `GET /api/v1/images/{filename}` (base resolved from `PUBLIC_API_URL` → `RENDER_EXTERNAL_URL` →
   `http://localhost:<port>`). The container stays private; the API proxies the bytes.
4. `LocalImageStoreImpl` remains the dev fallback when the connection string is unset, and returns
   the same absolute URL shape so the frontend behaves identically in both modes.
5. Frontend: the product form uploads the compressed image through `POST /api/v1/images/`
   (multipart field `image`, Bearer) and stores the returned URL as `image_url`. The
   base64/localStorage path stays only as legacy read support.
6. Secrets stay out of the repo: `sync: false` in `render.yaml`, empty in `.env.example`.

## Tasks

- [x] WU1 — Backend: Azure adapter + config + wiring + local adapter URL + env templates.
      Commit `29d2a5c`
- [x] WU2 — Frontend: FormData support in the HTTP helper + images service + form upload flow.
      Commit `a070eb6`
- [x] Follow-up from independent verification: stale-upload guard in the form (monotonic token;
      removing or cancelling a form no longer lets an in-flight upload re-attach its image).
      Commit `210ce0b`

## Acceptance criteria

- With `AZURE_STORAGE_CONNECTION_STRING` set, `POST /api/v1/images/` stores the blob in Azure and
  `GET /api/v1/images/{filename}` serves it back with its content type.
- Without it, the local fallback behaves as before but returns an absolute URL.
- Selecting an image in the product form uploads it and the saved offering renders it.
- `go build`, `go vet` pass; frontend `build`, `lint`, `test` pass.

## Checks

- `cd server && go build ./... && go vet ./...`
- `cd server && go test ./...` (report Testcontainers environment failures honestly)
- `cd frontend && npm run build && npm run lint && npm test`

## Route declaration

| Task | Route | Trigger evidence |
| --- | --- | --- |
| WU1–WU2 | delegated direct (one writer) | 2+ non-trivial files across two stacks → writer trigger |

## Delivery strategy

`ask-on-risk` (default). Forecast: ~250 authored lines, below the ~400 budget; single branch
`fix/azure-image-store` (from `develop`), one PR expected. Not pushed; push/PR is the user's call.

## Progress

- 2026-10-09: branch created from `develop`; diagnosis recorded (see Context). WU1–WU2 delegated
  to one writer.
- Writer evidence: `gofmt` clean on touched dirs; `go build ./...` + `go vet ./...` exit 0;
  storage package tests ok; full `go test ./...` green (integration suite included, 45s);
  frontend `build` + `lint` clean, vitest 19 files / 104 tests green. Azurite smoke test
  (upload → download → delete) passed.
- Comment cleanup on `azure_image.go` (user edit, kept): committed separately.
- Orchestrator live smoke against the real account `milpaimages` (connection string in
  `server/.env`, gitignored — never committed): client from connection string, container
  `images` created, upload with content type, download roundtrip, delete, `BlobNotFound`
  after delete → ALL PASSED. Temporary smoke module deleted.
- Independent verification (fresh read-only verifier over `78764e0..0737bb5`): VERIFIED, no
  blockers. Re-ran storage tests (7/7), build/vet, frontend lint + 104/104 tests; confirmed
  secret hygiene (no `AccountKey=` anywhere in the diff, `.env` untracked) and the config
  resolution order. One SUGGESTION (stale-upload race) was applied afterwards as `210ce0b`
  and validated with frontend build/lint + 104/104 tests.
- Native assess (RDD off for this clone): risk `medium` (executable change; 514 changed lines,
  slice budget reached). Review lifecycle NOT started — the switch is disabled, delivery follows
  ordinary repository policy; no receipt, no fabricated approval.
- Orchestrator end-to-end against the rebuilt local stack (`image store: azure blob storage` in the
  logs): registered a throwaway user, `POST /api/v1/images/` with `fondo-campo.jpeg` → returned
  `http://localhost:8080/api/v1/images/img-0f51c46a-...jpeg`; `GET` roundtrip byte-identical
  (sha256 `2e1d4992…`), `Content-Type: image/jpeg`; the `/app/uploads` volume stayed empty → the
  bytes live in Azure, not on disk. Test blob deleted from the account afterwards. Test user
  `imagetest-1791533586@example.com` remains in the local dev database.
- Pending: paste `AZURE_STORAGE_CONNECTION_STRING` into the Render dashboard (`sync: false`);
  push branch and open the PR to `develop` (user decisions). Offerings published before this
  fix keep their legacy base64 `ImageBase64:` values until re-uploaded.
