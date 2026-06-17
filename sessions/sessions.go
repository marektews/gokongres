package sessions

import (
	"encoding/gob"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
)

// SessionData przechowuje dane sesji
type SessionData struct {
	UserID   string
	Username string
}

var (
	// store przechowuje sesje HTTP
	store sessions.Store
	// sessionName to nazwa ciasteczka sesji
	sessionName = "kongres_session"
)

// InitSessions inicjalizuje magazyn sesji
func InitSessions(authKey, encKey [32]byte) error {
	// Rejestrujemy typ SessionData do gob encoding
	gob.Register(SessionData{})

	store = sessions.NewCookieStore(authKey[:], encKey[:])
	store.(*sessions.CookieStore).Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 dni
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	return nil
}

// GetSession pobiera sesję z żądania HTTP
func GetSession(r *http.Request) (*sessions.Session, error) {
	return store.Get(r, sessionName)
}

// SetSessionData ustawia dane sesji
func SetSessionData(w http.ResponseWriter, r *http.Request, data SessionData) error {
	session, err := store.Get(r, sessionName)
	if err != nil {
		return err
	}

	session.Values["uid"] = data.UserID
	session.Values["username"] = data.Username

	return session.Save(r, w)
}

// GetSessionData pobiera dane sesji
func GetSessionData(r *http.Request) *SessionData {
	session, err := store.Get(r, sessionName)
	if err != nil {
		log.Printf("GetSessionData: Error getting session: %v", err)
		return nil
	}

	userID, ok := session.Values["uid"].(string)
	if !ok || userID == "" {
		log.Printf("GetSessionData: No user ID in session")
		return nil
	}

	username, _ := session.Values["username"].(string)
	return &SessionData{
		UserID:   userID,
		Username: username,
	}
}

// ClearSession usuwa sesję
func ClearSession(w http.ResponseWriter, r *http.Request) error {
	session, err := store.Get(r, sessionName)
	if err != nil {
		log.Printf("ClearSession: Error getting session: %v", err)
		return err
	}

	session.Options.MaxAge = -1
	return session.Save(r, w)
}

// SessionMiddleware middleware do obsługi sesji
func SessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Ustawienie nagłówka aby sesja była zawsze dostępna
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		next.ServeHTTP(w, r)
	})
}

// AuthMiddleware middleware do sprawdzania czy użytkownik jest zalogowany
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionData := GetSessionData(r)
		if sessionData == nil || sessionData.UserID == "" {
			log.Printf("AuthMiddleware: Unauthorized access attempt")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
