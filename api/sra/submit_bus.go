package sra

import (
	"bytes"
	"encoding/json"
	"gokongres/db"
	"gokongres/mailer"
	"html/template"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// odbiór danych z żądania
type Phone struct {
	CountryCode string `json:"direct"`
	Number      string `json:"number"`
}
type Pilot struct {
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Email     string `json:"email"`
	Phone     Phone  `json:"phone"`
}
type RegistrationData struct {
	CongregationName string  `json:"congregation"`
	Bus              db.Bus  `json:"bus"`
	Info             *string `json:"info,omitempty"`
	OnePilot         bool    `json:"one_pilot"`
	Pilot            []Pilot `json:"pilot"`
}
type Request struct {
	ConfirmEmail string           `json:"confirmation_email"`
	Registration RegistrationData `json:"registration_data"`
}

func Post_SubmitBus(w http.ResponseWriter, r *http.Request) {
	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("Error decoding SRA submission request: %v", err)
		http.Error(w, "Error decoding request", http.StatusBadRequest)
		return
	}

	// budowanie obiektu SRA do zapisania w bazie danych
	sra := db.SRA{}

	// opcjonalne pole dodatkowej informacji
	if req.Registration.Info != nil {
		sra.Info = req.Registration.Info
	}

	// podłączenie zboru po nazwie
	var congregation db.Congregation
	err = db.Collection("congregations").FindOne(r.Context(), bson.M{"name": req.Registration.CongregationName}).Decode(&congregation)
	if err != nil {
		log.Printf("Error finding congregation: %v", err)
		http.Error(w, "Error finding congregation", http.StatusBadRequest)
		return
	}
	sra.CongregationID = congregation.ID

	// dane busa
	sra.Bus = req.Registration.Bus

	// dane pilotów
	res := db.Collection("pilots").FindOne(r.Context(), bson.M{"fn": req.Registration.Pilot[0].FirstName, "ln": req.Registration.Pilot[0].LastName, "email": req.Registration.Pilot[0].Email})
	if res.Err() != nil {
		// jeśli pilot nie istnieje, tworzymy nowy rekord w bazie danych
		newPilot := db.Pilot{
			FirstName: req.Registration.Pilot[0].FirstName,
			LastName:  req.Registration.Pilot[0].LastName,
			Email:     req.Registration.Pilot[0].Email,
			Phone: db.Phone{
				CountryCode: req.Registration.Pilot[0].Phone.CountryCode,
				Number:      req.Registration.Pilot[0].Phone.Number,
			},
		}
		result, err := db.Collection("pilots").InsertOne(r.Context(), newPilot)
		if err != nil {
			log.Printf("Error inserting new pilot into database: %v", err)
			http.Error(w, "Error saving pilot information", http.StatusInternalServerError)
			return
		}
		resultID := result.InsertedID.(primitive.ObjectID)
		sra.Pilot1ID = resultID
		log.Printf("Created new pilot with ID: %s", resultID.Hex())
	} else {
		var pilot db.Pilot
		err = res.Decode(&pilot)
		if err != nil {
			log.Printf("Error decoding pilot from database: %v", err)
			http.Error(w, "Error retrieving pilot information", http.StatusInternalServerError)
			return
		}
		sra.Pilot1ID = pilot.ID
		log.Printf("Found existing pilot with ID: %s", pilot.ID.Hex())
	}
	if !req.Registration.OnePilot {
		// pilot 2 jest opcjonalny, więc sprawdzamy, czy został podany
		res := db.Collection("pilots").FindOne(r.Context(), bson.M{"fn": req.Registration.Pilot[1].FirstName, "ln": req.Registration.Pilot[1].LastName, "email": req.Registration.Pilot[1].Email})
		if res.Err() != nil {
			// jeśli pilot nie istnieje, tworzymy nowy rekord w bazie danych
			newPilot := db.Pilot{
				FirstName: req.Registration.Pilot[1].FirstName,
				LastName:  req.Registration.Pilot[1].LastName,
				Email:     req.Registration.Pilot[1].Email,
				Phone: db.Phone{
					CountryCode: req.Registration.Pilot[1].Phone.CountryCode,
					Number:      req.Registration.Pilot[1].Phone.Number,
				},
			}
			result, err := db.Collection("pilots").InsertOne(r.Context(), newPilot)
			if err != nil {
				log.Printf("Error inserting new pilot 2 into database: %v", err)
				http.Error(w, "Error saving pilot 2 information", http.StatusInternalServerError)
				return
			}
			resultID := result.InsertedID.(primitive.ObjectID)
			sra.Pilot2ID = &resultID
			log.Printf("Created new pilot 2 with ID: %s", resultID.Hex())
		} else {
			var pilot db.Pilot
			err = res.Decode(&pilot)
			if err != nil {
				log.Printf("Error decoding pilot 2 from database: %v", err)
				http.Error(w, "Error retrieving pilot 2 information", http.StatusInternalServerError)
				return
			}
			sra.Pilot2ID = &pilot.ID
			log.Printf("Found existing pilot 2 with ID: %s", pilot.ID.Hex())
		}

		// pilot 3 jest opcjonalny, więc sprawdzamy, czy został podany
		res = db.Collection("pilots").FindOne(r.Context(), bson.M{"fn": req.Registration.Pilot[2].FirstName, "ln": req.Registration.Pilot[2].LastName, "email": req.Registration.Pilot[2].Email})
		if res.Err() != nil {
			// jeśli pilot nie istnieje, tworzymy nowy rekord w bazie danych
			newPilot := db.Pilot{
				FirstName: req.Registration.Pilot[2].FirstName,
				LastName:  req.Registration.Pilot[2].LastName,
				Email:     req.Registration.Pilot[2].Email,
				Phone: db.Phone{
					CountryCode: req.Registration.Pilot[2].Phone.CountryCode,
					Number:      req.Registration.Pilot[2].Phone.Number,
				},
			}
			result, err := db.Collection("pilots").InsertOne(r.Context(), newPilot)
			if err != nil {
				log.Printf("Error inserting new pilot 3 into database: %v", err)
				http.Error(w, "Error saving pilot 3 information", http.StatusInternalServerError)
				return
			}
			resultID := result.InsertedID.(primitive.ObjectID)
			sra.Pilot3ID = &resultID
			log.Printf("Created new pilot 3 with ID: %s", resultID.Hex())
		} else {
			var pilot db.Pilot
			err = res.Decode(&pilot)
			if err != nil {
				log.Printf("Error decoding pilot 3 from database: %v", err)
				http.Error(w, "Error retrieving pilot 3 information", http.StatusInternalServerError)
				return
			}
			sra.Pilot3ID = &pilot.ID
			log.Printf("Found existing pilot 3 with ID: %s", pilot.ID.Hex())
		}
	}

	// ustawienie timestampu na czas otrzymania zgłoszenia
	sra.Timestamp = primitive.NewDateTimeFromTime(time.Now())

	// zapis do bazy danych
	coll := db.Collection("sra")
	if coll == nil {
		log.Println("Collection 'sra' not found")
		http.Error(w, "Collection 'sra' not found", http.StatusInternalServerError)
		return
	}

	_, err = coll.InsertOne(r.Context(), sra)
	if err != nil {
		log.Printf("Error inserting SRA submission into database: %v", err)
		http.Error(w, "Error saving registration", http.StatusInternalServerError)
		return
	}

	// wysyłka maila z potwierdzeniem zgłoszenia
	recipients := []string{req.ConfirmEmail}
	body := prepareHtmlBody(req)
	mailer.SendHtmlMail(recipients, "Potwierdzenie zgłoszenia autokaru", body)
}

