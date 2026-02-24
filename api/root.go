package api

import (
    "fmt"
    "net/http"
)

// RootHandler zwraca http.HandlerFunc obsługujący pattern '/'.
func RootHandler(host string, port int) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte(fmt.Sprintf("Hello from %s:%d", host, port)))
    }
}
