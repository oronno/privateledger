package service

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/oronno/privateledger/internal/model"
	"github.com/oronno/privateledger/internal/parser"
	"github.com/oronno/privateledger/internal/repository"
)

// ImportService handles OFX file imports with deduplication and categorization
type ImportService struct {
	parser      *parser.OFXParser
	txnRepo     *repository.TransactionRepository
	accountRepo *repository.AccountRepository
	categorizer *Categorizer
	batchRepo   *repository.ImportBatchRepository
}

// NewImportService creates a new ImportService
func NewImportService(
	parser *parser.OFXParser,
	txnRepo *repository.TransactionRepository,
	accountRepo *repository.AccountRepository,
	categorizer *Categorizer,
	batchRepo *repository.ImportBatchRepository,
) *ImportService {
	return &ImportService{
		parser:      parser,
		txnRepo:     txnRepo,
		accountRepo: accountRepo,
		categorizer: categorizer,
		batchRepo:   batchRepo,
	}
}

// ImportResult contains the results of an OFX import operation
type ImportResult struct {
	TotalTransactions int                  `json:"total_transactions"`
	ImportedCount     int                  `json:"imported_count"`
	DuplicateCount    int                  `json:"duplicate_count"`
	CategorizedCount  int                  `json:"categorized_count"`
	ErrorCount        int                  `json:"error_count"`
	Errors            []string             `json:"errors,omitempty"`
	Transactions      []*model.Transaction `json:"transactions,omitempty"`
}

// ImportOFX imports transactions from an OFX file
func (s *ImportService) ImportOFX(reader io.Reader, accountID int, batchID *int) (*ImportResult, error) {
	// Verify account exists
	account, err := s.accountRepo.GetByID(accountID)
	if err != nil {
		slog.Error("Error getting account", slog.Int("account_id", accountID), slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to get account: %w", err)
	}
	if account == nil {
		slog.Warn("Account not found", slog.Int("account_id", accountID))
		return nil, fmt.Errorf("account not found: %d", accountID)
	}

	// Parse OFX file
	parseResult, err := s.parser.ParseOFXFile(reader, accountID)
	if err != nil {
		slog.Warn("Error Parsing OFX file", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to parse OFX file: %w", err)
	}

	result := &ImportResult{
		TotalTransactions: len(parseResult.Transactions),
		Errors:            make([]string, 0),
		Transactions:      make([]*model.Transaction, 0),
	}

	// Process each transaction
	for _, txn := range parseResult.Transactions {
		// Set batch ID if provided
		if batchID != nil {
			txn.ImportBatchID = batchID
		}

		// Check for duplicates
		duplicate, err := s.txnRepo.FindDuplicate(
			txn.AccountID,
			txn.TrnType,
			txn.FitID,
			txn.DatePosted,
		)
		if err != nil {
			slog.Error("Error in FindDuplicate", slog.String("error", err.Error()))
			result.ErrorCount++
			result.Errors = append(result.Errors,
				fmt.Sprintf("Error checking duplicate for %s: %v", txn.FitID, err))
			continue
		}

		if duplicate != nil {
			// Transaction already exists
			result.DuplicateCount++
			continue
		}

		// Apply categorization before inserting
		if s.categorizer.Categorize(txn) {
			result.CategorizedCount++
		}

		// Insert transaction
		err = s.txnRepo.Create(txn)
		if err != nil {
			slog.Error("Error in Inserting transaction", slog.String("error", err.Error()))
			result.ErrorCount++
			result.Errors = append(result.Errors,
				fmt.Sprintf("Error inserting transaction %s: %v", txn.FitID, err))
			continue
		}

		result.ImportedCount++
		result.Transactions = append(result.Transactions, txn)
	}

	// Update batch record if batch ID was provided
	if batchID != nil && s.batchRepo != nil {
		batch, err := s.batchRepo.GetByID(*batchID)
		if err != nil {
			slog.Error("Error getting batch record", slog.Int("batch_id", *batchID), slog.String("error", err.Error()))
		} else if batch != nil {
			batch.ImportedTransactions = &result.ImportedCount
			batch.DuplicateTransactions = &result.DuplicateCount
			batch.TotalAutoCategorized = &result.CategorizedCount
			err = s.batchRepo.Update(batch)
			if err != nil {
				slog.Error("Error updating batch record", slog.Int("batch_id", *batchID), slog.String("error", err.Error()))
			}
		}
	}

	return result, nil
}

// ValidateOFX validates an OFX file without importing
func (s *ImportService) ValidateOFX(reader io.Reader) error {
	return s.parser.ValidateOFXFile(reader)
}

// RevertResult contains the results of an import revert operation
type RevertResult struct {
	BatchID             int    `json:"batch_id"`
	FileName            string `json:"file_name"`
	DeletedTransactions int    `json:"deleted_transactions"`
}

// RevertImport deletes all transactions belonging to a batch and then deletes the batch record.
func (s *ImportService) RevertImport(batchID int) (*RevertResult, error) {
	batch, err := s.batchRepo.GetByID(batchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get import batch: %w", err)
	}
	if batch == nil {
		return nil, fmt.Errorf("import batch not found: %d", batchID)
	}

	deleted, err := s.txnRepo.DeleteByBatchID(batchID)
	if err != nil {
		slog.Error("Failed to delete transactions for batch", slog.Int("batch_id", batchID), slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to delete transactions: %w", err)
	}

	if err := s.batchRepo.Delete(batchID); err != nil {
		slog.Error("Failed to delete batch record", slog.Int("batch_id", batchID), slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to delete import batch: %w", err)
	}

	slog.Info("Import reverted", slog.Int("batch_id", batchID), slog.String("file_name", batch.FileName), slog.Int("deleted_transactions", deleted))

	return &RevertResult{
		BatchID:             batchID,
		FileName:            batch.FileName,
		DeletedTransactions: deleted,
	}, nil
}
