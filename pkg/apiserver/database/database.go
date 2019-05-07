package database

import (
	"errors"
	"github.com/kataras/iris"
	"sync"
)

type Databases struct {
	lock  sync.RWMutex
	dbMap map[string]Database
}

func (dbs *Databases) Add(name string, db Database) error {
	dbs.lock.Lock()
	defer dbs.lock.Unlock()
	if _, ok := dbs.dbMap[name]; ok {
		return errors.New("database is already exist: "+ name)
	}
	dbs.dbMap[name] = db
	return nil
}

func (dbs *Databases) Get(name string) (Database, bool) {
	dbs.lock.RLock()
	defer dbs.lock.RUnlock()
	if db, ok := dbs.dbMap[name]; ok {
		return db, true
	}
	return Database{}, false
}

var DatabaseList = &Databases{
	dbMap: make(map[string]Database, 0),
}

type DatabaseStatus string
type DatabaseAction string

const (
	NotInstalled DatabaseStatus = "not-installed"
	Running      DatabaseStatus = "running"
	Stoped       DatabaseStatus = "stoped"
	Failed       DatabaseStatus = "failed"
	Unknown      DatabaseStatus = "unknown"
)

const (
	Start   DatabaseAction = "start"
	Stop    DatabaseAction = "stop"
	Install DatabaseAction = "install"
	Restart DatabaseAction = "restart"
)

var DatabaseStatusMap = map[DatabaseStatus]struct{}{
	NotInstalled: {},
	Running:      {},
	Stoped:       {},
	Failed:       {},
	Unknown:      {},
}

var DatabaseActionMap = map[DatabaseAction]struct{}{
	Start:   {},
	Stop:    {},
	Install: {},
	Restart: {},
}

type Interface interface {
	UpdateStatus(expect DatabaseStatus)
}

type Database struct {
	// mysql-5.7-xxx-192.168.19.100
	Name string `json:"name"`
	Host []struct {
		IP       string `json:"ip"`
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"host"`
	App struct {
		RepoURL   string            `json:"repo_url"`  // http://192.168.19.101:8080/
		Install   string            `json:"install"`   // install.sh
		Start     string            `json:"start"`     // start.sh
		Stop      string            `json:"stop"`      // stop.sh
		Restart   string            `json:"restart"`   //restart.sh
		Uninstall string            `json:"uninstall"` // uninstall.sh
		Package   string            `json:"package"`   // mysql-5.7.tar.gz
		Metadata  map[string]string `json:"metadata"`
		Status    Statusx           `json:"status"`
	} `json:"app"`
}

type Statusx struct {
	Expect   DatabaseStatus `json:"expect"`   // running
	Realtime DatabaseStatus `json:"realtime"` // failed
}

// action can be [start / stop / install / restart / uninstall]
func (d *Database) UpdateStatus(expect DatabaseStatus) error {
	if _, ok := DatabaseStatusMap[expect]; !ok {
		return errors.New("expect status is illegal: " + string(expect))
	}

	d.App.Status.Expect = expect

	// TODO call agent
	return nil
}

func (d *Database) Status(ctx iris.Context) *Statusx {
	return &d.App.Status
}
