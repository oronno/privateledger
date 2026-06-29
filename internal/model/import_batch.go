package model

import "time"

// ImportBatch represents a batch of imported transactions
type ImportBatch struct {
	ImportBatchID         int       `json:"import_batch_id"`
	FileName              string    `json:"file_name"`
	AccountID             int       `json:"account_id"`
	CreatedAt             time.Time `json:"created_at"`
	ImportedTransactions  *int      `json:"imported_transactions"`
	DuplicateTransactions *int      `json:"duplicate_transactions"`
	TotalAutoCategorized  *int      `json:"total_auto_categorized"`

	// Joined fields
	AccountName        string `json:"account_name,omitempty"`
	ManuallyCategCount int    `json:"manually_categ_count"`
}
