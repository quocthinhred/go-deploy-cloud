package schedule

import (
	"errors"
	"gitlab.com/thuocsi.vn-sdk/go-sdk/sdk/common"
	"gitlab.com/thuocsi.vn-sdk/go-sdk/sdk/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

// Process (error, string, *time.Time)
// error = error returned of function
// string = note
// time.Time = next run time
type Process = func(*time.Time, *Config) (error, string, *time.Time)

// History ....
type History struct {
	ID              *primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	LastUpdatedTime *time.Time          `json:"lastUpdatedTime,omitempty" bson:"last_updated_time,omitempty"`
	CreatedTime     *time.Time          `json:"createdTime,omitempty" bson:"created_time,omitempty"`

	Topic  string `json:"topic,omitempty" bson:"topic,omitempty"`
	Result string `json:"result,omitempty" bson:"result,omitempty"`
}

type Config struct {
	ID              *primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	LastUpdatedTime *time.Time          `json:"lastUpdatedTime,omitempty" bson:"last_updated_time,omitempty"`
	CreatedTime     *time.Time          `json:"createdTime,omitempty" bson:"created_time,omitempty"`

	Topic   string     `bson:"topic,omitempty"`
	NextRun *time.Time `bson:"next_run,omitempty"`
}

// ConfigDB config for schedule
type ConfigDB struct {
	db         *db.Instance
	history    *db.Instance
	started    bool
	processors map[string]Process
}

// NewConfigDB create new config db
func NewConfigDB(colName string, processors map[string]Process) *ConfigDB {
	config := &ConfigDB{
		db: &db.Instance{
			ColName:        colName,
			TemplateObject: &Config{},
		},
		history: &db.Instance{
			ColName:        colName + "_history",
			TemplateObject: &History{},
		},
		processors: processors,
		started:    false,
	}

	return config
}

func (c *ConfigDB) GetConfigDB() *db.Instance {
	return c.db
}

func (c *ConfigDB) GetHistoryDB() *db.Instance {
	return c.history
}

// InitAndStart apply connected database and start the scheduled worker
func (c *ConfigDB) InitAndStart(database *mongo.Database) {
	c.Init(database)
	c.Start()
}

func (c *ConfigDB) Init(database *mongo.Database) {
	c.db.ApplyDatabase(database)
	c.history.ApplyDatabase(database)

	// create indexes
	t := true
	c.db.CreateIndex(bson.D{
		{"next_run", 1},
	}, &options.IndexOptions{
		Background: &t,
	})

	c.history.CreateIndex(bson.D{
		{"topic", 1},
	}, &options.IndexOptions{
		Background: &t,
	})
}

func (c *ConfigDB) Start() {
	// if started => return
	if c.started {
		return
	}

	c.started = true
	go (func() {

		for true {

			// try to get config
			now := time.Now()
			result := c.db.UpdateOne(bson.M{
				"next_run": bson.M{
					"$lt": now,
				},
			}, bson.M{
				"next_run": now.Add(1 * time.Hour),
			})

			// if exists config
			if result.Status == common.APIStatus.Ok {
				config := result.Data.([]*Config)[0]

				// check process exists for this topic
				if c.processors[config.Topic] != nil {

					// try - catch
					err, note, nextRun := (func() (err error, note string, next *time.Time) {
						defer func() {
							if rec := recover(); rec != nil {
								log.Println("panic: ", rec)
								err = errors.New("Panic when run " + config.Topic)
							}
						}()

						err, note, next = c.processors[config.Topic](&now, config)
						return err, note, next
					})()

					// if success
					if err == nil {
						c.db.UpdateOne(bson.M{
							"_id": config.ID,
						}, bson.M{
							"next_run": nextRun,
						})

						c.history.Create(History{
							Topic:  config.Topic,
							Result: "OK " + note,
						})
					} else { // if error
						c.db.UpdateOne(bson.M{
							"_id": config.ID,
						}, bson.M{
							"next_run": now.Add(5 * time.Minute),
						})

						c.history.Create(History{
							Topic:  config.Topic,
							Result: "ERROR " + err.Error(),
						})
					}
				}
			}

			// do for every 5 minutes
			time.Sleep(5 * time.Minute)

		}
	})()
}
