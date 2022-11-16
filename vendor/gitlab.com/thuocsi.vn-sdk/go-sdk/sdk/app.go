package sdk

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"gitlab.com/thuocsi.vn-sdk/go-sdk/sdk/db"
	"gitlab.com/thuocsi.vn-sdk/go-sdk/sdk/websocket"
	"os"
	"runtime"
	"sync"
	"time"
)

//App ..
type App struct {
	Name             string
	ServerList       []APIServer
	DBList           []*db.Client
	WorkerList       []*AppWorker
	WSServerList     []*websocket.WSServer
	onAllDBConnected Task
	launched         bool
	hostname         string
}

// NewApp Wrap application
func NewApp(name string) *App {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "undefined"
	}
	app := &App{
		Name:         name,
		ServerList:   []APIServer{},
		DBList:       []*db.Client{},
		WorkerList:   []*AppWorker{},
		WSServerList: []*websocket.WSServer{},
		launched:     false,
		hostname:     hostname,
	}
	return app
}

func (app *App) GetConfigFromEnv() (map[string]string, error) {
	var configMap map[string]string
	configStr := os.Getenv("config")
	decoded, err := base64.StdEncoding.DecodeString(configStr)
	if err != nil {
		fmt.Println("[Parse config] Convert B64 config string error: " + err.Error())
		return nil, err
	}
	err = json.Unmarshal(decoded, &configMap)
	if err != nil {
		fmt.Println("[Parse config] Parse JSON with config string error: " + err.Error())
		return nil, err
	}
	return configMap, err
}

func (app *App) GetHostname() string {
	return app.hostname
}

// SetupDBClient ...
func (app *App) SetupDBClient(config db.Configuration, handler db.OnConnectedHandler) *db.Client {
	dbname, _ := json.Marshal(config.Address)
	client := &db.Client{
		Name: string(dbname) + config.AuthDB + "@" + config.Username + ":" + config.Password[len(config.Password)-5:],
		Config: &config,
		OnConnected: handler,
	}

	app.DBList = append(app.DBList, client)
	return client
}

// SetupDBClient ...
func (app *App) OnAllDBConnected(task Task) {
	app.onAllDBConnected = task
}

// SetupAPIServer ...
func (app *App) SetupAPIServer(t string) (APIServer, error) {
	var newID = len(app.ServerList) + 1
	var server APIServer
	switch t {
	case "HTTP":
		server = newHTTPAPIServer(newID, app.hostname)
	case "THRIFT":
		server = newThriftServer(newID, app.hostname)
	}

	if server == nil {
		return nil, errors.New("server type " + t + " is invalid (HTTP/THRIFT)")
	}
	app.ServerList = append(app.ServerList, server)
	return server, nil
}

// SetupWSServer
func (app *App) SetupWSServer(name string) *websocket.WSServer {
	wss := websocket.NewWebSocketServer(name)
	app.WSServerList = append(app.WSServerList, wss)
	return wss
}

// SetupWorker ...
func (app *App) SetupWorker() *AppWorker {
	var worker = &AppWorker{}
	app.WorkerList = append(app.WorkerList, worker)
	return worker
}

// callGCManually
func callGCManually() {
	for {
		time.Sleep(2 * time.Minute)
		runtime.GC()
	}
}

// Launch Launch app
func (app *App) Launch() error {

	if app.launched {
		return nil
	}

	app.launched = true

	name := app.Name + " / " + app.hostname
	fmt.Println("[ App " + name + " ] Launching ...")
	var wg = sync.WaitGroup{}

	// start connect to DB
	if len(app.DBList) > 0 {
		for _, db := range app.DBList {
			err := db.Connect()
			if err != nil {
				fmt.Println("Connect DB " + db.Name + " error: " + err.Error())
				return err
			}
		}
		fmt.Println("[ App " + name + " ] DBs connected.")
	}

	if app.onAllDBConnected != nil {
		app.onAllDBConnected()
		fmt.Println("[ App " + name + " ] On-all-DBs-connected handler executed.")
	}

	// start servers
	if len(app.ServerList) > 0 {
		for _, s := range app.ServerList {
			wg.Add(1)
			go s.Start(&wg)
		}
		fmt.Println("[ App " + name + " ] API Servers started.")
	}

	if len(app.WSServerList) > 0 {
		for _, s := range app.WSServerList {
			wg.Add(1)
			go s.Start()
		}
		fmt.Println("[ App " + name + " ] WebSocket Servers started.")
	}

	// start workers
	if len(app.WorkerList) > 0 {
		for _, wk := range app.WorkerList {
			wg.Add(1)
			go wk.Execute()
		}
		fmt.Println("[ App " + name + " ] Workers started.")
	}
	fmt.Println("[ App " + name + " ] Totally launched!")
	go callGCManually()
	wg.Wait()

	return nil
}