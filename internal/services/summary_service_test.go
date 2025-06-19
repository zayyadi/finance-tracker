package services

import (
	"testing"
	"time"
)

func TestCalculatePeriodDates(t *testing.T) {
	loc, _ := time.LoadLocation("UTC") // Use a consistent timezone for tests

	testCases := []struct {
		name            string
		targetDate      time.Time
		summaryType     string
		expectedStart   time.Time
		expectedEnd     time.Time
		expectError     bool
		expectedErrorMsg string
	}{
		// Monthly tests
		{
			name:          "Monthly_MidMonth",
			targetDate:    time.Date(2023, time.November, 15, 0, 0, 0, 0, loc),
			summaryType:   "monthly",
			expectedStart: time.Date(2023, time.November, 1, 0, 0, 0, 0, loc),
			expectedEnd:   time.Date(2023, time.November, 30, 0, 0, 0, 0, loc),
			expectError:   false,
		},
		{
			name:          "Monthly_StartOfMonth",
			targetDate:    time.Date(2023, time.February, 1, 0, 0, 0, 0, loc),
			summaryType:   "monthly",
			expectedStart: time.Date(2023, time.February, 1, 0, 0, 0, 0, loc),
			expectedEnd:   time.Date(2023, time.February, 28, 0, 0, 0, 0, loc), // Non-leap year
			expectError:   false,
		},
		{
			name:          "Monthly_LeapYear",
			targetDate:    time.Date(2024, time.February, 10, 0, 0, 0, 0, loc),
			summaryType:   "monthly",
			expectedStart: time.Date(2024, time.February, 1, 0, 0, 0, 0, loc),
			expectedEnd:   time.Date(2024, time.February, 29, 0, 0, 0, 0, loc), // Leap year
			expectError:   false,
		},

		// Weekly tests (assuming Monday is the start of the week)
		{
			name:          "Weekly_MidWeek_Wednesday", // Wednesday
			targetDate:    time.Date(2023, time.November, 15, 0, 0, 0, 0, loc), // 2023-11-15 is a Wednesday
			summaryType:   "weekly",
			expectedStart: time.Date(2023, time.November, 13, 0, 0, 0, 0, loc), // Monday
			expectedEnd:   time.Date(2023, time.November, 19, 0, 0, 0, 0, loc), // Sunday
			expectError:   false,
		},
		{
			name:          "Weekly_Monday",
			targetDate:    time.Date(2023, time.November, 13, 0, 0, 0, 0, loc), // Monday
			summaryType:   "weekly",
			expectedStart: time.Date(2023, time.November, 13, 0, 0, 0, 0, loc),
			expectedEnd:   time.Date(2023, time.November, 19, 0, 0, 0, 0, loc),
			expectError:   false,
		},
		{
			name:          "Weekly_Sunday",
			targetDate:    time.Date(2023, time.November, 19, 0, 0, 0, 0, loc), // Sunday
			summaryType:   "weekly",
			expectedStart: time.Date(2023, time.November, 13, 0, 0, 0, 0, loc), // Monday of that week
			expectedEnd:   time.Date(2023, time.November, 19, 0, 0, 0, 0, loc), // Sunday
			expectError:   false,
		},
		 {
			name:          "Weekly_AcrossMonthBoundary",
			targetDate:    time.Date(2023, time.October, 30, 0, 0, 0, 0, loc), // Monday Oct 30
			summaryType:   "weekly",
			expectedStart: time.Date(2023, time.October, 30, 0, 0, 0, 0, loc),
			expectedEnd:   time.Date(2023, time.November, 5, 0, 0, 0, 0, loc),
			expectError:   false,
		},


		// Yearly tests
		{
			name:          "Yearly_AnyDate",
			targetDate:    time.Date(2023, time.July, 15, 0, 0, 0, 0, loc),
			summaryType:   "yearly",
			expectedStart: time.Date(2023, time.January, 1, 0, 0, 0, 0, loc),
			expectedEnd:   time.Date(2023, time.December, 31, 0, 0, 0, 0, loc),
			expectError:   false,
		},

		// Error cases
		{
			name:            "InvalidSummaryType",
			targetDate:      time.Now(),
			summaryType:     "daily", // Assuming 'daily' is not supported
			expectError:     true,
			expectedErrorMsg: "invalid summary type: daily",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			startDate, endDate, err := CalculatePeriodDates(tc.targetDate, tc.summaryType)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none. Start: %v, End: %v", startDate, endDate)
				} else if err.Error() != tc.expectedErrorMsg {
					t.Errorf("Expected error message '%s', but got '%s'", tc.expectedErrorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect error, but got: %v", err)
				}
				if !startDate.Equal(tc.expectedStart) {
					t.Errorf("Expected start date %v, but got %v", tc.expectedStart, startDate)
				}
				if !endDate.Equal(tc.expectedEnd) {
					t.Errorf("Expected end date %v, but got %v", tc.expectedEnd, endDate)
				}
			}
		})
	}
}
