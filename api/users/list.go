package users

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"
)

// Get_All zwraca listę kont admin (bez hashy haseł).
func Get_All(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if _, ok := guard(w, r); !ok {
		return
	}

	users, err := db.GetAllUsers(r.Context())
	if err != nil {
		log.Printf("users.Get_All: %v", err)
		http.Error(w, "Error fetching users", http.StatusInternalServerError)
		return
	}

	resp := make([]userResp, 0, len(users))
	for _, u := range users {
		resp = append(resp, userResp{
			ID:          u.ID.Hex(),
			Login:       u.Username,
			Fn:          u.FirstName,
			Ln:          u.LastName,
			Permissions: u.Permissions,
		})
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("users.Get_All: encode: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}
