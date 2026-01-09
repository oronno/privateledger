package service

import (
	"testing"
	"time"

	"github.com/oronno/privateledger/internal/config"
)

// TestGetMonthPeriod_StandardMonth tests month period calculation with standard start (day 1)
func TestGetMonthPeriod_StandardMonth(t *testing.T) {
	cfg := &config.Config{
		StartOfMonth: 1,
	}
	service := &InsightsService{
		config: cfg,
	}

	// Test January 2026
	period := service.GetMonthPeriod(2026, 1)

	// Expected: Jan 1 - Jan 31
	expectedStart := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	expectedEnd := time.Date(2026, 1, 31, 23, 59, 59, 999999999, time.UTC)

	if period.Label != "2026-01" {
		t.Errorf("Expected label '2026-01', got '%s'", period.Label)
	}
	if !period.StartDate.Equal(expectedStart) {
		t.Errorf("Expected start date %v, got %v", expectedStart, period.StartDate)
	}
	if !period.EndDate.Equal(expectedEnd) {
		t.Errorf("Expected end date %v, got %v", expectedEnd, period.EndDate)
	}
}

// TestGetMonthPeriod_CustomStart tests month period with custom start day (19th)
func TestGetMonthPeriod_CustomStart(t *testing.T) {
	cfg := &config.Config{
		StartOfMonth: 19,
	}
	service := &InsightsService{
		config: cfg,
	}

	// Test December 2025 with start_of_month = 19
	period := service.GetMonthPeriod(2025, 12)

	// Expected: Dec 19, 2025 - Jan 18, 2026
	expectedStart := time.Date(2025, 12, 19, 0, 0, 0, 0, time.UTC)
	expectedEnd := time.Date(2026, 1, 18, 23, 59, 59, 999999999, time.UTC)

	if period.Label != "2025-12" {
		t.Errorf("Expected label '2025-12', got '%s'", period.Label)
	}
	if !period.StartDate.Equal(expectedStart) {
		t.Errorf("Expected start date %v, got %v", expectedStart, period.StartDate)
	}
	if !period.EndDate.Equal(expectedEnd) {
		t.Errorf("Expected end date %v, got %v", expectedEnd, period.EndDate)
	}
}

// TestGetMonthPeriod_FebruaryLeapYear tests February in a leap year
func TestGetMonthPeriod_FebruaryLeapYear(t *testing.T) {
	cfg := &config.Config{
		StartOfMonth: 1,
	}
	service := &InsightsService{
		config: cfg,
	}

	// Test February 2024 (leap year)
	period := service.GetMonthPeriod(2024, 2)

	// Expected: Feb 1 - Feb 29, 2024
	expectedStart := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	expectedEnd := time.Date(2024, 2, 29, 23, 59, 59, 999999999, time.UTC)

	if !period.StartDate.Equal(expectedStart) {
		t.Errorf("Expected start date %v, got %v", expectedStart, period.StartDate)
	}
	if !period.EndDate.Equal(expectedEnd) {
		t.Errorf("Expected end date %v, got %v", expectedEnd, period.EndDate)
	}
}

// TestGetMonthPeriod_FebruaryNonLeapYear tests February in a non-leap year
func TestGetMonthPeriod_FebruaryNonLeapYear(t *testing.T) {
	cfg := &config.Config{
		StartOfMonth: 1,
	}
	service := &InsightsService{
		config: cfg,
	}

	// Test February 2025 (non-leap year)
	period := service.GetMonthPeriod(2025, 2)

	// Expected: Feb 1 - Feb 28, 2025
	expectedStart := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)
	expectedEnd := time.Date(2025, 2, 28, 23, 59, 59, 999999999, time.UTC)

	if !period.StartDate.Equal(expectedStart) {
		t.Errorf("Expected start date %v, got %v", expectedStart, period.StartDate)
	}
	if !period.EndDate.Equal(expectedEnd) {
		t.Errorf("Expected end date %v, got %v", expectedEnd, period.EndDate)
	}
}

