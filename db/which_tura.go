package db

import (
	"context"
	"log"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func WhichTura(ctx context.Context) *Tura {
	if ctx == nil {
		ctx = context.Background()
	}

	coll := Collection("const")
	if coll == nil {
		log.Print("Collection 'const' not initialized")
		return nil
	}

	now := time.Now()

	var cfg ConstConfig
	if err := coll.FindOne(ctx, bson.M{}).Decode(&cfg); err != nil {
		log.Printf("whichTuraFirstInOrder: %v", err)
		return nil
	}
	if len(cfg.Tury) == 0 {
		return nil
	}

	// wyszukaj pierwszą turę, która swoim range obejmuje terazniejszy czas
	for _, tura := range cfg.Tury {
		if now.After(tura.Range.Begin) && now.Before(tura.Range.End) {
			return &tura
		}
	}

	// jeśli żadna tura nie obejmuje teraz, zwróć pierwszą turę w kolejności
	sort.Slice(cfg.Tury, func(i, j int) bool {
		return cfg.Tury[i].Range.Begin.Before(cfg.Tury[j].Range.Begin)
	})
	return &cfg.Tury[0]
}
