package handler

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oronno/privateledger/internal/model"
	"github.com/oronno/privateledger/internal/repository"
	"github.com/oronno/privateledger/internal/service"
)

// PageHandler handles serving HTML pages
type PageHandler struct {
	files           embed.FS
	accountRepo     *repository.AccountRepository
	transactionRepo *repository.TransactionRepository
	categoryRepo    *repository.CategoryRepository
	patternRepo     *repository.CategoryPatternRepository
	insightsService *service.InsightsService
	version         string
}

// NewPageHandler creates a new PageHandler
func NewPageHandler(
	files embed.FS,
	accountRepo *repository.AccountRepository,
	transactionRepo *repository.TransactionRepository,
	categoryRepo *repository.CategoryRepository,
	patternRepo *repository.CategoryPatternRepository,
	insightsService *service.InsightsService,
	version string,
) *PageHandler {
	return &PageHandler{
		files:           files,
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
		categoryRepo:    categoryRepo,
		patternRepo:     patternRepo,
		insightsService: insightsService,
		version:         version,
	}
}

// parseTemplate parses layout.html with a specific page template
func (h *PageHandler) parseTemplate(pageName string) *template.Template {
	// Define custom template functions
	funcMap := template.FuncMap{
		"formatDate": func(date interface{}) string {
			var t time.Time
			switch v := date.(type) {
			case string:
				// Parse date string in format "2006-01-02 15:04:05"
				var err error
				t, err = time.Parse("2006-01-02 15:04:05", v)
				if err != nil {
					// If parsing fails, try date-only format
					t, err = time.Parse("2006-01-02", v)
					if err != nil {
						return v // Return as-is if parsing fails
					}
				}
			case time.Time:
				t = v
			default:
				return fmt.Sprintf("%v", date)
			}
			// Format as "Jan '26"
			return t.Format("Jan '06")
		},
		"abs": func(x float64) float64 {
			if x < 0 {
				return -x
			}
			return x
		},
		"formatDateForURL": func(date time.Time) string {
			return date.Format("2006-01-02")
		},
	}

	return template.Must(template.New("").Funcs(funcMap).ParseFS(h.files,
		"web/templates/layout.html",
		"web/templates/"+pageName+".html",
	))
}

// Dashboard renders the dashboard page
func (h *PageHandler) Dashboard(c *gin.Context) {
	// Check if onboarding is needed
	categories, err := h.categoryRepo.GetAll()
	if err == nil && len(categories) == 0 {
		c.Redirect(http.StatusFound, "/onboarding")
		return
	}

	stats, err := h.insightsService.GetDashboardStats(h.accountRepo)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error loading dashboard: %v", err)
		return
	}

	data := gin.H{
		"Title":      "Dashboard",
		"ActivePage": "dashboard",
		"Stats":      stats,
		"Version":    h.version,
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	h.parseTemplate("dashboard").ExecuteTemplate(c.Writer, "layout.html", data)
}