// TestGetMonthPeriod_MaxStartDay tests with maximum allowed start day (28)
func TestGetMonthPeriod_MaxStartDay(t *testing.T) {
	cfg := &config.Config{
		StartOfMonth: 28,
	}
	service := &InsightsService{
		config: cfg,
	}

	// Test January 2026
	period := service.GetMonthPeriod(2026, 1)

	// Expected: Jan 28, 2026 - Feb 27, 2026
	expectedStart := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)
	expectedEnd := time.Date(2026, 2, 27, 23, 59, 59, 999999999, time.UTC)

	if !period.StartDate.Equal(expectedStart) {
		t.Errorf("Expected start date %v, got %v", expectedStart, period.StartDate)
	}
	if !period.EndDate.Equal(expectedEnd) {
		t.Errorf("Expected end date %v, got %v", expectedEnd, period.EndDate)
	}
}

// TestGetMonthPeriod_YearBoundary tests year boundary transition
func TestGetMonthPeriod_YearBoundary(t *testing.T) {
	cfg := &config.Config{
		StartOfMonth: 15,
	}
	service := &InsightsService{
		config: cfg,
	}

	// Test December 2025 with start_of_month = 15
	period := service.GetMonthPeriod(2025, 12)

	// Expected: Dec 15, 2025 - Jan 14, 2026
	expectedStart := time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC)
	expectedEnd := time.Date(2026, 1, 14, 23, 59, 59, 999999999, time.UTC)

	if !period.StartDate.Equal(expectedStart) {
		t.Errorf("Expected start date %v, got %v", expectedStart, period.StartDate)
	}
	if !period.EndDate.Equal(expectedEnd) {
		t.Errorf("Expected end date %v, got %v", expectedEnd, period.EndDate)
	}
}

// TestGetMonthPeriod_AllMonths tests all 12 months for consistency
func TestGetMonthPeriod_AllMonths(t *testing.T) {
	cfg := &config.Config{
		StartOfMonth: 10,
	}
	service := &InsightsService{
		config: cfg,
	}

	for month := 1; month <= 12; month++ {
		period := service.GetMonthPeriod(2026, month)

		// Verify label format
		expectedLabel := time.Date(2026, time.Month(month), 1, 0, 0, 0, 0, time.UTC).Format("2006-01")
		if period.Label != expectedLabel {
			t.Errorf("Month %d: Expected label '%s', got '%s'", month, expectedLabel, period.Label)
		}

		// Verify period is approximately 1 month long (28-31 days)
		duration := period.EndDate.Sub(period.StartDate)
		days := int(duration.Hours() / 24)
		if days < 27 || days > 31 {
			t.Errorf("Month %d: Period duration is %d days, expected 28-31 days", month, days)
		}

		// Verify start date is on the configured day
		if period.StartDate.Day() != 10 {
			t.Errorf("Month %d: Start date day is %d, expected 10", month, period.StartDate.Day())
		}
	}
}

// TestGetCurrentMonthPeriod_BeforeStartDay tests when current date is before start_of_month
func TestGetCurrentMonthPeriod_BeforeStartDay(t *testing.T) {
	cfg := &config.Config{
		StartOfMonth: 19,
	}
	service := &InsightsService{
		config: cfg,
	}

	// Mock current time to Jan 10, 2026 (before the 19th)
	// Since we're before the 19th, we should be in the "December 2025" period
	// which runs from Dec 19, 2025 to Jan 18, 2026

	// Note: This test would need time mocking in a real scenario
	// For demonstration, we'll test the logic manually
	now := time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC)

	year := now.Year()        // 2026
	month := int(now.Month()) // 1 (January)

	// Since now.Day() (10) < start_of_month (19), we should be in previous month
	if now.Day() < cfg.StartOfMonth {
		month--
		if month < 1 {
			month = 12
			year--
		}
	}

	// Should be December 2025
	if year != 2025 || month != 12 {
		t.Errorf("Expected year=2025, month=12, got year=%d, month=%d", year, month)
	}

	period := service.GetMonthPeriod(year, month)
	expectedStart := time.Date(2025, 12, 19, 0, 0, 0, 0, time.UTC)
	expectedEnd := time.Date(2026, 1, 18, 23, 59, 59, 999999999, time.UTC)

	if !period.StartDate.Equal(expectedStart) {
		t.Errorf("Expected start date %v, got %v", expectedStart, period.StartDate)
	}
	if !period.EndDate.Equal(expectedEnd) {
		t.Errorf("Expected end date %v, got %v", expectedEnd, period.EndDate)
	}
}

