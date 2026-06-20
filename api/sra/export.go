package sra

import (
	"context"
	"fmt"
	"gokongres/db"
	"log"
	"net/http"

	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/**
*	Eksport zgłoszeń autokarów (SRA) do pliku XLSX (moduł ADMSRA).
*	Odwzorowuje układ kolumn z poprzedniej implementacji (python/kongres/api/sra/Export.py).
 */
func Get_Table_Export_Xlsx(w http.ResponseWriter, r *http.Request) {
	collSRA := db.Collection("sra")
	if collSRA == nil {
		log.Println("sra.Export: collection 'sra' not found")
		http.Error(w, "Collection 'sra' not found", http.StatusInternalServerError)
		return
	}
	collCongr := db.Collection("congregations")
	if collCongr == nil {
		log.Println("sra.Export: collection 'congregations' not found")
		http.Error(w, "Collection 'congregations' not found", http.StatusInternalServerError)
		return
	}
	collPilots := db.Collection("pilots")
	if collPilots == nil {
		log.Println("sra.Export: collection 'pilots' not found")
		http.Error(w, "Collection 'pilots' not found", http.StatusInternalServerError)
		return
	}

	// tylko zgłoszenia z autokarem, najnowsze na górze (jak w tabeli admina)
	opts := options.Find().SetSort(bson.M{"timestamp": -1})
	cur, err := collSRA.Find(r.Context(), bson.M{"nobus": bson.M{"$exists": false}}, opts)
	if err != nil {
		log.Printf("sra.Export: error finding SRA: %v", err)
		http.Error(w, "Error reading registrations", http.StatusInternalServerError)
		return
	}
	defer cur.Close(r.Context())

	var sras []db.SRA
	if err := cur.All(r.Context(), &sras); err != nil {
		log.Printf("sra.Export: error decoding SRA: %v", err)
		http.Error(w, "Error reading registrations", http.StatusInternalServerError)
		return
	}

	const sheet = "Zgłoszenia"
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("sra.Export: error closing xlsx file: %v", err)
		}
	}()
	f.SetSheetName(f.GetSheetName(0), sheet)

	header := []interface{}{
		"#", "Data zgłoszenia",
		"Zbór-Język", "Zbór-Numer", "Zbór-Nazwa", "Zbór-Tura",
		"Bus-Lp", "Bus-Identyfikator-Letter", "Bus-Nadany-Identyfikator", "Bus-Typ", "Bus-Trasa", "Bus-Parking",
		"Pilot-Piątek-Imię", "Pilot-Piątek-Nazwisko", "Pilot-Piątek-Email", "Pilot-Piątek-Telefon",
		"Pilot-Sobota-Imię", "Pilot-Sobota-Nazwisko", "Pilot-Sobota-Email", "Pilot-Sobota-Telefon",
		"Pilot-Niedziela-Imię", "Pilot-Niedziela-Nazwisko", "Pilot-Niedziela-Email", "Pilot-Niedziela-Telefon",
		"Info",
	}
	if err := f.SetSheetRow(sheet, "A1", &header); err != nil {
		log.Printf("sra.Export: error writing header: %v", err)
		http.Error(w, "Error building spreadsheet", http.StatusInternalServerError)
		return
	}

	for i, sra := range sras {
		row := []interface{}{i + 1, sra.Timestamp.Time().Format("2006-01-02 15:04:05")}

		// dane zboru
		var congregation db.Congregation
		if err := collCongr.FindOne(r.Context(), bson.M{"_id": sra.CongregationID}).Decode(&congregation); err != nil {
			log.Printf("sra.Export: congregation '%s' not found: %v", sra.CongregationID.Hex(), err)
			http.Error(w, "Error reading congregation data", http.StatusInternalServerError)
			return
		}
		row = append(row, congregation.Lang, congregation.Number, congregation.Name, fmt.Sprintf("W%d", congregation.Tura))

		// dane autokaru (kolumna "Letter" nie jest już używana - puste pole)
		lp := ""
		if sra.Lp != nil {
			lp = fmt.Sprintf("%d", *sra.Lp)
		}
		staticIdent := ""
		if sra.StaticIdentifier != nil {
			staticIdent = *sra.StaticIdentifier
		}
		row = append(row,
			lp, "", staticIdent,
			exportBusType(sra.Bus.Type),
			exportBusDistance(sra.Bus.Distance),
			exportParkingMode(sra.Bus.ParkingMode),
		)

		// dane pilotów (piątek / sobota / niedziela)
		row = append(row, pilotCells(r.Context(), collPilots, &sra.Pilot1ID)...)
		row = append(row, pilotCells(r.Context(), collPilots, sra.Pilot2ID)...)
		row = append(row, pilotCells(r.Context(), collPilots, sra.Pilot3ID)...)

		// dodatkowe informacje
		info := ""
		if sra.Info != nil {
			info = *sra.Info
		}
		row = append(row, info)

		cell, _ := excelize.CoordinatesToCellName(1, i+2)
		if err := f.SetSheetRow(sheet, cell, &row); err != nil {
			log.Printf("sra.Export: error writing row %d: %v", i+2, err)
			http.Error(w, "Error building spreadsheet", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", `attachment; filename="Ankiety autokarow.xlsx"`)
	if err := f.Write(w); err != nil {
		log.Printf("sra.Export: error writing xlsx to response: %v", err)
	}
}

/**
*	Zwraca 4 komórki danych pilota (imię, nazwisko, email, telefon).
*	Dla brakującego pilota (nil) zwraca puste komórki.
 */
func pilotCells(ctx context.Context, coll *mongo.Collection, id *primitive.ObjectID) []interface{} {
	if id == nil {
		return []interface{}{"", "", "", ""}
	}
	var pilot db.Pilot
	if err := coll.FindOne(ctx, bson.M{"_id": *id}).Decode(&pilot); err != nil {
		log.Printf("sra.Export: pilot '%s' not found: %v", id.Hex(), err)
		return []interface{}{"", "", "", ""}
	}
	phone := pilot.Phone.CountryCode + " " + pilot.Phone.Number
	return []interface{}{pilot.FirstName, pilot.LastName, pilot.Email, phone}
}

func exportBusType(bt string) string {
	switch bt {
	case "minibus_9":
		return "minibus 9"
	case "minibus_30":
		return "minibus 30"
	case "autokar_50":
		return "autokar 50"
	case "autokar_70":
		return "autokar 60-70"
	case "autobus_12m":
		return "autobus 12m"
	case "autobus_18m":
		return "autobus 18m"
	}
	return bt
}

func exportBusDistance(v string) string {
	switch v {
	case "more200km":
		return "> 200km"
	}
	return v
}

func exportParkingMode(pm string) string {
	switch pm {
	case "needed":
		return "TAK"
	case "not_needed":
		return "NIE"
	}
	return pm
}
