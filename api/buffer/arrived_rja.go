package buffer

import (
	"context"
	"fmt"
	"gokongres/db"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ArrivedRJA: RJA przyjeżdżający dziś wraz z sektorem, do którego należy.
type ArrivedRJA struct {
	RJA    db.RJA
	Sector db.Sector
}

// arrivedRJAs zwraca wszystkie RJA przypisane do sektorów bufora (terminal.Sectors),
// które mają godzinę przyjazdu na aktywny dzień (WasArrived), posortowane po sector_order.
//
// Powiązanie bufor → autobusy prowadzi przez osadzone terminal.Sectors[].Sid → rja.sector_id
// (RJA nie posiada pola terminal_id).
func arrivedRJAs(ctx context.Context, terminal db.Terminal) ([]ArrivedRJA, error) {
	collRJA := db.Collection("rja")
	if collRJA == nil {
		return nil, fmt.Errorf("collection 'rja' not found")
	}

	collation := options.Collation{Locale: "pl", NumericOrdering: true, Strength: 1}
	opts := options.Find().SetSort(bson.D{{Key: "sector_order", Value: 1}}).SetCollation(&collation)

	result := make([]ArrivedRJA, 0)
	for _, sector := range terminal.Sectors {
		cur, err := collRJA.Find(ctx, bson.M{"sector_id": sector.Sid}, opts)
		if err != nil {
			return nil, err
		}

		var rjas []db.RJA
		if err := cur.All(ctx, &rjas); err != nil {
			return nil, err
		}

		for _, rja := range rjas {
			if rja.WasArrived() {
				result = append(result, ArrivedRJA{RJA: rja, Sector: sector})
			}
		}
	}

	return result, nil
}
