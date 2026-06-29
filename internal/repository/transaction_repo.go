package repository

import (
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/oronno/privateledger/internal/model"
)

// TransactionRepository handles database operations for transactions
type TransactionRepository struct {
	db *sql.DB
}

// NewTransactionRepository creates a new TransactionRepository
func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// Create inserts a new transaction into the database
func (r *TransactionRepository) Create(txn *model.Transaction) error {
	query := `
		INSERT INTO ledger_transaction (
			account_id, import_batch_id, trn_type, fit_id, date_posted, amount,
			transaction_details, transaction_type, category_id, category_source
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(query,
		txn.AccountID,
		txn.ImportBatchID,
		txn.TrnType,
		txn.FitID,
		formatDatePosted(txn.DatePosted),
		txn.Amount,
		txn.TransactionDetails,
		txn.TransactionType,
		txn.CategoryID,
		txn.CategorySource,
	)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get transaction ID: %w", err)
	}

	txn.TransactionID = int(id)
	return nil
}

// formatDatePosted parse datePosted as string in simplified format without timezone ("2006-01-02 15:04:05") while interacting with DB.
// The modernc.org/sqlite driver would otherwise call time.Time.String() which includes timezone.
// In some cases, OFX-parsed timestamps look like "2025-10-22 12:00:00 -0500 -0500" in sqlite DB, which can cause issues when parsing the value back. Stripping the timezone eliminates these problems.
func formatDatePosted(datePosted time.Time) string {
	return datePosted.Format(time.DateTime)
}

// FindDuplicate checks if a transaction already exists (for deduplication)
func (r *TransactionRepository) FindDuplicate(accountID int, trnType, fitID string, datePosted time.Time) (*model.Transaction, error) {
	query := `
		SELECT transaction_id, account_id, trn_type, fit_id, date_posted, amount,
			transaction_details, transaction_type, category_id, category_source, created_at
		FROM ledger_transaction
		WHERE account_id = ? AND trn_type = ? AND fit_id = ? AND date_posted = ?
	`

	var txn model.Transaction
	err := r.db.QueryRow(query, accountID, trnType, fitID, formatDatePosted(datePosted)).Scan(
		&txn.TransactionID,
		&txn.AccountID,
		&txn.TrnType,
		&txn.FitID,
		&txn.DatePosted,
		&txn.Amount,
		&txn.TransactionDetails,
		&txn.TransactionType,
		&txn.CategoryID,
		&txn.CategorySource,
		&txn.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find duplicate: %w", err)
	}

	return &txn, nil
}

// GetByID retrieves a transaction by its ID
func (r *TransactionRepository) GetByID(transactionID int) (*model.Transaction, error) {
	query := `
		SELECT transaction_id, account_id, trn_type, fit_id, date_posted, amount,
			transaction_details, transaction_type, category_id, category_source, created_at
		FROM ledger_transaction
		WHERE transaction_id = ?
	`

	var txn model.Transaction
	err := r.db.QueryRow(query, transactionID).Scan(
		&txn.TransactionID,
		&txn.AccountID,
		&txn.TrnType,
		&txn.FitID,
		&txn.DatePosted,
		&txn.Amount,
		&txn.TransactionDetails,
		&txn.TransactionType,
		&txn.CategoryID,
		&txn.CategorySource,
		&txn.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return &txn, nil
}

// TransactionFilter holds filter criteria for listing transactions
type TransactionFilter struct {
	AccountID     *int                // Filter by account
	CategoryID    *int                // Filter by category
	CategoryType  *model.CategoryType // Filter by category type (Expense/Income/Investment)
	BatchID       *int                // Filter by import batch
	Uncategorized bool                // Only uncategorized transactions
	StartDate     *time.Time          // Date range start (inclusive)
	EndDate       *time.Time          // Date range end (inclusive)
	Limit         int                 // Max results (0 = no limit)
	Offset        int                 // Pagination offset
}

// List retrieves transactions with optional filters
func (r *TransactionRepository) List(filter TransactionFilter) ([]*model.Transaction, error) {
	slog.Info("fetching transactions in List()", "filter", filter)
	query := `
		SELECT
			t.transaction_id, t.account_id, t.import_batch_id, t.trn_type, t.fit_id, t.date_posted, t.amount,
			t.transaction_details, t.transaction_type, t.category_id, t.category_source, t.created_at,
			a.name as account_name,
			c.name as category_name,
			c.color as category_color,
			c.icon as category_icon
		FROM ledger_transaction t
		LEFT JOIN account a ON t.account_id = a.account_id
		LEFT JOIN category c ON t.category_id = c.category_id
		WHERE 1=1
	`
	var args []interface{}

	// Apply filters
	if filter.AccountID != nil {
		query += " AND t.account_id = ?"
		args = append(args, *filter.AccountID)
	}

	if filter.CategoryID != nil {
		query += " AND t.category_id = ?"
		args = append(args, *filter.CategoryID)
	}

	if filter.CategoryType != nil {
		query += " AND c.category_type = ?"
		args = append(args, *filter.CategoryType)
	}

	if filter.BatchID != nil {
		query += " AND t.import_batch_id = ?"
		args = append(args, *filter.BatchID)
	}

	if filter.Uncategorized {
		query += " AND (t.category_id IS NULL OR t.category_source = 0)"
	}

	if filter.StartDate != nil {
		query += " AND t.date_posted >= ?"
		args = append(args, *filter.StartDate)
	}

	if filter.EndDate != nil {
		query += " AND t.date_posted <= ?"
		args = append(args, *filter.EndDate)
	}

	// Order by date descending (most recent first)
	query += " ORDER BY t.date_posted DESC, t.transaction_id DESC"

	// Apply pagination
	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}

	if filter.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*model.Transaction
	for rows.Next() {
		var txn model.Transaction
		err := rows.Scan(
			&txn.TransactionID,
			&txn.AccountID,
			&txn.ImportBatchID,
			&txn.TrnType,
			&txn.FitID,
			&txn.DatePosted,
			&txn.Amount,
			&txn.TransactionDetails,
			&txn.TransactionType,
			&txn.CategoryID,
			&txn.CategorySource,
			&txn.CreatedAt,
			&txn.AccountName,
			&txn.CategoryName,
			&txn.CategoryColor,
			&txn.CategoryIcon,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, &txn)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transactions: %w", err)
	}

	slog.Info("fetching transactions: returning result", "total-transaction", len(transactions), "filter", filter)

	return transactions, nil
}

// UpdateCategory updates the category assignment for a transaction
func (r *TransactionRepository) UpdateCategory(transactionID int, categoryID *int, source model.CategorySource) error {
	query := `UPDATE ledger_transaction SET category_id = ?, category_source = ? WHERE transaction_id = ?`
	result, err := r.db.Exec(query, categoryID, source, transactionID)
	if err != nil {
		return fmt.Errorf("failed to update transaction category: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("transaction not found")
	}

	return nil
}

// GetUncategorized retrieves all transactions without a category (category_source = 0)
func (r *TransactionRepository) GetUncategorized() ([]*model.Transaction, error) {
	query := `
		SELECT transaction_id, account_id, trn_type, fit_id, date_posted, amount,
			transaction_details, transaction_type, category_id, category_source, created_at
		FROM ledger_transaction
		WHERE category_source = 0
		ORDER BY date_posted DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query uncategorized transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*model.Transaction
	for rows.Next() {
		var txn model.Transaction
		err := rows.Scan(
			&txn.TransactionID,
			&txn.AccountID,
			&txn.TrnType,
			&txn.FitID,
			&txn.DatePosted,
			&txn.Amount,
			&txn.TransactionDetails,
			&txn.TransactionType,
			&txn.CategoryID,
			&txn.CategorySource,
			&txn.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, &txn)
	}

	return transactions, nil
}