/**
*	Tworzenie body maila
 */
func prepareHtmlBody(req Request) string {
	if req.Registration.OnePilot {
		return onePilot(req)
	} else {
		// Obsługa przypadku z wieloma pilotami
		return morePilots(req)
	}
}

/**
*	Tworzenie body maila dla przypadku z jednym pilotem
 */
func onePilot(req Request) string {
	const tmpl = `
	<!DOCTYPE html>
	<html xmlns="http://www.w3.org/1999/xhtml">
	<head>
	<meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
	<meta name="viewport" content="width=device-width, initial-scale=1.0" />
	<title>Tytuł maila</title>
	<style type="text/css">
		/* Reset */
		body { margin: 0; padding: 0; min-width: 100%; }
		img  { border: 0; display: block; }
		/* Własne style — tylko podstawowe! */
	</style>
	</head>
	<body style="margin:0; padding:0; background-color:#f4f4f4;">
		<!-- Wrapper -->
		<table width="100%" cellpadding="0" cellspacing="0" border="0">
			<tr>
				<td align="center">

					<!-- Kontener treści (max 600px) -->
					<table width="600" cellpadding="0" cellspacing="0" border="0" style="background-color:#ffffff;">
						<!-- Nagłówek sekcji -->
						<tr>
							<td colspan="3" style="padding:20px; background-color:#333333; color:#ffffff; font-family:Arial,sans-serif; font-size:24px;">
								Dane pojazdu
							</td>
						</tr>
						<!-- Treść sekcji -->
						<tr>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								Nazwa zboru:
							</td>
							<td width="20" style="font-size:0; line-height:0;">&nbsp;</td>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								{{.Congregation}}
							</td>
						</tr>
						<tr>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								Typ pojazdu:
							</td>
							<td width="20" style="font-size:0; line-height:0;">&nbsp;</td>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								{{.VehicleType}}
							</td>
						</tr>
						<tr>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								Długość trasy:
							</td>
							<td width="20" style="font-size:0; line-height:0;">&nbsp;</td>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								{{.RouteLength}}
							</td>
						</tr>
						<tr>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								Parking:
							</td>
							<td width="20" style="font-size:0; line-height:0;">&nbsp;</td>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								{{.ParkingMode}}
							</td>
						</tr>

						<!-- Nagłówek sekcji -->
						<tr>
							<td colspan="3" style="padding:20px; background-color:#333333; color:#ffffff; font-family:Arial,sans-serif; font-size:24px;">
								Dane pilota
							</td>
						</tr>
						<!-- Treść sekcji -->
						<tr>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								Imię i nazwisko:
							</td>
							<td width="20" style="font-size:0; line-height:0;">&nbsp;</td>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								{{.Pilot}}
							</td>
						</tr>
						<tr>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								Numer telefonu:
							</td>
							<td width="20" style="font-size:0; line-height:0;">&nbsp;</td>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								{{.Phone}}
							</td>
						</tr>
						<tr>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								E-mail:
							</td>
							<td width="20" style="font-size:0; line-height:0;">&nbsp;</td>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								{{.Email}}
							</td>
						</tr>

						<!-- Nagłówek sekcji -->
						<tr>
							<td colspan="3" style="padding:20px; background-color:#333333; color:#ffffff; font-family:Arial,sans-serif; font-size:24px;">
								Uwagi i dodatkowe informacje
							</td>
						</tr>
						<!-- Treść sekcji -->
						<tr>
							<td  colspan="3" valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								{{.Info}}
							</td>
						</tr>

						<tr>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								Data wysłania zgłoszenia:
							</td>
							<td width="20" style="font-size:0; line-height:0;">&nbsp;</td>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								{{.Timestamp}}
							</td>
						</tr>
					</table>

				</td>
			</tr>
		</table>
	</body>
	</html>
	`

	t := template.Must(template.New("mail").Parse(tmpl))

	var buf bytes.Buffer
	t.Execute(&buf, map[string]string{
		"Congregation": req.Registration.CongregationName,
		"VehicleType":  busTypeAsString(req.Registration.Bus.Type),
		"RouteLength":  routeLengthAsString(req.Registration.Bus.Distance),
		"ParkingMode":  parkingModeAsString(req.Registration.Bus.ParkingMode),
		"Pilot":        req.Registration.Pilot[0].FirstName + " " + req.Registration.Pilot[0].LastName,
		"Phone":        req.Registration.Pilot[0].Phone.CountryCode + " " + req.Registration.Pilot[0].Phone.Number,
		"Email":        req.Registration.Pilot[0].Email,
		"Info":         getInfo(req),
		"Timestamp":    time.Now().Format("2006-01-02 15:04:05"),
	})

	return buf.String()
}

