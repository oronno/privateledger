# Execution Plan — Import Revert Feature + Bug Fix

## Detailed Analysis Summary

### Transformation Scope
- **Transformation Type**: Single-feature enhancement within existing component boundaries
- **Primary Changes**: Bug fix in import batch deletion + new revert capability
- **Related Components**: TransactionRepository, ImportBatchRepository, ImportService, ImportBatchHandler, main.go, import.html, transactions.html, page_handler.go

### Change Impact Assessment
| Area | Impact | Description |
|---|---|---|
| User-facing | Yes | New Revert button in Import History; Revert banner on Transactions page |
| Structural | Minimal | ImportBatchHandler gains ImportService dependency |
| Data model | Minimal | ImportBatch gains computed `manually_categ_count` field |
| API | Yes | `DELETE /api/import/history/:id` now also deletes transactions + returns richer response |
| NFR | No | No performance, security, or scalability concerns |

### Risk Assessment
- **Risk Level**: Low
- **Rollback Complexity**: Easy — changes are additive with one behavioral change to an existing endpoint
- **Testing Complexity**: Simple — straightforward CRUD + cascade delete logic

---

## Workflow Visualization

```mermaid
flowchart TD
    Start(["User Request"])

    subgraph INCEPTION["🔵 INCEPTION PHASE"]
        WD["Workspace Detection\nCOMPLETED"]
        RE["Reverse Engineering\nCOMPLETED"]
        RA["Requirements Analysis\nCOMPLETED"]
        US["User Stories\nSKIPPED"]
        WP["Workflow Planning\nIN PROGRESS"]
        AD["Application Design\nSKIPPED"]
        UG["Units Generation\nSKIPPED"]
    end

    subgraph CONSTRUCTION["🟢 CONSTRUCTION PHASE"]
        FD["Functional Design\nSKIPPED"]
        NFRA["NFR Requirements\nSKIPPED"]
        NFRD["NFR Design\nSKIPPED"]
        ID["Infrastructure Design\nSKIPPED"]
        CG["Code Generation\nEXECUTE"]
        BT["Build and Test\nEXECUTE"]
    end

    subgraph OPERATIONS["🟡 OPERATIONS PHASE"]
        OPS["Operations\nPLACEHOLDER"]
    end

    Start --> WD --> RE --> RA --> WP
    WP --> CG --> BT --> End(["Complete"])

    style WD fill:#4CAF50,stroke:#1B5E20,stroke-width:3px,color:#fff
    style RE fill:#4CAF50,stroke:#1B5E20,stroke-width:3px,color:#fff
    style RA fill:#4CAF50,stroke:#1B5E20,stroke-width:3px,color:#fff
    style WP fill:#4CAF50,stroke:#1B5E20,stroke-width:3px,color:#fff
    style CG fill:#FFA726,stroke:#E65100,stroke-width:3px,stroke-dasharray:5 5,color:#000
    style BT fill:#FFA726,stroke:#E65100,stroke-width:3px,stroke-dasharray:5 5,color:#000
    style US fill:#BDBDBD,stroke:#424242,stroke-width:2px,stroke-dasharray:5 5,color:#000
    style AD fill:#BDBDBD,stroke:#424242,stroke-width:2px,stroke-dasharray:5 5,color:#000
    style UG fill:#BDBDBD,stroke:#424242,stroke-width:2px,stroke-dasharray:5 5,color:#000
    style FD fill:#BDBDBD,stroke:#424242,stroke-width:2px,stroke-dasharray:5 5,color:#000
    style NFRA fill:#BDBDBD,stroke:#424242,stroke-width:2px,stroke-dasharray:5 5,color:#000
    style NFRD fill:#BDBDBD,stroke:#424242,stroke-width:2px,stroke-dasharray:5 5,color:#000
    style ID fill:#BDBDBD,stroke:#424242,stroke-width:2px,stroke-dasharray:5 5,color:#000
    style OPS fill:#BDBDBD,stroke:#424242,stroke-width:2px,stroke-dasharray:5 5,color:#000
    style Start fill:#CE93D8,stroke:#6A1B9A,stroke-width:3px,color:#000
    style End fill:#CE93D8,stroke:#6A1B9A,stroke-width:3px,color:#000
    style INCEPTION fill:#BBDEFB,stroke:#1565C0,stroke-width:3px,color:#000
    style CONSTRUCTION fill:#C8E6C9,stroke:#2E7D32,stroke-width:3px,color:#000
    style OPERATIONS fill:#FFF59D,stroke:#F57F17,stroke-width:3px,color:#000
    linkStyle default stroke:#333,stroke-width:2px
```

---

## Phases to Execute

### 🔵 INCEPTION PHASE
- [x] Workspace Detection — COMPLETED
- [x] Reverse Engineering — COMPLETED
- [x] Requirements Analysis — COMPLETED
- [x] Workflow Planning — IN PROGRESS
- [ ] User Stories — SKIP (bug fix + bounded enhancement, no new personas)
- [ ] Application Design — SKIP (no new components; changes within existing boundaries)
- [ ] Units Generation — SKIP (single tightly-coupled unit of work)

### 🟢 CONSTRUCTION PHASE
- [ ] Functional Design — SKIP (business logic is unambiguous from requirements)
- [ ] NFR Requirements — SKIP (all extensions opted out; local app, no NFR concerns)
- [ ] NFR Design — SKIP (per above)
- [ ] Infrastructure Design — SKIP (no infrastructure changes)
- [ ] **Code Generation — EXECUTE** (single unit: 8 files across repository/service/handler/frontend)
- [ ] **Build and Test — EXECUTE**

### 🟡 OPERATIONS PHASE
- [ ] Operations — PLACEHOLDER

---

## Code Generation Plan (Single Unit)

Files to change, in dependency order:

| # | File | Change |
|---|---|---|
| 1 | `internal/model/import_batch.go` | Add `ManuallyCategCount int` field |
| 2 | `internal/repository/transaction_repo.go` | Add `DeleteByBatchID()` |
| 3 | `internal/repository/import_batch_repo.go` | Extend `GetByID` + `GetAll` to include manually_categ_count subquery |
| 4 | `internal/service/import_service.go` | Add `RevertResult` struct + `RevertImport()` method |
| 5 | `internal/handler/import_batch_handler.go` | Update constructor; change `DeleteBatch` to call service |
| 6 | `cmd/privateledger/main.go` | Update `NewImportBatchHandler` call |
| 7 | `internal/handler/page_handler.go` | Add `batch_id` to filterValues in `Transactions()` |
| 8 | `cmd/privateledger/web/templates/import.html` | Add Revert button + confirmation modal |
| 9 | `cmd/privateledger/web/templates/transactions.html` | Add batch revert banner |

---

## Success Criteria
- `DELETE /api/import/history/:id` deletes both the batch record and all its transactions
- Response includes `deleted_transactions` count
- Import History UI shows a Revert button; clicking it shows confirmation modal with count and manual-categorization warning
- Transactions page shows a revert banner when filtered by `?batch_id=X`
- All existing functionality (deduplication, categorization, import flow) unaffected