// CountUncategorized returns the number of uncategorized transactions
func (r *TransactionRepository) CountUncategorized() (int, error) {
	query := `SELECT COUNT(*) FROM ledger_transaction WHERE category_source = 0`

	var count int
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count uncategorized transactions: %w", err)
	}

	return count, nil
}

// DeleteByBatchID deletes all transactions belonging to a given import batch and returns the count deleted.
func (r *TransactionRepository) DeleteByBatchID(batchID int) (int, error) {
	query := `DELETE FROM ledger_transaction WHERE import_batch_id = ?`
	result, err := r.db.Exec(query, batchID)
	if err != nil {
		return 0, fmt.Errorf("failed to delete transactions for batch %d: %w", batchID, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return int(rows), nil
}

// Delete deletes a transaction by ID
func (r *TransactionRepository) Delete(transactionID int) error {
	query := `DELETE FROM ledger_transaction WHERE transaction_id = ?`
	result, err := r.db.Exec(query, transactionID)
	if err != nil {
		return fmt.Errorf("failed to delete transaction: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("transaction not found")
	}

	return nil
}

// BulkUpdateCategory updates categories for multiple transactions matching a pattern
func (r *TransactionRepository) BulkUpdateCategory(categoryID int, source model.CategorySource, transactionIDs []int) error {
	if len(transactionIDs) == 0 {
		return nil
	}

	// Build placeholders for IN clause
	placeholders := make([]string, len(transactionIDs))
	args := []interface{}{categoryID, source}
	for i, id := range transactionIDs {
		placeholders[i] = "?"
		args = append(args, id)
	}

	query := fmt.Sprintf(
		`UPDATE ledger_transaction SET category_id = ?, category_source = ? WHERE transaction_id IN (%s)`,
		strings.Join(placeholders, ","),
	)

	_, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to bulk update categories: %w", err)
	}

	return nil
}

// GetMostRecentTransactionDate returns the date of the most recent transaction
func (r *TransactionRepository) GetMostRecentTransactionDate() (string, error) {
	query := `SELECT date_posted FROM ledger_transaction ORDER BY date_posted DESC LIMIT 1`

	var datePosted string
	err := r.db.QueryRow(query).Scan(&datePosted)
	if err == sql.ErrNoRows {
		return "", nil // No transactions
	}
	if err != nil {
		return "", fmt.Errorf("failed to get most recent transaction date: %w", err)
	}

	return datePosted, nil
}
