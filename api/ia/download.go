package ia

import (
	"fmt"
	"gokongres/db"
	"gokongres/qr"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/antchfx/xmlquery"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Get_Download(w http.ResponseWriter, r *http.Request) {
	sra_id := r.PathValue("sra_id")
	sraID, err := primitive.ObjectIDFromHex(sra_id)
	if err != nil {
		log.Printf("IA download: invalid sra_id '%s': %v", sra_id, err)
		http.Error(w, "Invalid sra_id format", http.StatusBadRequest)
		return
	}

	var sra db.SRA
	if err := db.Collection("sra").FindOne(r.Context(), bson.M{"_id": sraID}).Decode(&sra); err != nil {
		log.Printf("IA download: SRA '%s' not found: %v", sra_id, err)
		http.Error(w, "SRA not found", http.StatusNotFound)
		return
	}

	var cong db.Congregation
	if err := db.Collection("congregations").FindOne(r.Context(), bson.M{"_id": sra.CongregationID}).Decode(&cong); err != nil {
		log.Printf("IA download: congregation '%s' not found: %v", sra.CongregationID.Hex(), err)
		http.Error(w, "Congregation not found", http.StatusInternalServerError)
		return
	}

	var rja db.RJA
	if err := db.Collection("rja").FindOne(r.Context(), bson.M{"sra_id": sraID}).Decode(&rja); err != nil {
		log.Printf("IA download: RJA for sra '%s' not found: %v", sra_id, err)
		http.Error(w, "RJA not found for this bus", http.StatusNotFound)
		return
	}

	// terminal + sektor (sektory osadzone w dokumencie terminala)
	var terminal db.Terminal
	if err := db.Collection("terminals").FindOne(r.Context(), bson.M{"sectors.sid": rja.SectorID}).Decode(&terminal); err != nil {
		log.Printf("IA download: terminal for sector '%s' not found: %v", rja.SectorID.Hex(), err)
		http.Error(w, "Terminal not found", http.StatusInternalServerError)
		return
	}
	sectorName := ""
	for _, s := range terminal.Sectors {
		if s.Sid == rja.SectorID {
			sectorName = s.Name
			break
		}
	}

	// wczytanie i wypełnienie szablonu SVG
	xmlData, err := os.ReadFile("templates/identyfikator-autokaru-template.svg")
	if err != nil {
		log.Printf("IA download: error reading template: %v", err)
		http.Error(w, "Error reading template", http.StatusInternalServerError)
		return
	}
	doc, err := xmlquery.Parse(strings.NewReader(string(xmlData)))
	if err != nil {
		log.Printf("IA download: error parsing template: %v", err)
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}

	congName := cong.Name
	if sra.Lp != nil {
		congName = fmt.Sprintf("%s %d", cong.Name, *sra.Lp)
	}

	setText(doc, "terminal", terminal.Name)
	setText(doc, "sektor", digits(sectorName))
	setText(doc, "identyfikator", db.CreateShortBusID(&sra, sectorName, rja.SectorOrder))
	setText(doc, "congregation", congName)
	setText(doc, "arrive1", deref(rja.A1))
	setText(doc, "arrive2", deref(rja.A2))
	setText(doc, "arrive3", deref(rja.A3))
	setText(doc, "departure1", deref(rja.D1))
	setText(doc, "departure2", deref(rja.D2))
	setText(doc, "departure3", deref(rja.D3))

	pdfBytes, err := qr.SVGToPDF(doc.OutputXML(true))
	if err != nil {
		log.Printf("IA download: error converting SVG to PDF: %v", err)
		http.Error(w, "Error converting SVG to PDF", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\"identyfikator-autokaru.pdf\"")
	if _, err := w.Write(pdfBytes); err != nil {
		log.Printf("IA download: error writing PDF response: %v", err)
		return
	}
}

// setText ustawia treść tekstową węzła SVG o podanym id (jeśli istnieje).
func setText(doc *xmlquery.Node, id, value string) {
	n := xmlquery.FindOne(doc, ".//*[@id='"+id+"']")
	if n != nil && n.FirstChild != nil {
		n.FirstChild.Data = value
	}
}

// digits zwraca same cyfry z napisu (np. "T10" → "10").
func digits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func deref(p *string) string {
	if p != nil {
		return *p
	}
	return ""
}
