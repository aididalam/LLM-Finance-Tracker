package middleware

import (
	"context"
	"net/http"

	"github.com/aididalam/llmexpensetracker/internal/domain"
)

type userContextKey string

const currentUserKey userContextKey = "current_user"

func DefaultUser(next http.Handler) http.Handler {
	user := &domain.User{
		ID:          domain.DefaultUserID,
		FirstName:   "Demo",
		LastName:    "User",
		DisplayName: "Demo User",
		Email:       "demo.user@example.com",
		Status:      "active",
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), currentUserKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func CurrentUser(ctx context.Context) *domain.User {
	user, _ := ctx.Value(currentUserKey).(*domain.User)
	return user
}

func CurrentUserID(ctx context.Context) string {
	if user := CurrentUser(ctx); user != nil && user.ID != "" {
		return user.ID
	}
	return domain.DefaultUserID
}
