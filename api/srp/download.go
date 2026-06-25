package srp

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

func Get_DownloadPassData(w http.ResponseWriter, r *http.Request) {
	collSRP := db.Collection("srp")
	if collSRP == nil {
		log.Println("Collection 'srp' not found")
		http.Error(w, "Collection 'srp' not found", http.StatusInternalServerError)
		return
	}

	collCongr := db.Collection("congregations")
	if collCongr == nil {
		log.Println("Collection 'congregations' not found")
		http.Error(w, "Collection 'congregations' not found", http.StatusInternalServerError)
		return
	}

	passID := r.PathValue("pass_id")
	passIdent, err := primitive.ObjectIDFromHex(passID)
	if err != nil {
		log.Printf("Invalid pass ID format: %v", err)
		http.Error(w, "Invalid pass ID format", http.StatusBadRequest)
		return
	}

	var srp db.SRP
	err = collSRP.FindOne(r.Context(), bson.M{"_id": passIdent}).Decode(&srp)
	if err != nil {
		log.Println("Error finding SRP document:", err)
		http.Error(w, "Error finding SRP document", http.StatusInternalServerError)
		return
	}

	var congregation db.Congregation
	err = collCongr.FindOne(r.Context(), bson.M{"_id": srp.CongregationID}).Decode(&congregation)
	if err != nil {
		log.Println("Error finding congregation document:", err)
		http.Error(w, "Error finding congregation document", http.StatusInternalServerError)
		return
	}

	qrSrcData := fmt.Sprintf("%d-%s", srp.PassNr, srp.Car1.RegNum)
	if srp.Car2 != nil && srp.Car2.RegNum != "" {
		qrSrcData += fmt.Sprintf("-%s", srp.Car2.RegNum)
	} else {
		qrSrcData += fmt.Sprintf("-%s", srp.Car1.RegNum)
	}
	if srp.Car3 != nil && srp.Car3.RegNum != "" {
		qrSrcData += fmt.Sprintf("-%s", srp.Car3.RegNum)
	} else {
		qrSrcData += fmt.Sprintf("-%s", srp.Car1.RegNum)
	}

	svgQRCode, err := qr.GenQRCode(qrSrcData, 10)
	if err != nil {
		log.Println("Error generating QR code:", err)
		http.Error(w, "Error generating QR code", http.StatusInternalServerError)
		return
	}

	hasCar2 := srp.Car2 != nil && srp.Car2.RegNum != ""
	hasCar3 := srp.Car3 != nil && srp.Car3.RegNum != ""

	// "ten sam pojazd przez 3 dni" także wtedy, gdy wszystkie pola są identyczne
	allSameRegNum := hasCar2 && hasCar3 &&
		srp.Car1.RegNum == srp.Car2.RegNum &&
		srp.Car1.RegNum == srp.Car3.RegNum

	var svg string
	if hasCar2 && hasCar3 && !allSameRegNum {
		svg, err = gen_3(srp, congregation, svgQRCode)
		if err != nil {
			log.Println("Error generating SVG:", err)
			http.Error(w, "Error generating SVG", http.StatusInternalServerError)
			return
		}
	} else {
		svg, err = gen_1(srp, congregation, svgQRCode)
		if err != nil {
			log.Println("Error generating SVG:", err)
			http.Error(w, "Error generating SVG", http.StatusInternalServerError)
			return
		}
	}

	// wysyłanie SVG
	// w.Header().Set("Content-Type", "image/svg+xml")
	// w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"identyfikator-parkingowy-%d.svg\"", srp.PassNr))
	// _, err = w.Write([]byte(svg))
	// if err != nil {
	// 	log.Println("Error writing SVG response:", err)
	// 	return
	// }

	// Wysyłanie PDF'a
	pdfBytes, err := qr.SVGToPDF(svg)
	if err != nil {
		log.Println("Error converting SVG to PDF:", err)
		http.Error(w, "Error converting SVG to PDF", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\"identyfikator-parkingowy.pdf\"")
	_, err = w.Write(pdfBytes)
	if err != nil {
		log.Println("Error writing PDF response:", err)
		return
	}
}

/**
 * Funkcja osadzająca kod QR w szablonie SVG
 * Przyjmuje węzeł XML reprezentujący szablon SVG oraz ciąg SVG kodu QR
 * Modyfikuje węzeł XML, osadzając kod QR w odpowiednim miejscu
 * Zwraca błąd, jeśli wystąpi problem z osadzeniem kodu QR
 */
func embededQRCode(svgPass *xmlquery.Node, svgQRCode string) error {
	// Znajdź rect#qrCode
	rectNode, err := xmlquery.Query(svgPass, `//*[@id="qrCode"]`)
	if err != nil || rectNode == nil {
		return fmt.Errorf("nie znaleziono elementu id=\"qrCode\": %w", err)
	}

	// Wyciągnij atrybuty pozycji i rozmiaru z rect
	attr := func(name string) string {
		if a := rectNode.SelectAttr(name); a != "" {
			return a
		}
		return "0"
	}
	x, y, width, height := attr("x"), attr("y"), attr("width"), attr("height")

	// Parsuj SVG z kodem QR
	qrDoc, err := xmlquery.Parse(strings.NewReader(svgQRCode))
	if err != nil {
		return fmt.Errorf("błąd parsowania svgQRCode: %w", err)
	}

	// Pobierz korzeń <svg> z QR kodu
	qrRoot := xmlquery.FindOne(qrDoc, "//svg")
	if qrRoot == nil {
		return fmt.Errorf("nie znaleziono elementu <svg> w svgQRCode")
	}

	// Ustaw atrybuty pozycji na qrRoot <svg>
	// żeby dopasować go do miejsca po rect
	newAttrs := [][2]string{
		{"x", x}, {"y", y}, {"width", width}, {"height", height},
	}
	for _, pair := range newAttrs {
		if qrRoot.HasAttr(pair[0]) {
			qrRoot.SetAttr(pair[0], pair[1])
		} else {
			xmlquery.AddAttr(qrRoot, pair[0], pair[1])
		}
	}

	// Zastąp rectNode węzłem <svg> z QR kodem
	xmlquery.AddChild(rectNode.Parent, qrRoot)

	removeNode := func(n *xmlquery.Node) {
		if n.Parent == nil {
			return
		}
		if n.PrevSibling != nil {
			n.PrevSibling.NextSibling = n.NextSibling
		} else {
			// n był pierwszym dzieckiem
			n.Parent.FirstChild = n.NextSibling
		}
		if n.NextSibling != nil {
			n.NextSibling.PrevSibling = n.PrevSibling
		} else {
			// n był ostatnim dzieckiem
			n.Parent.LastChild = n.PrevSibling
		}
		n.Parent = nil
		n.PrevSibling = nil
		n.NextSibling = nil
	}
	removeNode(rectNode)
	return nil
}

/**
 * Funkcja generująca ostateczny SVG identyfikatora parkingowego (ten sam samochód 3 dni)
 * Przyjmuje dane SRP, dane zboru oraz ciąg SVG kodu QR
 * Wczytuje szablon SVG z pliku, modyfikuje go, osadzając odpowiednie dane i kod QR
 * Zwraca wygenerowany SVG jako string lub błąd, jeśli wystąpi problem podczas generowania
 */
func gen_1(srp db.SRP, congregation db.Congregation, qrCode string) (string, error) {
	// wczytanie szablonu XML z pliku
	xmlData, err := os.ReadFile("templates/parking-lazienkowska-pass-id-template-1.svg")
	if err != nil {
		return "", fmt.Errorf("error reading XML template: %v", err)
	}

	doc, err := xmlquery.Parse(strings.NewReader(string(xmlData)))
	if err != nil {
		return "", fmt.Errorf("error parsing XML template: %v", err)
	}

	setText(doc, "numerIdentyfikatora", fmt.Sprintf("%d", srp.PassNr))
	setText(doc, "nazwaZboru", congregation.Name)
	setText(doc, "parking", parking(srp))
	setText(doc, "rejnum", srp.Car1.RegNum)

	// osadzanie qrcode
	err = embededQRCode(doc, qrCode)
	if err != nil {
		return "", fmt.Errorf("error embedding QR code: %v", err)
	}

	// konwersja dokumentu XML z powrotem na string
	outputXML := doc.OutputXML(true)
	return outputXML, nil
}

/**
 * Funkcja generująca ostateczny SVG identyfikatora parkingowego (dla 2 lub 3 samochodów)
 * Przyjmuje dane SRP, dane zboru oraz ciąg SVG kodu QR
 * Wczytuje szablon SVG z pliku, modyfikuje go, osadzając odpowiednie dane i kod QR
 * Zwraca wygenerowany SVG jako string lub błąd, jeśli wystąpi problem podczas generowania
 */
func gen_3(srp db.SRP, congregation db.Congregation, qrCode string) (string, error) {
	// wczytanie szablonu XML z pliku
	xmlData, err := os.ReadFile("templates/parking-lazienkowska-pass-id-template-3.svg")
	if err != nil {
		return "", fmt.Errorf("error reading XML template: %v", err)
	}

	doc, err := xmlquery.Parse(strings.NewReader(string(xmlData)))
	if err != nil {
		return "", fmt.Errorf("error parsing XML template: %v", err)
	}

	// nr identyfikatora
	setText(doc, "numerIdentyfikatora", fmt.Sprintf("%d", srp.PassNr))
	setText(doc, "nazwaZboru", congregation.Name)
	setText(doc, "parking", parking(srp))
	setText(doc, "d1rejnum", srp.Car1.RegNum)
	setText(doc, "d2rejnum", deref(srp.Car2))
	setText(doc, "d3rejnum", deref(srp.Car3))

	// osadzanie qrcode
	err = embededQRCode(doc, qrCode)
	if err != nil {
		return "", fmt.Errorf("error embedding QR code: %v", err)
	}

	// konwersja dokumentu XML z powrotem na string
	outputXML := doc.OutputXML(true)
	return outputXML, nil
}

// setText ustawia treść tekstową węzła SVG o podanym id (jeśli istnieje).
func setText(doc *xmlquery.Node, id, value string) {
	n := xmlquery.FindOne(doc, ".//*[@id='"+id+"']")
	if n != nil && n.FirstChild != nil {
		n.FirstChild.Data = value
	}
}

func deref(p *db.CarInfo) string {
	if p != nil {
		return p.RegNum
	}
	return ""
}

func parking(srp db.SRP) string {
	if srp.MobilityRestrictions || hasLpg(srp) {
		return "Torwar"
	}
	return "Stadion"
}

func hasLpg(srp db.SRP) bool {
	return srp.Car1.Lpg ||
		(srp.Car2 != nil && srp.Car2.Lpg) ||
		(srp.Car3 != nil && srp.Car3.Lpg)
}
