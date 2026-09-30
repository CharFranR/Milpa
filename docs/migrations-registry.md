# Migration registry

This file fixes the migration numbers **before** anyone writes SQL.

Two workstreams are adding migrations at the same time, and neither one can see
the other's branch. Left to their own devices they will both reach for `000024`,
one of them will win the merge, and the other will silently rewrite a number that
is already applied in somebody's database. This document is the arbitration: the
numbers below are assigned, and a workstream that wants one takes the one it was
given.

The highest migration currently on disk is `000023`
(`000023_supplier_inventory_quantity_check`). Everything from `000024` up is
allocated here.

## Registry

| # | Name | Closes | Depends on |
|---|---|---|---|
| `000024` | `units_of_measure` (+ seed) | RF-06, RF-17 | — |
| `000025` | `categories_catalogue` (+ seed) | RF-06 | `000024` |
| `000026` | `offerings_catalogue` | RF-05, RF-09 | `000024`, `000025` |
| `000027` | `supply_offers_price_and_comments` | RF-11 | — |
| `000028` | `reviews_targets` | RF-15 | — |
| `000029` | `offerings_duplicate_prevention` | RF-05 | — |
| `000030` | `transactions_cancel_reasons` | RF-12 | — |
| `000031` | `liquidations_buyer_visibility` (+ `closed_at` backfill) | RF-14 | `users` exists (000003) |
| `000032` | `liquidation_interests` (+ `assigned_buyer_id`) | RF-14 | `000031` |
| `000033` | `report_target_types` | RF-16 | — |
| `000034` | `users_registration_invariants` | RF-01 | — |
| `000035` | `companies_agricultural` | RF-04 | `companies` exists (000004) |
| `000036` | `conversations_match_link` | RF-13 | `matches` exists (000018) |
| `000037`+ | **RESERVED BLOCK — roles/RBAC stream** | RF-01, RF-03, RF-08, RF-14, RF-17 | — |

## Rules

### A foreign key may only point backwards

No migration may declare a foreign key to a table that does not exist at a
**lower** version number. A migration runner applies versions in ascending order,
so a reference to a table that only appears later in the sequence cannot resolve
at the moment it is needed.

This is why `000024` (units) and `000025` (categories) come **before** `000026`
(offerings), even though offerings has the lower RF number. `offerings_catalogue`
classifies a product by unit and by category, so both of those tables have to
exist first. RF numbering expresses priority; migration numbering expresses
dependency order, and only the second one is constrained.

### Every migration ships with its `.down.sql`

Every `.sql` migration is committed together with its matching `.down.sql`, in
the same commit. A migration without its down leaves the rollback path broken,
and there is no later commit in which to add it.

This is not a convention the suite enforces directly, but it is enforced
indirectly and completely: `tests/integration/migration_source_test.go`
(`TestEmbeddedMigrationsCoverEveryFileOnDisk`) fails if a `.sql` file exists on
disk but is missing from the embedded filesystem, and the embed pattern is
`//go:embed migrations/*.sql`. An orphan `.sql` is a failing test, not a review
comment.

### The blocks do not overlap

- The workstream handling **RF-01 … RF-14** owns `000024`–`000036`.
- The workstream handling the **roles / RBAC** migration owns `000037` and above.
- Neither may use the other's block. If a workstream needs a migration outside
  its range, that is a change to this document, agreed first — not a number
  picked at the keyboard.

`000037`+ is a block and not a single migration because the RBAC stream is
expected to need more than one: roles, permissions, and the user-to-role
assignment are three separate changes and they will not land in one commit.

### The embed pattern stays as it is

Migrations are embedded at boot via `//go:embed migrations/*.sql` in
`infrastructure/adapters/secondary/repository/migrations.go`. Keep that pattern.

**No subdirectories are allowed in the migrations directory.** The pattern
matches one level only, so a migration placed in a subdirectory would be
silently absent from the binary while still sitting on disk — booting would then
stop short of the version the repository contains, and
`TestEmbeddedMigrationsCoverEveryFileOnDisk` would fail. If a set of migrations
genuinely needs grouping, the grouping goes in the **file name**, not in the
directory.