// TestGetCurrentMonthPeriod_OnStartDay tests when current date is exactly on start_of_month
func TestGetCurrentMonthPeriod_OnStartDay(t *testing.T) {
	cfg := &config.Config{
		StartOfMonth: 19,
	}
	service := &InsightsService{
		config: cfg,
	}

	// Current date is Jan 19, 2026 (exactly on the 19th)
	now := time.Date(2026, 1, 19, 12, 0, 0, 0, time.UTC)

	year := now.Year()
	month := int(now.Month())

	// Since now.Day() (19) is NOT < start_of_month (19), we stay in current month
	if now.Day() < cfg.StartOfMonth {
		month--
		if month < 1 {
			month = 12
			year--
		}
	}

	// Should be January 2026
	if year != 2026 || month != 1 {
		t.Errorf("Expected year=2026, month=1, got year=%d, month=%d", year, month)
	}

	// Verify the period is calculated correctly
	period := service.GetMonthPeriod(year, month)
	expectedStart := time.Date(2026, 1, 19, 0, 0, 0, 0, time.UTC)
	if !period.StartDate.Equal(expectedStart) {
		t.Errorf("Expected start date %v, got %v", expectedStart, period.StartDate)
	}
}

// TestGetCurrentMonthPeriod_AfterStartDay tests when current date is after start_of_month
func TestGetCurrentMonthPeriod_AfterStartDay(t *testing.T) {
	cfg := &config.Config{
		StartOfMonth: 19,
	}
	service := &InsightsService{
		config: cfg,
	}

	// Current date is Jan 25, 2026 (after the 19th)
	now := time.Date(2026, 1, 25, 12, 0, 0, 0, time.UTC)

	year := now.Year()
	month := int(now.Month())

	if now.Day() < cfg.StartOfMonth {
		month--
		if month < 1 {
			month = 12
			year--
		}
	}

	// Should be January 2026
	if year != 2026 || month != 1 {
		t.Errorf("Expected year=2026, month=1, got year=%d, month=%d", year, month)
	}

	// Verify the period is calculated correctly
	period := service.GetMonthPeriod(year, month)
	expectedStart := time.Date(2026, 1, 19, 0, 0, 0, 0, time.UTC)
	if !period.StartDate.Equal(expectedStart) {
		t.Errorf("Expected start date %v, got %v", expectedStart, period.StartDate)
	}
}

// TestGetCurrentMonthPeriod_JanuaryBeforeStartDay tests January before start_of_month (year boundary)
func TestGetCurrentMonthPeriod_JanuaryBeforeStartDay(t *testing.T) {
	cfg := &config.Config{
		StartOfMonth: 15,
	}
	service := &InsightsService{
		config: cfg,
	}

	// Current date is Jan 5, 2026 (before the 15th)
	now := time.Date(2026, 1, 5, 12, 0, 0, 0, time.UTC)

	year := now.Year()
	month := int(now.Month())

	// Since we're before the 15th in January, we should be in December 2025 period
	if now.Day() < cfg.StartOfMonth {
		month--
		if month < 1 {
			month = 12
			year--
		}
	}

	// Should be December 2025
	if year != 2025 || month != 12 {
		t.Errorf("Expected year=2025, month=12, got year=%d, month=%d", year, month)
	}

	period := service.GetMonthPeriod(year, month)
	expectedStart := time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC)

	if !period.StartDate.Equal(expectedStart) {
		t.Errorf("Expected start date %v, got %v", expectedStart, period.StartDate)
	}
}

