package action

import (
	"example.com/micro/model"
	"fmt"
	"gitlab.com/thuocsi.vn-sdk/go-sdk/sdk/schedule"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

func AutoLogin(timeNew *time.Time, scheduleConfig *schedule.Config) (error, string, *time.Time) {
	now := time.Now().UTC()
	nextTime := now.Add(24 * time.Hour)
	nextTime = time.Date(nextTime.Year(), nextTime.Month(), nextTime.Day(), 17, 0, 0, 0, time.UTC)
	fmt.Println("Schedule running!")
	return nil, "Successfully!", &nextTime
}

var processor = make(map[string]schedule.Process)
var ScheduleDb = schedule.NewConfigDB("schedule_auto_login", processor)

func InitScheduleAutoLogin(database *mongo.Database) {
	model.InitSchedule(database)
	processor[model.AutoLogin] = AutoLogin
	ScheduleDb.Init(database)
}
