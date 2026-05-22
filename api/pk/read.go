package pk

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

/**
* 	Odczyt stanu identyfikatora
*   :param pk_id: private key w bazie
*   :return: {
*       "passid": "<private key rekordu>",
*       "pass_nr": <numer identyfikatora>,
*       "regnum1": "<numer rejestracyjny na piątek lub na wszystkie dni>",
*       "regnum2": "<numer rejestracyjny na sobotę>",
*       "regnum3": "<numer rejestracyjny na niedzielę>",
*   }
 */
func Get_ReadPassData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	pk_id := r.PathValue("pk_id")
	pkID, err := primitive.ObjectIDFromHex(pk_id)
	if err != nil {
		log.Printf("Invalid pk_id format: %v", err)
		http.Error(w, "Invalid pk_id format", http.StatusBadRequest)
		return
	}

	collDepPK := db.Collection("departments_pk")
	if collDepPK == nil {
		log.Println("Collection 'departments_pk' not found")
		http.Error(w, "Collection 'departments_pk' not found", http.StatusInternalServerError)
		return
	}

	var depPK db.DepartmentPK
	err = collDepPK.FindOne(r.Context(), bson.M{"_id": pkID}).Decode(&depPK)
	if err != nil {
		log.Println("Error finding document:", err)
		http.Error(w, "Error finding document", http.StatusInternalServerError)
		return
	}

	type ResponseData struct {
		PassID  string `json:"passid"`
		PassNr  int    `json:"pass_nr"`
		RegNum1 string `json:"regnum1"`
		RegNum2 string `json:"regnum2"`
		RegNum3 string `json:"regnum3"`
	}
	respData := ResponseData{
		PassID:  depPK.ID.Hex(),
		PassNr:  depPK.PassNr,
		RegNum1: depPK.RegNum1,
	}
	if depPK.RegNum2 != nil {
		respData.RegNum2 = *depPK.RegNum2
	}
	if depPK.RegNum3 != nil {
		respData.RegNum3 = *depPK.RegNum3
	}
	err = json.NewEncoder(w).Encode(respData)
	if err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
