# Architecture — Reverse Engineering

## System Overview

PrivateLedger is a local-only personal finance app: a single Go binary serving an HTTP API + HTMX/Bootstrap UI backed by an embedded SQLite database.

## Layered Architecture

```
Browser (Bootstrap 5 + HTMX + Vanilla JS)
        |
        v
HTTP Router (Gin)
        |
   +----+----+
   | Handlers |  ← import_handler.go, import_batch_handler.go, transaction_handler.go, ...
   +----+----+
        |
   +----+----+
   | Services |  ← import_service.go, categorizer.go, insights_service.go
   +----+----+
        |
   +----+------+
   | Repository |  ← transaction_repo.go, import_batch_repo.go, account_repo.go, ...
   +----+------+
        |
     SQLite
```

## Import Flow (Critical Path)

```
Browser                  Handler              Service                Repository
   |                        |                    |                       |
   |-- POST /api/import/history (Step 1) ------->|                       |
   |                        |-- Create batch ---->|-- INSERT import_batch->|
   |<-- {import_batch_id} --|                    |                       |
   |                        |                    |                       |
   |-- POST /api/import (Step 2) + batch_id ---->|                       |
   |                        |-- ImportOFX() ----->|                       |
   |                        |                    |-- FindDuplicate() ---->|
   |                        |                    |-- Create(txn) -------->|
   |                        |                    |-- batchRepo.Update() ->|
   |<-- {result} -----------|                    |                       |
   |                        |                    |                       |
   |-- PUT /api/import/history/:id (Step 3) --->|                       |
   |   [REDUNDANT: service already did this]     |                       |
```

## Schema (Relevant Tables)

```sql
import_batch
  import_batch_id PK
  file_name
  account_id FK → account(CASCADE)
  created_at
  imported_transactions
  duplicate_transactions
  total_auto_categorized

ledger_transaction
  transaction_id PK
  account_id FK → account(CASCADE)
  import_batch_id FK → import_batch(SET NULL)   ← KEY FK
  trn_type, fit_id, date_posted                 ← dedup composite key
  amount, transaction_details, transaction_type
  category_id FK → category(SET NULL)
  category_source (0=none, 1=rule, 2=manual)
  UNIQUE(account_id, trn_type, fit_id, date_posted)
```

## Deduplication Mechanism

- `TransactionRepository.FindDuplicate()` checks the composite key before every insert
- If duplicate found → skip insert, increment `DuplicateCount`
- Guarantees: re-importing same file is safe — no duplicate rows created
