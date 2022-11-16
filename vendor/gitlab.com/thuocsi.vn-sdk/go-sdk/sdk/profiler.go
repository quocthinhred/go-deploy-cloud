package sdk

import (
	"gitlab.com/thuocsi.vn-sdk/go-sdk/sdk/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"sync"
	"time"

)

type timeInfo struct {
	Action    string    `bson:"action"`
	StartTime time.Time `bson:"start_time,omitempty"`
	EndTime   time.Time `bson:"end_time,omitempty"`
	LatencyMS int       `bson:"latency_ms"`
	Message   string    `bson:"msg,omitempty"`
}

// Profiler represent time-profiler
type Profiler struct {
	name         string               `bson:"name"`
	keys         []string             `bson:"keys,omitempty"`
	startTime    time.Time            `bson:"start_time,omitempty"`
	latencies    map[string]*timeInfo `bson:"latencies"`
	slowTheshold int                  `bson:"slow_threshold"`
	lock         *sync.Mutex
}

var profilerDBName string
var profilerDBMap = map[string]*db.Instance{}
var profilerDBSession *mongo.Database
var localLock *sync.Mutex = &sync.Mutex{}

// NewProfiler setup new profiler
func NewProfiler(name string, keys []string, slowThreshold int) *Profiler {
	p := &Profiler{
		name:         name,
		slowTheshold: slowThreshold,
		keys:         keys,
		startTime:    time.Now(),
		latencies:    map[string]*timeInfo{},
		lock:         &sync.Mutex{},
	}
	if profilerDBMap[name] == nil && profilerDBName != "" && profilerDBSession != nil {
		db := &db.Instance{
			ColName: name,
			TemplateObject: &Profiler{},
		}
		db.ApplyDatabase(profilerDBSession)
		t := true
		db.CreateIndex(bson.D{
			{"keys", 1},
		}, &options.IndexOptions{
			Background: &t,
		})

		localLock.Lock()
		profilerDBMap[name] = db
		localLock.Unlock()
	}

	return p
}

// Start start an action
func (p *Profiler) Start(action string) {
	info := &timeInfo{
		Action:    action,
		StartTime: time.Now(),
	}
	p.lock.Lock()
	p.latencies[action] = info
	p.lock.Unlock()
}

// End end an action
func (p *Profiler) End(action string, msg string) {
	actionInfo := p.latencies[action]
	if actionInfo != nil {
		actionInfo.EndTime = time.Now()
		actionInfo.LatencyMS = int(p.latencies[action].EndTime.Sub(p.latencies[action].StartTime) / time.Millisecond)
		actionInfo.Message = msg
	}
}

// Done write log if over threshold
func (p *Profiler) Done() {
	if int(time.Now().Sub(p.startTime)/time.Millisecond) < p.slowTheshold {
		return
	}

	if profilerDBMap[p.name] != nil {
		profilerDBMap[p.name].Create(p)
	}
}

// InitProfiler ...
func InitProfiler(s *mongo.Database) {
	profilerDBSession = s
	profilerDBName = s.Name()
}
