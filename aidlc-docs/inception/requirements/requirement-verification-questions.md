# Requirement Verification Questions

**Project**: PrivateLedger — Import Revert Feature + Bug Fix  
**Stage**: Requirements Analysis  
**Date**: 2026-06-28

---

## Bug Confirmed

Before asking questions, here is what the code analysis found:

### Critical Bug: Deleting an import batch does NOT delete its transactions

When you click "delete" on an import history record (or call `DELETE /api/import/history/:id`), the app only removes the batch record. All associated transactions remain in the ledger — orphaned, with no batch linkage. This is because the database schema uses `ON DELETE SET NULL` on the foreign key:

```sql
import_batch_id INTEGER REFERENCES import_batch(import_batch_id) ON DELETE SET NULL
```

**Impact**: A user who imports a file by mistake and tries to "undo" it by deleting the history record will silently lose the audit trail but keep all the transactions. There is currently no way to remove a set of transactions by import batch.

---

## Clarifying Questions

Please fill in the `[Answer]:` tag for each question.

---

## Question 1: Revert Behavior for Manually Categorized Transactions

When reverting an import, some transactions may have been manually categorized by the user after import (category_source = 2). What should happen to those?

A) Delete them anyway — a revert means remove everything from that import, regardless of manual work done on them

B) Preserve them — keep transactions that have been manually categorized; only delete auto-categorized or uncategorized ones from that batch

C) Ask the user at revert time — show a warning and let the user decide (e.g., "X transactions have been manually categorized — still revert?")

X) Other (please describe after [Answer]: tag below)

[Answer]: C

---

## Question 2: API Behavior — Replace or Add?

Should the existing `DELETE /api/import/history/:id` endpoint behavior change to also delete the associated transactions? Or should a new separate endpoint be added for "revert"?

A) Change the existing endpoint — `DELETE /api/import/history/:id` should now delete both the batch record AND all its transactions (breaking change to current API behavior)

B) Add a new endpoint — keep the existing delete as-is (deletes only the batch record); add a new `POST /api/import/history/:id/revert` endpoint that deletes both transactions and the batch record

C) Add a query parameter — keep existing endpoint but add `?revert=true` to make it also delete transactions

X) Other (please describe after [Answer]: tag below)

[Answer]: A

---

## Question 3: UI — Where Should the Revert Action Appear?

Where in the UI should the user be able to trigger the revert?

A) Import History tab only — add a "Revert" button (with confirmation dialog) in the Import History table row actions

B) Import History tab + Transactions page — add revert on Import History AND allow filtering transactions by batch then bulk-deleting

C) Import History tab only is sufficient for now

X) Other (please describe after [Answer]: tag below)

[Answer]: B

---

## Question 4: Revert Confirmation

Should there be a confirmation step before reverting an import?

A) Yes — show a confirmation modal with a summary ("This will delete X transactions from batch 'filename.ofx'. This cannot be undone. Continue?")

B) No — just do it immediately (suitable since the app is local-only)

X) Other (please describe after [Answer]: tag below)

[Answer]: A

---

## Question 5: Minor Bug — Redundant Stats Update

The import UI JavaScript makes a redundant `PUT /api/import/history/:id` call after each import to update batch stats — but the import service already saves these stats internally. The JS code even has a comment acknowledging this (`// optional, service already does this`).

Should this redundant call be removed as part of this fix?

A) Yes — remove the redundant PUT call from the JS

B) No — leave it as-is to be safe (it's harmless, same values written twice)

X) Other (please describe after [Answer]: tag below)

[Answer]: B

---

## Extension Questions

---

## Question 6: Security Extension

Should security extension rules be enforced for this project?

A) Yes — enforce all SECURITY rules as blocking constraints (recommended for production-grade applications)

B) No — skip all SECURITY rules (suitable for PoCs, prototypes, and experimental projects)

X) Other (please describe after [Answer]: tag below)

[Answer]: Skip for now

---

## Question 7: Property-Based Testing Extension

Should property-based testing (PBT) rules be enforced for this project?

A) Yes — enforce all PBT rules as blocking constraints (recommended for projects with business logic, data transformations, serialization, or stateful components)

B) Partial — enforce PBT rules only for pure functions and serialization round-trips

C) No — skip all PBT rules (suitable for simple CRUD applications or thin integration layers)

X) Other (please describe after [Answer]: tag below)

[Answer]: Skip for now

---

## Question 8: Resiliency Extension

Should the resiliency baseline be applied to this project?

A) Yes — apply the resiliency baseline as directional best practices and design-time guidance

B) No — skip the resiliency baseline (suitable for local-only tools where HA and DR don't apply)

X) Other (please describe after [Answer]: tag below)

[Answer]: Skip for now
