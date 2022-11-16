package db

import (
	"context"
	"errors"
	"gitlab.com/thuocsi.vn-sdk/go-sdk/sdk/common"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"os"
	"time"
)

type DBConnectionLog struct {
	Host          string    `bson:"host,omitempty"`
	ConnectedTime time.Time `bson:"connected_time,omitempty"`
}

type Configuration struct {
	AuthMechanism      string
	Username           string
	Password           string
	AuthDB             string
	DBName             string
	SSL                bool
	Address            string
	SecondaryPreferred bool
	DoWriteTest        bool
}

type Client struct {
	Name        string
	Config      *Configuration
	OnConnected OnConnectedHandler
}

type OnConnectedHandler = func(session *mongo.Database) error

func (c *Client) Connect() error {
	// setup default config
	hostname, _ := os.Hostname()
	hostname = hostname + " " + os.Getenv("env")
	heartBeat := 60 * time.Second
	maxIdle := 180 * time.Second
	min := uint64(2)

	// setup options
	opt := &options.ClientOptions{
		AppName: &hostname,
		Auth: &options.Credential{
			AuthMechanism: c.Config.AuthMechanism,
			AuthSource:    c.Config.AuthDB,
			Username:      c.Config.Username,
			Password:      c.Config.Password,
		},
		HeartbeatInterval: &heartBeat,
		MaxConnIdleTime: &maxIdle,
		MinPoolSize: &min,
	}
	opt.ApplyURI(c.Config.Address)
	if c.Config.SecondaryPreferred {
		opt.ReadPreference = readpref.SecondaryPreferred()
	}

	// new client
	client, err := mongo.NewClient(opt)
	if err != nil {
		return err
	}

	// connect
	err = client.Connect(context.TODO())
	if err != nil {
		return err
	}

	database := client.Database(c.Config.DBName)

	// try to test write & log connection
	if c.Config.DoWriteTest {
		inst := Instance{
			ColName:        "_db_connection",
			TemplateObject: &DBConnectionLog{},
		}
		inst.ApplyDatabase(database)
		name, err := os.Hostname()
		if err != nil {
			name = "Unknown"
		}
		testResult := inst.Create(DBConnectionLog{
			Host:          name,
			ConnectedTime: time.Now(),
		})

		if testResult.Status != common.APIStatus.Ok {
			return errors.New(testResult.Status + " / " + testResult.ErrorCode + " => " + testResult.Message)
		}
	}

	// on connected
	c.OnConnected(database)
	return nil
}