/**
*	Tworzenie body maila dla przypadku z wieloma pilotami
 */
func morePilots(req Request) string {
	const tmpl = `
	<!DOCTYPE html>
	<html xmlns="http://www.w3.org/1999/xhtml">
	<head>
	<meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
	<meta name="viewport" content="width=device-width, initial-scale=1.0" />
	<title>Tytuł maila</title>
	<style type="text/css">
		/* Reset */
		body { margin: 0; padding: 0; min-width: 100%; }
		img  { border: 0; display: block; }
		/* Własne style — tylko podstawowe! */
	</style>
	</head>
	<body style="margin:0; padding:0; background-color:#f4f4f4;">
		<!-- Wrapper -->
		<table width="100%" cellpadding="0" cellspacing="0" border="0">
			<tr>
				<td align="center">

					<!-- Kontener treści (max 600px) -->
					<table width="600" cellpadding="0" cellspacing="0" border="0" style="background-color:#ffffff;">
						<!-- Nagłówek sekcji -->
						<tr>
							<td colspan="3" style="padding:20px; background-color:#333333; color:#ffffff; font-family:Arial,sans-serif; font-size:24px;">
								Dane pojazdu
							</td>
						</tr>
						<!-- Treść sekcji -->
						<tr>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								Nazwa zboru:
							</td>
							<td width="20" style="font-size:0; line-height:0;">&nbsp;</td>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								{{.Congregation}}
							</td>
						</tr>
						<tr>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								Typ pojazdu:
							</td>
							<td width="20" style="font-size:0; line-height:0;">&nbsp;</td>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								{{.VehicleType}}
							</td>
						</tr>
						<tr>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								Długość trasy:
							</td>
							<td width="20" style="font-size:0; line-height:0;">&nbsp;</td>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								{{.RouteLength}}
							</td>
						</tr>
						<tr>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								Parking:
							</td>
							<td width="20" style="font-size:0; line-height:0;">&nbsp;</td>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								{{.ParkingMode}}
							</td>
						</tr>

						<!-- Nagłówek sekcji -->
						<tr>
							<td colspan="3" style="padding:20px; background-color:#333333; color:#ffffff; font-family:Arial,sans-serif; font-size:24px;">
								Dane pilota
							</td>
						</tr>
						<!-- Treść sekcji -->
						<tr>
							<td colspan="3" width="100%" valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:16px; color:#333333;">
								Piątek
							</td>
						</tr>
						<tr>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								Imię i nazwisko:
							</td>
							<td width="20" style="font-size:0; line-height:0;">&nbsp;</td>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								{{.Pilot1}}
							</td>
						</tr>
						<tr>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								Numer telefonu:
							</td>
							<td width="20" style="font-size:0; line-height:0;">&nbsp;</td>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								{{.Phone1}}
							</td>
						</tr>
						<tr>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								E-mail:
							</td>
							<td width="20" style="font-size:0; line-height:0;">&nbsp;</td>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								{{.Email1}}
							</td>
						</tr>

						<tr>
							<td colspan="3" width="100%" valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:16px; color:#333333;">
								Sobota
							</td>
						</tr>
						<tr>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								Imię i nazwisko:
							</td>
							<td width="20" style="font-size:0; line-height:0;">&nbsp;</td>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								{{.Pilot2}}
							</td>
						</tr>
						<tr>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								Numer telefonu:
							</td>
							<td width="20" style="font-size:0; line-height:0;">&nbsp;</td>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								{{.Phone2}}
							</td>
						</tr>
						<tr>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								E-mail:
							</td>
							<td width="20" style="font-size:0; line-height:0;">&nbsp;</td>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								{{.Email2}}
							</td>
						</tr>

						<tr>
							<td colspan="3" width="100%" valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:16px; color:#333333;">
								Niedziela
							</td>
						</tr>
						<tr>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								Imię i nazwisko:
							</td>
							<td width="20" style="font-size:0; line-height:0;">&nbsp;</td>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								{{.Pilot3}}
							</td>
						</tr>
						<tr>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								Numer telefonu:
							</td>
							<td width="20" style="font-size:0; line-height:0;">&nbsp;</td>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								{{.Phone3}}
							</td>
						</tr>
						<tr>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								E-mail:
							</td>
							<td width="20" style="font-size:0; line-height:0;">&nbsp;</td>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								{{.Email3}}
							</td>
						</tr>

						<!-- Nagłówek sekcji -->
						<tr>
							<td colspan="3" style="padding:20px; background-color:#333333; color:#ffffff; font-family:Arial,sans-serif; font-size:24px;">
								Uwagi i dodatkowe informacje
							</td>
						</tr>
						<!-- Treść sekcji -->
						<tr>
							<td colspan="3" valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								{{.Info}}
							</td>
						</tr>

						<tr>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								Data wysłania zgłoszenia:
							</td>
							<td width="20" style="font-size:0; line-height:0;">&nbsp;</td>
							<td valign="top" style="padding:10px; font-family:Arial,sans-serif; font-size:14px; color:#333333;">
								{{.Timestamp}}
							</td>
						</tr>
					</table>

				</td>
			</tr>
		</table>
	</body>
	</html>
	`

	t := template.Must(template.New("mail").Parse(tmpl))

	var buf bytes.Buffer
	t.Execute(&buf, map[string]string{
		"Congregation": req.Registration.CongregationName,
		"VehicleType":  busTypeAsString(req.Registration.Bus.Type),
		"RouteLength":  routeLengthAsString(req.Registration.Bus.Distance),
		"ParkingMode":  parkingModeAsString(req.Registration.Bus.ParkingMode),
		"Pilot1":       req.Registration.Pilot[0].FirstName + " " + req.Registration.Pilot[0].LastName,
		"Phone1":       req.Registration.Pilot[0].Phone.CountryCode + " " + req.Registration.Pilot[0].Phone.Number,
		"Email1":       req.Registration.Pilot[0].Email,
		"Pilot2":       req.Registration.Pilot[1].FirstName + " " + req.Registration.Pilot[1].LastName,
		"Phone2":       req.Registration.Pilot[1].Phone.CountryCode + " " + req.Registration.Pilot[1].Phone.Number,
		"Email2":       req.Registration.Pilot[1].Email,
		"Pilot3":       req.Registration.Pilot[2].FirstName + " " + req.Registration.Pilot[2].LastName,
		"Phone3":       req.Registration.Pilot[2].Phone.CountryCode + " " + req.Registration.Pilot[2].Phone.Number,
		"Email3":       req.Registration.Pilot[2].Email,
		"Info":         getInfo(req),
		"Timestamp":    time.Now().Format("2006-01-02 15:04:05"),
	})

	return buf.String()
}

