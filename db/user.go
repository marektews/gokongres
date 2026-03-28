package db

import "go.mongodb.org/mongo-driver/bson/primitive"

type UserPermissions struct {
	IsSra        bool `bson:"is_sra" json:"is_sra"`
	IsSrp        bool `bson:"is_srp" json:"is_srp"`
	IsPk         bool `bson:"is_pk" json:"is_pk"`
	IsRja        bool `bson:"is_rja" json:"is_rja"`
	IsMonitoring bool `bson:"is_monitoring" json:"is_monitoring"`
	IsUsers      bool `bson:"is_users" json:"is_users"`
	IsLimits     bool `bson:"is_limits" json:"is_limits"`
}

type User struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Uid         int                `bson:"uid"`
	Username    string             `bson:"username"`
	Hash        string             `bson:"hash"`
	FirstName   string             `bson:"first_name"`
	LastName    string             `bson:"last_name"`
	Permissions UserPermissions    `bson:"permissions"`
}
