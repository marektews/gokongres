package buffer

import (
	"log"
	"net/http"
)

func Get_States(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	log.Println("buffer.GetStates called")

	w.WriteHeader(http.StatusNotImplemented)
}
