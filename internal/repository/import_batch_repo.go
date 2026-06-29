package repository

import (
	"database/sql"
	"fmt"

	"github.com/oronno/privateledger/internal/model"
)

// ImportBatchRepository handles database operations for import batches
type ImportBatchRepository struct {
	db *sql.DB
}

// NewImportBatchRepository creates a new ImportBatchRepository
func NewImportBatchRepository(db *sql.DB) *ImportBatchRepository {
	return &ImportBatchRepository{db: db}
}

// Create inserts a new import batch into the database
func (r *ImportBatchRepository) Create(batch *model.ImportBatch) error {
	query := `
		INSERT INTO import_batch (file_name, account_id)
		VALUES (?, ?)
	`

	result, err := r.db.Exec(query, batch.FileName, batch.AccountID)
	if err != nil {
		return fmt.Errorf("failed to create import batch: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get import batch ID: %w", err)
	}

	batch.ImportBatchID = int(id)
	return nil
}

// Update updates the import batch statistics
func (r *ImportBatchRepository) Update(batch *model.ImportBatch) error {
	query := `
		UPDATE import_batch
		SET imported_transactions = ?, duplicate_transactions = ?, total_auto_categorized = ?
		WHERE import_batch_id = ?
	`

	result, err := r.db.Exec(query,
		batch.ImportedTransactions,
		batch.DuplicateTransactions,
		batch.TotalAutoCategorized,
		batch.ImportBatchID,
	)
	if err != nil {
		return fmt.Errorf("failed to update import batch: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("import batch not found")
	}

	return nil
}

// GetByID retrieves an import batch by its ID
func (r *ImportBatchRepository) GetByID(batchID int) (*model.ImportBatch, error) {
	query := `
		SELECT
			ib.import_batch_id, ib.file_name, ib.account_id, ib.created_at,
			ib.imported_transactions, ib.duplicate_transactions, ib.total_auto_categorized,
			a.name as account_name,
			(SELECT COUNT(*) FROM ledger_transaction
			 WHERE import_batch_id = ib.import_batch_id AND category_source = 2) AS manually_categ_count
		FROM import_batch ib
		LEFT JOIN account a ON ib.account_id = a.account_id
		WHERE ib.import_batch_id = ?
	`

	var batch model.ImportBatch
	err := r.db.QueryRow(query, batchID).Scan(
		&batch.ImportBatchID,
		&batch.FileName,
		&batch.AccountID,
		&batch.CreatedAt,
		&batch.ImportedTransactions,
		&batch.DuplicateTransactions,
		&batch.TotalAutoCategorized,
		&batch.AccountName,
		&batch.ManuallyCategCount,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get import batch: %w", err)
	}

	return &batch, nil
}

// GetAll retrieves all import batches
func (r *ImportBatchRepository) GetAll() ([]*model.ImportBatch, error) {
	query := `
		SELECT
			ib.import_batch_id, ib.file_name, ib.account_id, ib.created_at,
			ib.imported_transactions, ib.duplicate_transactions, ib.total_auto_categorized,
			a.name as account_name,
			(SELECT COUNT(*) FROM ledger_transaction
			 WHERE import_batch_id = ib.import_batch_id AND category_source = 2) AS manually_categ_count
		FROM import_batch ib
		LEFT JOIN account a ON ib.account_id = a.account_id
		ORDER BY ib.created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query import batches: %w", err)
	}
	defer rows.Close()

	batches := []*model.ImportBatch{}
	for rows.Next() {
		var batch model.ImportBatch
		err := rows.Scan(
			&batch.ImportBatchID,
			&batch.FileName,
			&batch.AccountID,
			&batch.CreatedAt,
			&batch.ImportedTransactions,
			&batch.DuplicateTransactions,
			&batch.TotalAutoCategorized,
			&batch.AccountName,
			&batch.ManuallyCategCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan import batch: %w", err)
		}
		batches = append(batches, &batch)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating import batches: %w", err)
	}

	return batches, nil
}

// Delete deletes an import batch by ID
func (r *ImportBatchRepository) Delete(batchID int) error {
	query := `DELETE FROM import_batch WHERE import_batch_id = ?`
	result, err := r.db.Exec(query, batchID)
	if err != nil {
		return fmt.Errorf("failed to delete import batch: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("import batch not found")
	}

	return nil
}
