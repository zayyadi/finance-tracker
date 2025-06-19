package models

import (
	"time"

	"github.com/zayyadi/finance-tracker/internal/types"
	"gorm.io/gorm"
)

type Notification struct {
	gorm.Model // ID, CreatedAt, UpdatedAt, DeletedAt
	// Unchanged lines	// UserID  uint      `json:"user_id"` // To associate notification with a user
	Message string `json:"message" binding:"required"`
	IsRead  bool   `json:"is_read" gorm:"default:false"`
	// DueDate for reminders about upcoming due dates
	DueDate types.CustomDate `json:"due_date"`
	// RelatedType can be 'debt', 'savings_goal', etc.
	RelatedType string `json:"related_type,omitempty"`
	RelatedID   uint   `json:"related_id,omitempty"` // ID of the related debt or savings goal
}

// NotificationCreateRequest defines the expected request body for creating a notification.
type NotificationCreateRequest struct {
	Message     string    `json:"message" binding:"required"`
	DueDate     time.Time `json:"due_date"`
	RelatedType string    `json:"related_type"`
	RelatedID   uint      `json:"related_id"`
}

// NotificationUpdateRequest defines the expected request body for updating a notification.
type NotificationUpdateRequest struct {
	Message     *string    `json:"message,omitempty"`      // Pointer to allow optional updates
	IsRead      *bool      `json:"is_read,omitempty"`      // Pointer to allow optional updates
	DueDate     *time.Time `json:"due_date,omitempty"`     // Pointer to allow optional updates
	RelatedType *string    `json:"related_type,omitempty"` // Pointer to allow optional updates
	RelatedID   *uint      `json:"related_id,omitempty"`   // Pointer to allow optional updates
}