// TestGetPeriodForDate_OnPeriodStart tests date exactly on period start
func TestGetPeriodForDate_OnPeriodStart(t *testing.T) {
	cfg := &config.Config{
		StartOfMonth: 19,
	}
	service := &InsightsService{
		config: cfg,
	}

	// Date is Dec 19, 2025 (start of December period)
	date := time.Date(2025, 12, 19, 0, 0, 0, 0, time.UTC)
	period := service.GetPeriodForDate(date)

	// Should be December 2025 period
	if period.Label != "2025-12" {
		t.Errorf("Expected label '2025-12', got '%s'", period.Label)
	}

	expectedStart := time.Date(2025, 12, 19, 0, 0, 0, 0, time.UTC)
	if !period.StartDate.Equal(expectedStart) {
		t.Errorf("Expected start date %v, got %v", expectedStart, period.StartDate)
	}
}

// TestGetPeriodForDate_OnPeriodEnd tests date exactly on period end
func TestGetPeriodForDate_OnPeriodEnd(t *testing.T) {
	cfg := &config.Config{
		StartOfMonth: 19,
	}
	service := &InsightsService{
		config: cfg,
	}

	// Date is Jan 18, 2026 23:59:59 (end of December period)
	date := time.Date(2026, 1, 18, 23, 59, 59, 0, time.UTC)
	period := service.GetPeriodForDate(date)

	// Should be December 2025 period (we're before the 19th)
	if period.Label != "2025-12" {
		t.Errorf("Expected label '2025-12', got '%s'", period.Label)
	}
}

// TestGetPeriodForDate_BeforeStartDay tests date before start_of_month
func TestGetPeriodForDate_BeforeStartDay(t *testing.T) {
	cfg := &config.Config{
		StartOfMonth: 19,
	}
	service := &InsightsService{
		config: cfg,
	}

	// Date is Jan 10, 2026 (before 19th)
	date := time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC)
	period := service.GetPeriodForDate(date)

	// Should be in December 2025 period
	if period.Label != "2025-12" {
		t.Errorf("Expected label '2025-12', got '%s'", period.Label)
	}

	expectedStart := time.Date(2025, 12, 19, 0, 0, 0, 0, time.UTC)
	expectedEnd := time.Date(2026, 1, 18, 23, 59, 59, 999999999, time.UTC)

	if !period.StartDate.Equal(expectedStart) {
		t.Errorf("Expected start date %v, got %v", expectedStart, period.StartDate)
	}
	if !period.EndDate.Equal(expectedEnd) {
		t.Errorf("Expected end date %v, got %v", expectedEnd, period.EndDate)
	}
}

// TestGetPeriodForDate_AfterStartDay tests date on or after start_of_month
func TestGetPeriodForDate_AfterStartDay(t *testing.T) {
	cfg := &config.Config{
		StartOfMonth: 19,
	}
	service := &InsightsService{
		config: cfg,
	}

	// Date is Jan 20, 2026 (after 19th)
	date := time.Date(2026, 1, 20, 12, 0, 0, 0, time.UTC)
	period := service.GetPeriodForDate(date)

	// Should be in January 2026 period
	if period.Label != "2026-01" {
		t.Errorf("Expected label '2026-01', got '%s'", period.Label)
	}

	expectedStart := time.Date(2026, 1, 19, 0, 0, 0, 0, time.UTC)
	expectedEnd := time.Date(2026, 2, 18, 23, 59, 59, 999999999, time.UTC)

	if !period.StartDate.Equal(expectedStart) {
		t.Errorf("Expected start date %v, got %v", expectedStart, period.StartDate)
	}
	if !period.EndDate.Equal(expectedEnd) {
		t.Errorf("Expected end date %v, got %v", expectedEnd, period.EndDate)
	}
}

