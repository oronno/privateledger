package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/oronno/privateledger/internal/model"
	"github.com/oronno/privateledger/internal/repository"
	"github.com/oronno/privateledger/internal/service"
)

// ImportBatchHandler handles import batch-related HTTP requests
type ImportBatchHandler struct {
	batchRepo     *repository.ImportBatchRepository
	importService *service.ImportService
}

// NewImportBatchHandler creates a new ImportBatchHandler
func NewImportBatchHandler(batchRepo *repository.ImportBatchRepository, importService *service.ImportService) *ImportBatchHandler {
	return &ImportBatchHandler{
		batchRepo:     batchRepo,
		importService: importService,
	}
}

// CreateBatch creates a new import batch record
// POST /api/import/history
// Content-Type: application/json
// Body: {"file_name": "example.ofx", "account_id": 1}
func (h *ImportBatchHandler) CreateBatch(c *gin.Context) {
	var req struct {
		FileName  string `json:"file_name" binding:"required"`
		AccountID int    `json:"account_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	batch := &model.ImportBatch{
		FileName:  req.FileName,
		AccountID: req.AccountID,
	}

	if err := h.batchRepo.Create(batch); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create import batch"})
		return
	}

	c.JSON(http.StatusCreated, batch)
}

// UpdateBatch updates an import batch with results
// PUT /api/import/history/:id
// Content-Type: application/json
// Body: {"imported_transactions": 10, "duplicate_transactions": 2, "total_auto_categorized": 5}
func (h *ImportBatchHandler) UpdateBatch(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid batch ID"})
		return
	}

	var req struct {
		ImportedTransactions  int `json:"imported_transactions"`
		DuplicateTransactions int `json:"duplicate_transactions"`
		TotalAutoCategorized  int `json:"total_auto_categorized"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	batch, err := h.batchRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get import batch"})
		return
	}
	if batch == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Import batch not found"})
		return
	}

	batch.ImportedTransactions = &req.ImportedTransactions
	batch.DuplicateTransactions = &req.DuplicateTransactions
	batch.TotalAutoCategorized = &req.TotalAutoCategorized

	if err := h.batchRepo.Update(batch); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update import batch"})
		return
	}

	c.JSON(http.StatusOK, batch)
}

// ListBatches returns all import batches
// GET /api/import/history
func (h *ImportBatchHandler) ListBatches(c *gin.Context) {
	batches, err := h.batchRepo.GetAll()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get import batches"})
		return
	}

	c.JSON(http.StatusOK, batches)
}

// GetBatch returns a single import batch by ID
// GET /api/import/history/:id
func (h *ImportBatchHandler) GetBatch(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid batch ID"})
		return
	}

	batch, err := h.batchRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get import batch"})
		return
	}
	if batch == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Import batch not found"})
		return
	}

	c.JSON(http.StatusOK, batch)
}

// DeleteBatch reverts an import batch by deleting all its transactions and then the batch record.
// DELETE /api/import/history/:id
func (h *ImportBatchHandler) DeleteBatch(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid batch ID"})
		return
	}

	result, err := h.importService.RevertImport(id)
	if err != nil {
		if err.Error() == fmt.Sprintf("import batch not found: %d", id) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Import batch not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revert import batch"})
		return
	}

	c.JSON(http.StatusOK, result)
}
