package ws

import (
	"context"
	"log"
	"time"

	"gokongres/db"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PublishState rozgłasza zmianę stanu autokaru do ekranów obserwujących jego sektor.
// Sektor jest ustalany na podstawie dokumentu RJA. Pola rja_id/status/ts są celowo
// zgodne z odpowiedzią endpointów /notify/*, dzięki czemu front podaje je wprost do
// istniejącego apply_notification_response.
func PublishState(ctx context.Context, rjaID primitive.ObjectID, status string, ts time.Time) {
	coll := db.Collection("rja")
	if coll == nil {
		log.Println("ws.PublishState: collection 'rja' not found")
		return
	}

	var rja db.RJA
	if err := coll.FindOne(ctx, bson.M{"_id": rjaID}).Decode(&rja); err != nil {
		log.Printf("ws.PublishState: cannot find RJA '%s': %v", rjaID.Hex(), err)
		return
	}

	sid := rja.SectorID.Hex()
	payload := mustJSON(map[string]string{
		"type":   "state",
		"rja_id": rjaID.Hex(),
		"status": status,
		"ts":     ts.Format("02.01.2006 15:04:05"),
		"sid":    sid,
	})

	Default.Publish("sector:"+sid, payload)
}

// sectorIDsOfTerminal zwraca hex-id sektorów należących do terminala o danej nazwie.
// Używane przy subskrypcji buffer:<name> do rozwinięcia na tematy sector:<sid>.
func sectorIDsOfTerminal(name string) []string {
	coll := db.Collection("terminals")
	if coll == nil {
		log.Println("ws.sectorIDsOfTerminal: collection 'terminals' not found")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var term db.Terminal
	if err := coll.FindOne(ctx, bson.M{"name": name}).Decode(&term); err != nil {
		log.Printf("ws.sectorIDsOfTerminal: cannot find terminal '%s': %v", name, err)
		return nil
	}

	out := make([]string, 0, len(term.Sectors))
	for _, s := range term.Sectors {
		out = append(out, s.Sid.Hex())
	}
	return out
}
