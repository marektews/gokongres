package pk

import (
	"context"
	"encoding/json"
	"fmt"
	"gokongres/db"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RequestData struct {
	DepartmentID string  `json:"dep_id"`
	RegNum1      string  `json:"regnum1"`
	RegNum2      *string `json:"regnum2,omitempty"`
	RegNum3      *string `json:"regnum3,omitempty"`
}

/**
* GetCreate to handler HTTP, który obsługuje żądania tworzenia nowego wpisu PK (parking księżycowy/torwar).
* Przyjmuje dane z żądania w formacie JSON, które zawierają ID działu oraz numery rejestracyjne pojazdu.
* Handler sprawdza, czy podany dział istnieje, a następnie poszukuje nieużywanego numeru identyfikatora dla tego działu.
* Jeśli znajdzie dostępny numer, tworzy nowy wpis PK z tym numerem i zwraca jego ID w odpowiedzi JSON.
* Jeśli wystąpią błędy (np. dział nie istnieje, brak dostępnych numerów, numery rejestracyjne są już używane), handler zwraca odpowiedni kod błędu HTTP i komunikat.
 */
func Get_CreatePassID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var reqData RequestData
	err := json.NewDecoder(r.Body).Decode(&reqData)
	if err != nil {
		log.Printf("Invalid request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	dep_ID, err := primitive.ObjectIDFromHex(reqData.DepartmentID)
	if err != nil {
		log.Printf("Invalid department ID: %v", err)
		http.Error(w, "Invalid department ID", http.StatusBadRequest)
		return
	}

	collDeps := db.Collection("departments")
	if collDeps == nil {
		log.Println("Collection 'departments' not found")
		http.Error(w, "Collection 'departments' not found", http.StatusInternalServerError)
		return
	}

	var department db.Department
	err = collDeps.FindOne(r.Context(), bson.M{"_id": dep_ID}).Decode(&department)
	if err != nil {
		log.Printf("Department not found: %v", err)
		http.Error(w, "Department not found", http.StatusNotFound)
		return
	}

	// poszukiwanie nieużywanego jeszcze numeru identyfikatora
	usedNumbers := make([]int, department.Plimit)
	for i := range usedNumbers {
		usedNumbers[i] = i + 1
	}

	unusedNumbers, err := findUnusedNumbers(r.Context(), &reqData, department.ID, usedNumbers)
	if err != nil {
		log.Printf("Error finding unused numbers: %v", err)
		http.Error(w, "Error finding unused numbers", http.StatusInternalServerError)
		return
	}

	if len(unusedNumbers) == 0 {
		log.Printf("No unused number available for department %s", department.Name)
		http.Error(w, "No unused number available for this department", http.StatusConflict)
		return
	}

	// tworzenie nowego wpisu z nowym identyfikatorem
	dpk := db.DepartmentPK{
		DepartmentID: department.ID,
		PassNr:       unusedNumbers[0],
		Registered:   primitive.NewDateTimeFromTime(time.Now()),
		RegNum1:      reqData.RegNum1,
	}
	if reqData.RegNum2 != nil {
		dpk.RegNum2 = reqData.RegNum2
	}
	if reqData.RegNum3 != nil {
		dpk.RegNum3 = reqData.RegNum3
	}

	collDepsPK := db.Collection("departments_pk")
	if collDepsPK == nil {
		log.Println("Collection 'departments_pk' not found")
		http.Error(w, "Collection 'departments_pk' not found", http.StatusInternalServerError)
		return
	}

	insRes, err := collDepsPK.InsertOne(r.Context(), dpk)
	if err != nil {
		log.Printf("Error inserting new DepartmentPK entry: %v", err)
		http.Error(w, "Error creating new PK entry", http.StatusInternalServerError)
		return
	}

	type ResponseData struct {
		PassID primitive.ObjectID `json:"passID"`
	}
	respData := ResponseData{
		PassID: insRes.InsertedID.(primitive.ObjectID),
	}
	err = json.NewEncoder(w).Encode(respData)
	if err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

/**
 * Funkcja pomocnicza do znajdowania nieużywanych numerów identyfikatorów oraz sprawdzania, czy podane numery rejestracyjne nie są już używane w innych wpisach DepartmentPK.
 * Przyjmuje kontekst, dane z żądania, ID działu oraz listę używanych numerów identyfikatorów.
 * Zwraca listę nieużywanych numerów identyfikatorów lub błąd, jeśli wystąpił problem podczas przeszukiwania bazy danych.
 */
func findUnusedNumbers(ctx context.Context, reqData *RequestData, congregationID primitive.ObjectID, usedNumbers []int) ([]int, error) {
	coll := db.Collection("departments_pk")
	if coll == nil {
		log.Println("Collection 'departments_pk' not found")
		return nil, fmt.Errorf("collection 'departments_pk' not found")
	}
	cur, err := coll.Find(ctx, bson.M{"department_id": congregationID})
	if err != nil {
		return nil, fmt.Errorf("error finding departments_pk entries: %v", err)
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var departmentPK db.DepartmentPK
		if err := cur.Decode(&departmentPK); err != nil {
			log.Printf("Error decoding DepartmentPK document: %v", err)
			continue
		}

		for i, v := range usedNumbers {
			if v == departmentPK.PassNr {
				usedNumbers = append(usedNumbers[:i], usedNumbers[i+1:]...)
				break
			}
		}

		// testowanie czy pojazd nie występuje już na innym identyfikatorze
		tmpRN1 := reqData.RegNum1
		tmpRN2 := ""
		if reqData.RegNum2 != nil {
			tmpRN2 = *reqData.RegNum2
		}
		tmpRN3 := ""
		if reqData.RegNum3 != nil {
			tmpRN3 = *reqData.RegNum3
		}

		if departmentPK.RegNum1 == tmpRN1 || departmentPK.RegNum1 == tmpRN2 || departmentPK.RegNum1 == tmpRN3 {
			return nil, fmt.Errorf("registration number %s is already used in another DepartmentPK entry", departmentPK.RegNum1)
		}
		if departmentPK.RegNum2 != nil && (*departmentPK.RegNum2 == tmpRN1 || *departmentPK.RegNum2 == tmpRN2 || *departmentPK.RegNum2 == tmpRN3) {
			return nil, fmt.Errorf("registration number %s is already used in another DepartmentPK entry", *departmentPK.RegNum2)
		}
		if departmentPK.RegNum3 != nil && (*departmentPK.RegNum3 == tmpRN1 || *departmentPK.RegNum3 == tmpRN2 || *departmentPK.RegNum3 == tmpRN3) {
			return nil, fmt.Errorf("registration number %s is already used in another DepartmentPK entry", *departmentPK.RegNum3)
		}
	}

	return usedNumbers, nil
}
