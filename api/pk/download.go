package pk

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
	pk_id := r.PathValue("pk_id")
	pkID, err := primitive.ObjectIDFromHex(pk_id)
	if err != nil {
		log.Println("Invalid pk_id format:", err)
		http.Error(w, "Invalid pk_id format", http.StatusBadRequest)
		return
	}

	collDepsPK := db.Collection("departments_pk")
	if collDepsPK == nil {
		log.Println("Collection 'departments_pk' not found")
		http.Error(w, "Collection 'departments_pk' not found", http.StatusInternalServerError)
		return
	}

	collDepartments := db.Collection("departments")
	if collDepartments == nil {
		log.Println("Collection 'departments' not found")
		http.Error(w, "Collection 'departments' not found", http.StatusInternalServerError)
		return
	}

	var pk db.DepartmentPK
	err = collDepsPK.FindOne(r.Context(), bson.M{"_id": pkID}).Decode(&pk)
	if err != nil {
		log.Println("Error finding pk entry:", err)
		http.Error(w, "Error finding pk entry", http.StatusInternalServerError)
		return
	}

	var department db.Department
	err = collDepartments.FindOne(r.Context(), bson.M{"_id": pk.DepartmentID}).Decode(&department)
	if err != nil {
		log.Println("Error finding department entry:", err)
		http.Error(w, "Error finding department entry", http.StatusInternalServerError)
		return
	}

	qrSrcData := fmt.Sprintf("pk-%d-%s", pk.PassNr, pk.RegNum1)
	if pk.RegNum2 != nil && *pk.RegNum2 != "" {
		qrSrcData += fmt.Sprintf("-%s", *pk.RegNum2)
	} else {
		qrSrcData += fmt.Sprintf("-%s", pk.RegNum1)
	}
	if pk.RegNum3 != nil && *pk.RegNum3 != "" {
		qrSrcData += fmt.Sprintf("-%s", *pk.RegNum3)
	} else {
		qrSrcData += fmt.Sprintf("-%s", pk.RegNum1)
	}

	svgQRCode, err := qr.GenQRCode(qrSrcData, 10)
	if err != nil {
		log.Println("Error generating QR code:", err)
		http.Error(w, "Error generating QR code", http.StatusInternalServerError)
		return
	}

	var svg string
	if pk.RegNum2 != nil && *pk.RegNum2 != "" && pk.RegNum3 != nil && *pk.RegNum3 != "" {
		svg, err = gen_3(pk, department, svgQRCode)
		if err != nil {
			log.Println("Error generating SVG:", err)
			http.Error(w, "Error generating SVG", http.StatusInternalServerError)
			return
		}
	} else {
		svg, err = gen_1(pk, department, svgQRCode)
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
 * Funkcja generująca ostateczny SVG identyfikatora parkingowego
 * Przyjmuje dane SRP, dane zboru oraz ciąg SVG kodu QR
 * Wczytuje szablon SVG z pliku, modyfikuje go, osadzając odpowiednie dane i kod QR
 * Zwraca wygenerowany SVG jako string lub błąd, jeśli wystąpi problem podczas generowania
 */
func gen_1(pk db.DepartmentPK, department db.Department, qrCode string) (string, error) {
	// wczytanie szablonu XML z pliku
	xmlData, err := os.ReadFile("templates/parking-dzialy-pass-id-template-1.svg")
	if err != nil {
		return "", fmt.Errorf("error reading XML template: %v", err)
	}

	doc, err := xmlquery.Parse(strings.NewReader(string(xmlData)))
	if err != nil {
		return "", fmt.Errorf("error parsing XML template: %v", err)
	}

	// nr identyfikatora
	xmlPassID := xmlquery.FindOne(doc, ".//*[@id='numerIdentyfikatora']")
	if xmlPassID != nil {
		xmlPassID.FirstChild.Data = fmt.Sprintf("%d", pk.PassNr)
	} else {
		return "", fmt.Errorf("error finding XML node for pass number")
	}

	// nazwa zboru
	xmlCongregationName := xmlquery.FindOne(doc, ".//*[@id='department']")
	if xmlCongregationName != nil {
		xmlCongregationName.FirstChild.Data = department.Name
	} else {
		return "", fmt.Errorf("error finding XML node for department name")
	}

	// nr rejestracyjny na 3 dni kongresowe
	xmlRejNum := xmlquery.FindOne(doc, ".//*[@id='rejnum']")
	if xmlRejNum != nil {
		xmlRejNum.FirstChild.Data = pk.RegNum1
	} else {
		return "", fmt.Errorf("error finding XML node for registration number")
	}

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
 */
func gen_3(pk db.DepartmentPK, department db.Department, qrCode string) (string, error) {
	// wczytanie szablonu XML z pliku
	xmlData, err := os.ReadFile("templates/parking-dzialy-pass-id-template-3.svg")
	if err != nil {
		return "", fmt.Errorf("error reading XML template: %v", err)
	}

	doc, err := xmlquery.Parse(strings.NewReader(string(xmlData)))
	if err != nil {
		return "", fmt.Errorf("error parsing XML template: %v", err)
	}

	// nr identyfikatora
	xmlPassID := xmlquery.FindOne(doc, ".//*[@id='numerIdentyfikatora']")
	if xmlPassID != nil {
		xmlPassID.FirstChild.Data = fmt.Sprintf("%d", pk.PassNr)
	} else {
		return "", fmt.Errorf("error finding XML node for pass number")
	}

	// nazwa zboru
	xmlCongregationName := xmlquery.FindOne(doc, ".//*[@id='department']")
	if xmlCongregationName != nil {
		xmlCongregationName.FirstChild.Data = department.Name
	} else {
		return "", fmt.Errorf("error finding XML node for department name")
	}

	// nr rejestracyjny na poszczególne dni kongresowe
	xmlRejNum := xmlquery.FindOne(doc, ".//*[@id='d1rejnum']")
	if xmlRejNum != nil {
		xmlRejNum.FirstChild.Data = pk.RegNum1
	} else {
		return "", fmt.Errorf("error finding XML node for registration number 1")
	}
	xmlRejNum = xmlquery.FindOne(doc, ".//*[@id='d2rejnum']")
	if xmlRejNum != nil {
		xmlRejNum.FirstChild.Data = *pk.RegNum2
	} else {
		return "", fmt.Errorf("error finding XML node for registration number 2")
	}
	xmlRejNum = xmlquery.FindOne(doc, ".//*[@id='d3rejnum']")
	if xmlRejNum != nil {
		xmlRejNum.FirstChild.Data = *pk.RegNum3
	} else {
		return "", fmt.Errorf("error finding XML node for registration number 3")
	}

	// osadzanie qrcode
	err = embededQRCode(doc, qrCode)
	if err != nil {
		return "", fmt.Errorf("error embedding QR code: %v", err)
	}

	// konwersja dokumentu XML z powrotem na string
	outputXML := doc.OutputXML(true)
	return outputXML, nil
}
