package services

import (
	"log"
	"time"

	"github.com/zayyadi/finance-tracker/internal/database" // Assuming database.GetDB() is available
	"github.com/zayyadi/finance-tracker/internal/models"
	"gorm.io/gorm"
)

// NotificationService handles scheduled tasks like checking due dates and goals.
type NotificationService struct {
	DB *gorm.DB
}

// NewNotificationService creates a new NotificationService.
func NewNotificationService(db *gorm.DB) *NotificationService {
	if db == nil {
		log.Println("Warning: NewNotificationService called with nil DB, attempting to use global GetDB()")
		db = database.GetDB() // Fallback, ensure DB is initialized before this service
	}
	return &NotificationService{DB: db}
}

// CheckDueDatesAndGoals logs reminders for upcoming debts and savings goals.
func (s *NotificationService) CheckDueDatesAndGoals() {
	if s.DB == nil {
		log.Println("NotificationService: DB not initialized. Skipping checks.")
		return
	}
	log.Println("NotificationService: Starting CheckDueDatesAndGoals...")

	now := time.Now()
	sevenDaysFromNow := now.AddDate(0, 0, 7)

	// Check for upcoming debts
	var upcomingDebts []models.Debt
	// Removed UserID from query
	err := s.DB.Where("status = ? AND due_date BETWEEN ? AND ?", "Pending", now, sevenDaysFromNow).
		Find(&upcomingDebts).Error
	if err != nil {
		log.Printf("NotificationService: Error fetching upcoming debts: %v", err)
	} else {
		if len(upcomingDebts) > 0 {
			log.Printf("NotificationService: Found %d upcoming debt(s).", len(upcomingDebts))
			for _, debt := range upcomingDebts {
				// Removed UserID from log message
				log.Printf("Reminder: Debt for '%s' of amount %.2f is due on %s.",
					debt.DebtorName, debt.Amount, debt.DueDate.Format("2006-01-02"))
			}
		} else {
			log.Println("NotificationService: No upcoming debts found in the next 7 days.")
		}
	}

	// Check for approaching savings goals
	var upcomingSavings []models.Savings
	// Removed UserID from query
	err = s.DB.Where("target_date IS NOT NULL AND target_date BETWEEN ? AND ? AND current_amount < goal_amount", now, sevenDaysFromNow).
		Find(&upcomingSavings).Error
	if err != nil {
		log.Printf("NotificationService: Error fetching upcoming savings goals: %v", err)
	} else {
		if len(upcomingSavings) > 0 {
			log.Printf("NotificationService: Found %d approaching savings goal(s).", len(upcomingSavings))
			for _, sg := range upcomingSavings {
				// Removed UserID from log message
				log.Printf("Reminder: Savings goal '%s' (Target: %.2f, Current: %.2f) is approaching its target date %s.",
					sg.GoalName, sg.GoalAmount, sg.CurrentAmount, sg.TargetDate.Format("2006-01-02"))
			}
		} else {
			log.Println("NotificationService: No savings goals approaching target date in the next 7 days or they are already met.")
		}
	}
	log.Println("NotificationService: CheckDueDatesAndGoals finished.")
}
