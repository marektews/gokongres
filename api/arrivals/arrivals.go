package arrivals

import "sync"

type Arrival struct {
    ID      int    `json:"id"`
    Message string `json:"message"`
}

var (
    mu           sync.Mutex
    arrivalsList []Arrival
    nextID       = 1
)
