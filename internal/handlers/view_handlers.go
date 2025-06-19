package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	// "github.com/zayyadi/finance-tracker/internal/models" // No longer needed for UserResponse here
)

// ViewHandler handles rendering of HTML pages.
type ViewHandler struct {
}

// NewViewHandler creates a new ViewHandler.
func NewViewHandler() *ViewHandler {
	return &ViewHandler{}
}

// ShowHomePage redirects to the dashboard page.
func (vh *ViewHandler) ShowHomePage(c *gin.Context) {
	c.Redirect(http.StatusFound, "/dashboard")
}

// ShowDashboardPage renders the main dashboard page.
// User-specific data is no longer passed from server-side for template rendering.
// The Vue app will handle data fetching and display.
func (vh *ViewHandler) ShowDashboardPage(c *gin.Context) {
	c.HTML(http.StatusOK, "layouts/main.html", gin.H{
		"CurrentYear": time.Now().Year(),
		// "IsAuthenticated" and "User" are removed as auth is handled client-side or not at all
	})
}

// IsUserAuthenticated and GetUserFromContext are removed as they are no longer needed.
// The GetUserIDFromContext in utils.go will be removed in a subsequent step if unused.
/*
// IsUserAuthenticated checks if a user is authenticated.
// Since AuthMiddleware is removed, this will always return false or needs redefinition.
// For a single-user mode, we might consider the "user" to always be "authenticated" locally.
// For now, let's make it return false as no auth mechanism is in place.
// This will affect template conditionals.
func IsUserAuthenticated(c *gin.Context) bool {
	// _, exists := c.Get("userID") // "userID" is no longer set by middleware
	return false // No authentication in place
}

// GetUserFromContext retrieves user ID and email from context if available.
// Since AuthMiddleware is removed, "userID" and "email" are no longer in the context.
// This function will return placeholder/default values.
func GetUserFromContext(c *gin.Context) (userID uint, email string, isAuthenticated bool) {
	// "userID" and "email" are no longer set by AuthMiddleware.
	// We use the modified GetUserIDFromContext which returns a placeholder ID.
	placeholderUserID, _ := GetUserIDFromContext(c) // Error is ignored as it's a placeholder.

	// For single-user mode, we can assume a default "user" is always active.
	// This is a placeholder until UserID is fully removed from services.
	// The 'isAuthenticated' flag here reflects the old multi-user auth concept.
	// In a true single-user app, this concept might change.
	return placeholderUserID, "user@local.host", false // isAuthenticated is false as no login.
}
*/
