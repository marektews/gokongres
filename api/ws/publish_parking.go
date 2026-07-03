package ws

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PublishParking rozgłasza zmianę zajętości identyfikatora parkingowego do
// ekranów monitoringu zapisanych na temat parking:<parking>.
// parking = "srp" (parking pod trybuną) albo "pk" (parking księżycowy działów).
// used=true przy wjeździe (check), false przy wyjeździe (free).
func PublishParking(parking string, id primitive.ObjectID, passNr int, used bool, ts time.Time) {
	payload := mustJSON(map[string]any{
		"type":    "parking",
		"parking": parking,
		"id":      id.Hex(),
		"pass_nr": passNr,
		"used":    used,
		"ts":      ts.Format("02.01.2006 15:04:05"),
	})

	Default.Publish("parking:"+parking, payload)
}
