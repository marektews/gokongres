package db

import "go.mongodb.org/mongo-driver/bson/primitive"

type Users struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	Login        int                `bson:"login"`
	Hash         string             `bson:"hash"`
	Fn           string             `bson:"fn"`
	Ln           string             `bson:"ln"`
	IsSra        int                `bson:"is_sra"`
	IsSrp        int                `bson:"is_srp"`
	IsPk         int                `bson:"is_pk"`
	IsRja        int                `bson:"is_rja"`
	IsMonitoring int                `bson:"is_monitoring"`
	IsUsers      int                `bson:"is_users"`
	IsLimits     int                `bson:"is_limits"`
}
