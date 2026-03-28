# Obsługa Sesji HTTP w gokongres

## Przegląd

Projekt `gokongres` teraz ma wbudowaną obsługę sesji HTTP za pomocą biblioteki `gorilla/sessions`. Sesje są przechowywane w ciasteczkach (cookies) z szyfrowaniem.

## Konfiguracja

W pliku `main.go` sesje są inicjalizowane z kluczami autentykacji i szyfrowania:

```go
authKey := []byte("auth-key-change-me-in-production")
encKey := []byte("enc-key-change-me-in-production-1234567890ab")
if err := api.InitSessions(authKey, encKey); err != nil {
    log.Fatalf("Failed to initialize sessions: %v", err)
}
```

**Ważne:** W środowisku produkcyjnym zmień klucze na bezpieczne wartości!

## API Endpointy

### 1. Logowanie
**POST** `/api/auth/login`

Request:
```json
{
  "username": "user123",
  "password": "password123"
}
```

Response:
```json
{
  "success": true,
  "message": "Logged in successfully",
  "user": {
    "user_id": "user-123",
    "username": "user123",
    "roles": ["user"]
  }
}
```

### 2. Wylogowanie
**POST** `/api/auth/logout`

Response:
```json
{
  "success": true,
  "message": "Logged out successfully"
}
```

### 3. Pobranie informacji o zalogowanym użytkowniku
**GET** `/api/auth/me`

Response:
```json
{
  "success": true,
  "user": {
    "user_id": "user-123",
    "username": "user123",
    "roles": ["user"]
  }
}
```

## Middleware

### SessionMiddleware
Middleware globalne dla wszystkich żądań. Dodaje nagłówki cache-control dla sesji.

### AuthMiddleware
Middleware do sprawdzania czy użytkownik jest zalogowany. Zwraca 401 jeśli nie.

Przykład użycia:
```go
r.HandleFunc("/api/protected", AuthMiddleware(protectedHandler)).Methods(http.MethodGet)
```

## Struktura danych sesji

```go
type SessionData struct {
    UserID   string
    Username string
    Roles    []string
}
```

## Funkcje pomocnicze

- `InitSessions(authKey, encKey []byte)` - Inicjalizuje magazyn sesji
- `GetSession(r *http.Request)` - Pobiera obiekt sesji
- `SetSessionData(w, r, data)` - Ustawia dane sesji
- `GetSessionData(r)` - Pobiera dane sesji
- `ClearSession(w, r)` - Usuwa sesję

## Cechy sesji

- **Duracja:** 7 dni
- **HttpOnly:** Tak (ochrona przed XSS)
- **SameSite:** Lax (ochrona przed CSRF)
- **Ścieżka:** / (dostępne na całej stronie)

## Integracja z bazą danych

### Inicjalizacja bazy danych

Przy starcie aplikacji należy również zainicjalizować połączenie z MongoDB (w pliku `main.go`):

```go
import "gokongres/db"

func main() {
    // ... reszta kodu
    
    // Inicjalizuj połączenie z bazą danych
    if err := db.Connect(context.Background(), ""); err != nil {
        log.Fatalf("Failed to connect to MongoDB: %v", err)
    }
    defer db.Disconnect(context.Background())
    
    // Inicjalizuj sesje
    authKey := []byte("auth-key-change-me-in-production")
    encKey := []byte("enc-key-change-me-in-production-1234567890ab")
    if err := api.InitSessions(authKey, encKey); err != nil {
        log.Fatalf("Failed to initialize sessions: %v", err)
    }
    
    api.RegisterHandlers(*host, *port)
    // ... reszta kodu
}
```

**Zmienne środowiskowe:**
- `MONGODB_URI` - URI połączenia MongoDB (domyślnie: `mongodb://localhost:27017`)
- `MONGODB_DB` - nazwa bazy danych (domyślnie: `gokongres`)

### Logowanie użytkownika

Endpoint logowania `POST /api/auth/login` teraz:

1. Pobiera użytkownika z kolekcji `Users` w MongoDB po polu `username`
2. Weryfikuje hasło porównując hash bcrypt
3. Mapuje pola uprawnień na listę ról:
   - `is_sra` → `"sra"`
   - `is_srp` → `"srp"`
   - `is_pk` → `"pk"`
   - `is_rja` → `"rja"`
   - `is_monitoring` → `"monitoring"`
   - `is_users` → `"users"`
   - `is_limits` → `"limits"`

### Struktura użytkownika w bazie

```go
type User struct {
    ID           primitive.ObjectID
    Login        int
    Username     string
    Hash         string              // bcrypt hash hasła
    FirstName    string
    LastName     string
    IsSra        bool                // true lub false
    IsSrp        bool                // true lub false
    IsPk         bool                // true lub false
    IsRja        bool                // true lub false
    IsMonitoring bool                // true lub false
    IsUsers      bool                // true lub false
    IsLimits     bool                // true lub false
}
```

### Funkcje dostępu do bazy

#### Pobieranie użytkownika po nazwie

```go
user, err := db.GetUserByUsername(ctx, "username")
```

#### Pobieranie użytkownika po ID

```go
user, err := db.GetUserByID(ctx, "507f1f77bcf86cd799439011")
```

#### Pobieranie użytkownika po loginie

```go
user, err := db.GetUserByLogin(ctx, 123)
```

#### Pobieranie ról użytkownika

```go
roles := user.GetRoles()  // []string{"sra", "monitoring"}
```

#### Pobieranie pełnego nazwiska

```go
fullName := user.GetFullName()  // "Jan Kowalski"
```

## Integracja z autoryzacją

Logowanie jest już w pełni zintegrowane z bazą danych MongoDB. `LoginHandler` w pliku `auth.go`:

1. ✅ Pobiera dane użytkownika z kolekcji `Users` po nazwie użytkownika
2. ✅ Weryfikuje hasło za pomocą bcrypt
3. ✅ Pobiera role użytkownika i mapuje je na podstawie pól uprawnień
4. ✅ Tworzy sesję z pełnymi danymi użytkownika

### Wymogi dla bazy danych

Kolekcja `Users` musi zawierać dokumenty z następującymi polami:

```json
{
  "_id": ObjectId(...),
  "login": 1,
  "username": "jankowalski",
  "hash": "$2a$12$...",
  "first_name": "Jan",
  "last_name": "Kowalski",
  "is_sra": true,
  "is_srp": false,
  "is_pk": true,
  "is_rja": false,
  "is_monitoring": true,
  "is_users": false,
  "is_limits": false
}
```

### Generowanie hasła bcrypt

Aby dodać użytkownika do bazy z poprawnym hashem:

```go
import "golang.org/x/crypto/bcrypt"

password := "haslo123"
hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
// Wstaw hashedPassword do pola 'hash'
```