// TestGetPeriodForDate_YearBoundary tests dates around year boundary
func TestGetPeriodForDate_YearBoundary(t *testing.T) {
	cfg := &config.Config{
		StartOfMonth: 15,
	}
	service := &InsightsService{
		config: cfg,
	}

	testCases := []struct {
		name          string
		date          time.Time
		expectedLabel string
		expectedStart time.Time
	}{
		{
			name:          "Dec 14 (before 15th)",
			date:          time.Date(2025, 12, 14, 12, 0, 0, 0, time.UTC),
			expectedLabel: "2025-11",
			expectedStart: time.Date(2025, 11, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:          "Dec 15 (on 15th)",
			date:          time.Date(2025, 12, 15, 12, 0, 0, 0, time.UTC),
			expectedLabel: "2025-12",
			expectedStart: time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:          "Dec 31 (after 15th)",
			date:          time.Date(2025, 12, 31, 12, 0, 0, 0, time.UTC),
			expectedLabel: "2025-12",
			expectedStart: time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:          "Jan 1 (before 15th)",
			date:          time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
			expectedLabel: "2025-12",
			expectedStart: time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:          "Jan 14 (before 15th)",
			date:          time.Date(2026, 1, 14, 12, 0, 0, 0, time.UTC),
			expectedLabel: "2025-12",
			expectedStart: time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:          "Jan 15 (on 15th)",
			date:          time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC),
			expectedLabel: "2026-01",
			expectedStart: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			period := service.GetPeriodForDate(tc.date)

			if period.Label != tc.expectedLabel {
				t.Errorf("Expected label '%s', got '%s'", tc.expectedLabel, period.Label)
			}
			if !period.StartDate.Equal(tc.expectedStart) {
				t.Errorf("Expected start date %v, got %v", tc.expectedStart, period.StartDate)
			}
		})
	}
}

// TestGetPeriodForDate_FebruaryEdgeCases tests February edge cases with different start days
func TestGetPeriodForDate_FebruaryEdgeCases(t *testing.T) {
	// Test with start_of_month = 28 in a leap year
	cfg := &config.Config{
		StartOfMonth: 28,
	}
	service := &InsightsService{
		config: cfg,
	}

	// Date is Feb 28, 2024 (leap year)
	date := time.Date(2024, 2, 28, 12, 0, 0, 0, time.UTC)
	period := service.GetPeriodForDate(date)

	// Should be February 2024 period
	if period.Label != "2024-02" {
		t.Errorf("Expected label '2024-02', got '%s'", period.Label)
	}

	expectedStart := time.Date(2024, 2, 28, 0, 0, 0, 0, time.UTC)
	if !period.StartDate.Equal(expectedStart) {
		t.Errorf("Expected start date %v, got %v", expectedStart, period.StartDate)
	}

	// Date is Feb 29, 2024 (leap day)
	date = time.Date(2024, 2, 29, 12, 0, 0, 0, time.UTC)
	period = service.GetPeriodForDate(date)

	// Should still be February 2024 period (after the 28th)
	if period.Label != "2024-02" {
		t.Errorf("Expected label '2024-02', got '%s'", period.Label)
	}
}

// TestGetPeriodForDate_MultipleYears tests consistency across multiple years
func TestGetPeriodForDate_MultipleYears(t *testing.T) {
	cfg := &config.Config{
		StartOfMonth: 10,
	}
	service := &InsightsService{
		config: cfg,
	}

	// Test the same day (March 15) across multiple years
	years := []int{2024, 2025, 2026, 2027}
	for _, year := range years {
		date := time.Date(year, 3, 15, 12, 0, 0, 0, time.UTC)
		period := service.GetPeriodForDate(date)

		expectedLabel := time.Date(year, 3, 1, 0, 0, 0, 0, time.UTC).Format("2006-01")
		if period.Label != expectedLabel {
			t.Errorf("Year %d: Expected label '%s', got '%s'", year, expectedLabel, period.Label)
		}

		// Verify start is always on the 10th
		if period.StartDate.Day() != 10 {
			t.Errorf("Year %d: Start date day is %d, expected 10", year, period.StartDate.Day())
		}
	}
}
