package users

import (
	"gokongres/db"
	"gokongres/sessions"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// guard sprawdza sesję i uprawnienie is_users. Zwraca zalogowanego użytkownika
// oraz false (po wysłaniu 401/403), gdy dostęp jest niedozwolony.
func guard(w http.ResponseWriter, r *http.Request) (*db.User, bool) {
	sd := sessions.GetSessionData(r)
	if sd == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return nil, false
	}
	oid, err := primitive.ObjectIDFromHex(sd.UserID)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return nil, false
	}
	u, err := db.GetUserByID(r.Context(), oid)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return nil, false
	}
	if !u.Permissions.IsUsers {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return nil, false
	}
	return u, true
}

type userResp struct {
	ID          string             `json:"id"`
	Login       string             `json:"login"`
	Fn          string             `json:"fn"`
	Ln          string             `json:"ln"`
	Permissions db.UserPermissions `json:"permissions"`
}

type userReq struct {
	ID          string             `json:"id"`
	Login       string             `json:"login"`
	Password    string             `json:"password"`
	Fn          string             `json:"fn"`
	Ln          string             `json:"ln"`
	Permissions db.UserPermissions `json:"permissions"`
}
