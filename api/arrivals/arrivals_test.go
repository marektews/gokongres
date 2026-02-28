package arrivals

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
)

func resetState() {
    mu.Lock()
    arrivalsList = nil
    nextID = 1
    mu.Unlock()
}

func TestGetAll_Empty(t *testing.T) {
    resetState()

    rr := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/api/arrivals/all", nil)
    GetAll(rr, req)

    var got []Arrival
    if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
        t.Fatalf("decode response: %v", err)
    }
    if len(got) != 0 {
        t.Fatalf("expected empty list, got %d items", len(got))
    }
}

func TestSetAndGetAll(t *testing.T) {
    resetState()

    payload := map[string]string{"message": "hello"}
    b, _ := json.Marshal(payload)
    req := httptest.NewRequest(http.MethodPost, "/api/arrivals", bytes.NewReader(b))
    req.Header.Set("Content-Type", "application/json")
    rr := httptest.NewRecorder()
    Set(rr, req)

    if rr.Code != http.StatusCreated {
        t.Fatalf("expected status %d, got %d", http.StatusCreated, rr.Code)
    }

    var created Arrival
    if err := json.NewDecoder(rr.Body).Decode(&created); err != nil {
        t.Fatalf("decode created: %v", err)
    }
    if created.ID != 1 || created.Message != "hello" {
        t.Fatalf("unexpected created arrival: %+v", created)
    }

    // now get all
    rr2 := httptest.NewRecorder()
    req2 := httptest.NewRequest(http.MethodGet, "/api/arrivals/all", nil)
    GetAll(rr2, req2)

    var list []Arrival
    if err := json.NewDecoder(rr2.Body).Decode(&list); err != nil {
        t.Fatalf("decode list: %v", err)
    }
    if len(list) != 1 {
        t.Fatalf("expected 1 item, got %d", len(list))
    }
    if list[0].ID != 1 || list[0].Message != "hello" {
        t.Fatalf("unexpected item: %+v", list[0])
    }
}

func TestSet_MethodNotAllowed(t *testing.T) {
    resetState()

    req := httptest.NewRequest(http.MethodGet, "/api/arrivals", nil)
    rr := httptest.NewRecorder()
    Set(rr, req)

    if rr.Code != http.StatusMethodNotAllowed {
        t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
    }
}

func TestSet_BadJSON(t *testing.T) {
    resetState()

    req := httptest.NewRequest(http.MethodPost, "/api/arrivals", bytes.NewReader([]byte("not json")))
    req.Header.Set("Content-Type", "application/json")
    rr := httptest.NewRecorder()
    Set(rr, req)

    if rr.Code != http.StatusBadRequest {
        t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
    }
}
