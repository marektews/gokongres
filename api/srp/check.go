package srp

import (
	"log"
	"net/http"
)

func Get_CheckPass(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	passNr := r.PathValue("pass_nr")
	regNum1 := r.PathValue("regnum1")
	regNum2 := r.PathValue("regnum2")
	regNum3 := r.PathValue("regnum3")

	// Walidacja argumentów
	if passNr == "" || regNum1 == "" || regNum2 == "" || regNum3 == "" {
		log.Println("Missing required arguments")
		http.Error(w, "Missing required arguments", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNotImplemented)
}
