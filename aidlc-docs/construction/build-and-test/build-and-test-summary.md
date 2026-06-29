# Build and Test Instructions — import-revert

## Build

```bash
# Verify compilation (already confirmed clean)
go build ./...

# Build binary
make build

# Run from source (development)
go run ./cmd/privateledger
```

## Unit Tests

```bash
# Run all tests
make test

# Run service-layer tests specifically
go test -v ./internal/service -run TestCategorizer
```

## Manual Integration Test Checklist

### Preconditions
- App running (`make run` or `go run ./cmd/privateledger`)
- At least one account created
- One OFX/QFX file available

### Test 1: Import + Revert via Import History tab
1. Navigate to `/import`
2. Import an OFX file → note the `imported_transactions` count in the success banner
3. Click "Import History" tab
4. Verify the batch row shows the correct file name, imported count, and a revert button (↩ icon)
5. Click the revert button (↩)
6. Verify confirmation modal appears with:
   - Correct file name
   - Correct transaction count
   - Manual-categorization warning (if applicable)
7. Click "Revert Import"
8. Verify the batch row disappears from the history list
9. Navigate to `/transactions` → verify the transactions from that import are gone

### Test 2: Re-import same file after revert
1. After Test 1, re-import the same OFX file
2. Verify `imported_transactions` matches original count (not 0 / all-duplicates)
3. Confirms the revert truly removed the transactions

### Test 3: Revert via Transactions page batch filter
1. Import an OFX file, note the `batch_id` from the Import History tab (or the View link URL)
2. Navigate to `/transactions?batch_id=<id>`
3. Verify the blue info banner appears: "Viewing import: filename.ofx  [↩ Revert this import]"
4. Click "Revert this import"
5. Verify confirmation modal, then confirm
6. Verify redirect to `/transactions` (no batch filter) and transactions are gone

### Test 4: Re-import deduplication still works
1. Import a file
2. Import the **same file** a second time
3. Verify second import shows `imported_count = 0`, `duplicate_count = N`
4. Verify only one batch record exists per import attempt in Import History

### Test 5: Manually categorized warning
1. Import an OFX file
2. Manually categorize one or more transactions (category_source = 2)
3. Navigate to Import History and click Revert on that batch
4. Verify the confirmation modal shows the warning: "X transactions were manually categorized and will also be deleted"
5. Confirm revert and verify all transactions (including manually categorized) are deleted

### Test 6: Delete non-existent batch
```bash
curl -X DELETE http://localhost:8844/api/import/history/99999
# Expected: 404 {"error": "Import batch not found"}
```

### Test 7: API response shape
```bash
# After a successful revert:
curl -X DELETE http://localhost:8844/api/import/history/<id>
# Expected 200:
# {"batch_id": N, "file_name": "example.ofx", "deleted_transactions": N}
```
