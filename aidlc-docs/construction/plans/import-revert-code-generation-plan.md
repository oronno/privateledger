# Code Generation Plan — import-revert

**Unit**: import-revert  
**Date**: 2026-06-28  
**Workspace Root**: /Users/tanzil/Documents/GitHub/privateledger

## Unit Context

**Goal**: Fix BUG-001 (DELETE batch orphans transactions) and add import revert capability across all layers.

**Dependencies**: None (single self-contained unit)

**Files to modify** (in dependency order — bottom-up through the layers):

| Step | File | Type | Change |
|---|---|---|---|
| 1 | `internal/model/import_batch.go` | Modify | Add `ManuallyCategCount int` computed field |
| 2 | `internal/repository/transaction_repo.go` | Modify | Add `DeleteByBatchID()` |
| 3 | `internal/repository/import_batch_repo.go` | Modify | Extend `GetByID` + `GetAll` with manually_categ_count subquery |
| 4 | `internal/service/import_service.go` | Modify | Add `RevertResult` struct + `RevertImport()` method |
| 5 | `internal/handler/import_batch_handler.go` | Modify | Update constructor; replace `DeleteBatch` body with service call |
| 6 | `cmd/privateledger/main.go` | Modify | Pass `importService` to `NewImportBatchHandler` |
| 7 | `internal/handler/page_handler.go` | Modify | Add `BatchID` to filterValues in `Transactions()` |
| 8 | `cmd/privateledger/web/templates/import.html` | Modify | Add Revert button + confirmation modal to Import History tab |
| 9 | `cmd/privateledger/web/templates/transactions.html` | Modify | Add batch revert banner when `?batch_id=X` is active |

---

## Steps

- [x] **Step 1**: Modify `internal/model/import_batch.go`
  - Add `ManuallyCategCount int \`json:"manually_categ_count"\`` to `ImportBatch` struct
  
- [x] **Step 2**: Modify `internal/repository/transaction_repo.go`
  - Add `DeleteByBatchID(batchID int) (int, error)` method
  - Runs `DELETE FROM ledger_transaction WHERE import_batch_id = ?`
  - Returns count of deleted rows via `RowsAffected()`

- [x] **Step 3**: Modify `internal/repository/import_batch_repo.go`
  - Extend `GetByID` SELECT to include subquery for `manually_categ_count`
  - Extend `GetAll` SELECT to include same subquery
  - Add `batch.ManuallyCategCount` to both Scan calls

- [x] **Step 4**: Modify `internal/service/import_service.go`
  - Add `RevertResult` struct: `{BatchID int, FileName string, DeletedTransactions int}`
  - Add `RevertImport(batchID int) (*RevertResult, error)`:
    1. Call `batchRepo.GetByID(batchID)` → error if nil
    2. Call `txnRepo.DeleteByBatchID(batchID)` → capture deleted count
    3. Call `batchRepo.Delete(batchID)`
    4. Return `&RevertResult{...}`

- [x] **Step 5**: Modify `internal/handler/import_batch_handler.go`
  - Add `importService *service.ImportService` field to `ImportBatchHandler`
  - Update `NewImportBatchHandler` to accept `importService *service.ImportService`
  - Replace `DeleteBatch` body: call `h.importService.RevertImport(id)`, return `{message, deleted_transactions, batch_id, file_name}`

- [x] **Step 6**: Modify `cmd/privateledger/main.go`
  - Change `handler.NewImportBatchHandler(importBatchRepo)` → `handler.NewImportBatchHandler(importBatchRepo, importService)`

- [x] **Step 7**: Modify `internal/handler/page_handler.go`
  - In `Transactions()`, add `"BatchID": c.Query("batch_id")` to the `filterValues` map

- [x] **Step 8**: Modify `cmd/privateledger/web/templates/import.html`
  - In the history table row `<td class="text-center">`, add Revert button alongside View button
  - Add `data-testid="revert-batch-btn-${batch.import_batch_id}"` to the button
  - Add Bootstrap modal `#revertConfirmModal` with:
    - Dynamic content: file name, transaction count, manual-categorization warning
    - Cancel + Confirm buttons (`data-testid="revert-confirm-btn"`, `data-testid="revert-cancel-btn"`)
  - Add JS `revertBatch(batchId, fileName, importedCount, manualCount)` function:
    - Populates modal content
    - Shows modal
    - On confirm: calls `DELETE /api/import/history/${batchId}`, refreshes history list

- [x] **Step 9**: Modify `cmd/privateledger/web/templates/transactions.html`
  - Add JS block: on page load, if `batch_id` is in `window.location.search`:
    - Fetch `GET /api/import/history/{batchId}` to get batch details
    - Render a dismissible alert banner above the transaction table:
      `"Viewing import: filename.ofx — [↩ Revert this import]"`
    - "Revert this import" link triggers the same confirmation flow and calls `DELETE /api/import/history/{batchId}`
    - On success: redirect to `/transactions`
  - Add `data-testid="batch-revert-banner"` and `data-testid="batch-revert-link"` to elements
