# Bug Analysis — Import Revert

## Bug 1 (Critical): Deleting an Import Batch Does NOT Delete Its Transactions

### Location
- `internal/repository/import_batch_repo.go:150` — `Delete()`
- `internal/handler/import_batch_handler.go:135` — `DeleteBatch()`
- `internal/database/schema.sql:22` — FK constraint

### What the Schema Says

```sql
import_batch_id INTEGER REFERENCES import_batch(import_batch_id) ON DELETE SET NULL
```

### What Happens When `DELETE /api/import/history/:id` Is Called

1. `DeleteBatch` handler calls `batchRepo.Delete(id)`
2. This runs `DELETE FROM import_batch WHERE import_batch_id = ?`
3. SQLite FK `ON DELETE SET NULL` fires → sets `import_batch_id = NULL` on all transactions that belonged to this batch
4. The `import_batch` record is gone
5. **The transactions remain in the ledger, now orphaned (no batch linkage)**

### User Impact

A user who imports a file and later wants to "undo" it clicks delete on the import history record. They see it disappear from the history list. But ALL the transactions remain in the ledger with no way to identify which batch they came from. The "delete" silently fails to revert the data.

### Evidence

- `import_batch_repo.go:150-167` — Delete only touches `import_batch` table
- `import_batch_handler.go:135-148` — Handler passes through to repo with no transaction cleanup
- No `DeleteByBatchID` method exists anywhere in `transaction_repo.go`
- Import history UI (`import.html:262`) has only a "View" action button — no revert/delete action

---

## Bug 2 (Minor): Redundant Batch Stats Update from Frontend

### Location
- `cmd/privateledger/web/templates/import.html:152-163`

### What Happens

After a successful import:
1. `import_service.go:124-136` already calls `batchRepo.Update()` internally to record stats
2. The frontend JS then makes a redundant `PUT /api/import/history/:id` call with the same stats

The JS even acknowledges this with a comment: `// optional, service already does this`

### Impact
Low — same values written twice, no data corruption. But it's an unnecessary API call.

---

## Summary Table

| # | Severity | Issue | Location |
|---|----------|-------|----------|
| 1 | Critical | `DELETE /api/import/history/:id` orphans transactions instead of deleting them | `import_batch_handler.go`, `import_batch_repo.go`, schema FK |
| 2 | Minor | Frontend redundantly updates batch stats that the service already updated | `import.html` JS |