// Accounts renders the accounts management page
func (h *PageHandler) Accounts(c *gin.Context) {
	accounts, err := h.accountRepo.GetAll()
	if err != nil {
		c.String(http.StatusInternalServerError, "Error loading accounts: %v", err)
		return
	}

	data := gin.H{
		"Title":      "Accounts",
		"ActivePage": "accounts",
		"Accounts":   accounts,
		"Version":    h.version,
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	h.parseTemplate("accounts").ExecuteTemplate(c.Writer, "layout.html", data)
}

// Categories renders the categories management page
func (h *PageHandler) Categories(c *gin.Context) {
	categories, err := h.categoryRepo.GetAllWithPatterns(h.patternRepo)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error loading categories: %v", err)
		return
	}

	data := gin.H{
		"Title":      "Categories",
		"ActivePage": "categories",
		"Categories": categories,
		"Version":    h.version,
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	h.parseTemplate("categories").ExecuteTemplate(c.Writer, "layout.html", data)
}

// Transactions renders the transactions list page
func (h *PageHandler) Transactions(c *gin.Context) {
	filter := repository.TransactionFilter{
		// No limit - return all transactions for client-side pagination
	}

	// Parse filters from query params
	if accountIDStr := c.Query("account_id"); accountIDStr != "" {
		if accountID, err := strconv.Atoi(accountIDStr); err == nil {
			filter.AccountID = &accountID
		}
	}

	if categoryIDStr := c.Query("category_id"); categoryIDStr != "" {
		if categoryID, err := strconv.Atoi(categoryIDStr); err == nil {
			filter.CategoryID = &categoryID
		}
	}

	if categoryTypeStr := c.Query("category_type"); categoryTypeStr != "" {
		if categoryType, err := strconv.Atoi(categoryTypeStr); err == nil {
			ct := model.CategoryType(categoryType)
			filter.CategoryType = &ct
		}
	}

	if batchIDStr := c.Query("batch_id"); batchIDStr != "" {
		if batchID, err := strconv.Atoi(batchIDStr); err == nil {
			filter.BatchID = &batchID
		}
	}

	if c.Query("uncategorized") == "true" {
		filter.Uncategorized = true
	}

	// Parse date filters
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			filter.StartDate = &startDate
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			// Set to end of day to include all transactions on the end date
			endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			filter.EndDate = &endDate
		}
	}

	transactions, err := h.transactionRepo.List(filter)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error loading transactions: %v", err)
		return
	}

	// Get accounts for filter dropdown
	accounts, _ := h.accountRepo.GetAll()

	// Get categories for filter dropdown
	categories, _ := h.categoryRepo.GetAll()

	// Prepare filter values for template
	filterValues := gin.H{
		"AccountID":     c.Query("account_id"),
		"CategoryID":    c.Query("category_id"),
		"Uncategorized": c.Query("uncategorized"),
		"StartDate":     c.Query("start_date"),
		"EndDate":       c.Query("end_date"),
		"BatchID":       c.Query("batch_id"),
	}

	data := gin.H{
		"Title":        "Transactions",
		"ActivePage":   "transactions",
		"Transactions": transactions,
		"Accounts":     accounts,
		"Categories":   categories,
		"Filter":       filterValues,
		"Version":      h.version,
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	h.parseTemplate("transactions").ExecuteTemplate(c.Writer, "layout.html", data)
}

// Import renders the OFX import page
func (h *PageHandler) Import(c *gin.Context) {
	accounts, err := h.accountRepo.GetAll()
	if err != nil {
		c.String(http.StatusInternalServerError, "Error loading accounts: %v", err)
		return
	}

	data := gin.H{
		"Title":      "Import",
		"ActivePage": "import",
		"Accounts":   accounts,
		"Version":    h.version,
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	h.parseTemplate("import").ExecuteTemplate(c.Writer, "layout.html", data)
}

// Onboarding renders the onboarding wizard page
func (h *PageHandler) Onboarding(c *gin.Context) {
	// If categories already exist, redirect to dashboard
	categories, err := h.categoryRepo.GetAll()
	if err == nil && len(categories) > 0 {
		c.Redirect(http.StatusFound, "/")
		return
	}

	data := gin.H{
		"Title":   "Welcome to PrivateLedger",
		"Version": h.version,
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	tmpl := template.Must(template.ParseFS(h.files, "web/templates/onboarding.html"))
	tmpl.Execute(c.Writer, data)
}

// HowToDownload renders the how-to guide page for downloading transactions
func (h *PageHandler) HowToDownload(c *gin.Context) {
	data := gin.H{
		"Title":      "How to Download Transactions",
		"ActivePage": "import",
		"Version":    h.version,
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	h.parseTemplate("how-to-download-transaction").ExecuteTemplate(c.Writer, "layout.html", data)
}

// About renders the about page
func (h *PageHandler) About(c *gin.Context) {
	data := gin.H{
		"Title":      "About",
		"ActivePage": "",
		"Version":    h.version,
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	h.parseTemplate("about").ExecuteTemplate(c.Writer, "layout.html", data)
}
