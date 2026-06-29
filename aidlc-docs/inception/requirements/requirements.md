# Requirements — Import Revert Feature + Bug Fix

## Intent Analysis

| Field | Value |
|---|---|
| **User Request** | Bug report: deleting an import history record does not revert the transactions. Feature request: ability to delete/revert all data belonging to a specific import file. |
| **Request Type** | Bug Fix + Enhancement |
| **Scope** | Multiple Components (repository, service, handler, frontend) |
| **Complexity** | Moderate |

---

## Bug Confirmation

### BUG-001 (Critical): Delete import batch silently orphans transactions

`DELETE /api/import/history/:id` currently calls only `batchRepo.Delete()`, which issues:
```sql
DELETE FROM import_batch WHERE import_batch_id = ?
```
The FK constraint `ON DELETE SET NULL` then fires, setting `import_batch_id = NULL` on all associated transactions — but the **transactions are never deleted**. A user who imports a file and tries to undo it via the Import History UI loses the audit trail but retains all data. No API or UI path currently exists to delete transactions by import batch.

---

## Functional Requirements

### FR1 — Transaction bulk delete by batch ID
Add `TransactionRepository.DeleteByBatchID(batchID int) (int, error)` that deletes all `ledger_transaction` rows where `import_batch_id = ?` and returns the count of deleted rows.

### FR2 — Manually-categorized count on import batch
Extend `ImportBatch` model with a computed field `ManuallyCategCount int`. Extend `GetByID` and `GetAll` queries in `ImportBatchRepository` to include a subquery:
```sql
(SELECT COUNT(*) FROM ledger_transaction
 WHERE import_batch_id = ib.import_batch_id AND category_source = 2) AS manually_categ_count
```

### FR3 — Import revert service method
Add `ImportService.RevertImport(batchID int) (*RevertResult, error)` that:
1. Verifies the batch exists; returns error if not found
2. Calls `txnRepo.DeleteByBatchID(batchID)` and captures deleted count
3. Calls `batchRepo.Delete(batchID)`
4. Returns `RevertResult{BatchID, FileName, DeletedTransactions}`

Add `RevertResult` struct to `import_service.go`.

### FR4 — DELETE endpoint revert behavior (Q2: A)
Modify `ImportBatchHandler.DeleteBatch()` to call `importService.RevertImport()` instead of `batchRepo.Delete()` directly. Response body changes to include `deleted_transactions` count.

`ImportBatchHandler` constructor updated to accept `*ImportService` as a dependency.

### FR5 — Dependency wiring in main.go
Update `NewImportBatchHandler(...)` call in `main.go` to pass `importService` as the second argument.

### FR6 — Import History UI: Revert button with confirmation (Q3: B, Q4: A)
In `import.html`, Import History tab:
- Add a Revert button (trash icon) to each row's Actions column alongside the existing View button
- On click: show a confirmation modal containing:
  - Batch file name
  - Total imported transactions count (from `imported_transactions`)
  - Warning if `manually_categ_count > 0` (e.g., "⚠️ X transactions were manually categorized")
  - Statement: "This will permanently delete all [N] transactions from this import. This cannot be undone."
  - Cancel / Confirm buttons
- On confirm: call `DELETE /api/import/history/:id`, refresh the history list

### FR7 — Transactions page: Revert banner when filtered by batch (Q3: B)
In `transactions.html`, when the page URL contains `?batch_id=X`:
- Client-side JS fetches `GET /api/import/history/:id` to retrieve batch details
- Show a dismissible banner above the transaction list:
  - "Viewing import: **filename.ofx**  [Revert this import ↩]"
  - Clicking "Revert this import" triggers the same confirmation modal flow as FR6

### FR8 — Fix missing batch_id in page handler filter values
In `page_handler.go`, `Transactions()` handler: add `"BatchID": c.Query("batch_id")` to `filterValues`. Currently the batch_id filter is applied to the DB query but not surfaced to the template, making the revert banner impossible to render server-side.

---

## Non-Functional Requirements

### NFR1 — Atomicity
The revert operation (delete transactions + delete batch) should be treated as an atomic logical operation at the service level. If `DeleteByBatchID` fails, `batchRepo.Delete` should not be called (existing error-return pattern is sufficient; no DB transaction wrapper required since SQLite is single-writer).

### NFR2 — Irreversibility warning
The UI must make clear that a revert cannot be undone (Q4: A). The confirmation modal must include explicit wording.

### NFR3 — Redundant stats update (Q5: B — leave as-is)
The frontend's redundant `PUT /api/import/history/:id` call after import is acknowledged and intentionally left unchanged.

---

## Out of Scope

- Database schema FK change (`ON DELETE SET NULL` → `ON DELETE CASCADE`) — the service-level deletion makes this unnecessary
- Partial revert (deleting only some transactions from a batch)
- Security, property-based testing, and resiliency extensions — all opted out