func busTypeAsString(t string) string {
	switch t {
	default:
		return "Nieznany typ pojazdu"
	case "minibus_9":
		return "minibus do 9 osób"
	case "minibus_30":
		return "minibus do 30 osób"
	case "autokar_50":
		return "autokar do 50 osób"
	case "autokar_70":
		return "autokar 60-70 osób"
	case "autobus_12m":
		return "autobus miejski - 12m (pojedyńczy)"
	case "autobus_18m":
		return "autobus miejski - 18m (przegubowy)"
	}
}

func routeLengthAsString(rl string) string {
	switch rl {
	default:
		return "Nieznana długość trasy"
	case "15km":
		return "do 15 km"
	case "25km":
		return "do 25 km"
	case "50km":
		return "do 50 km"
	case "100km":
		return "do 100 km"
	case "200km":
		return "do 200 km"
	case "more200km":
		return "powyżej 200 km"
	}
}

func parkingModeAsString(mode string) string {
	switch mode {
	default:
		return "Nieznany"
	case "needed":
		return "potrzebne miejsce parkingowe"
	case "not_needed":
		return "pojazd tylko dowozi pasażerów, odjeżdza i przyjeżdza odebrać ich po programie"
	}
}

func getInfo(req Request) string {
	if req.Registration.Info != nil {
		return *req.Registration.Info
	}
	return ""
}
