package model

import (
	"example.com/micro/model/core/obj"
	"gitlab.com/thuocsi.vn-sdk/go-sdk/sdk/db"
	"gitlab.com/thuocsi.vn-sdk/go-sdk/sdk/schedule"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const (
	AutoLogin         = "AUTO_LOGIN"
	TimeToDeleteDraft = 24 * 60 * 60
)

var DBSchedule = &db.Instance{
	ColName:        "schedule_auto_login",
	TemplateObject: &schedule.Config{},
}

func InitSchedule(database *mongo.Database) {
	DBSchedule.ApplyDatabase(database)
	DBSchedule.CreateIndex(bson.D{
		{"topic", 1},
	}, &options.IndexOptions{
		Background: obj.WithBool(true),
		Unique:     obj.WithBool(true),
	})
	DBSchedule.Create(&schedule.Config{
		Topic:   AutoLogin,
		NextRun: obj.WithTime(time.Now().Add(24 * time.Hour)),
	})
}
