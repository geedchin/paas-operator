package database

import (
	"github.com/kataras/iris"
)

type DatabaseStatus string // eg. [ running, stopped, ... ]

type DatabaseAction string // eg. [ install, stop, ... ]

const (
	// user can asks db's status to be not-installed/running/stopped and restart
	// restart means running -> stopped -> running
	NotInstalled DatabaseStatus = "not-installed"
	Running      DatabaseStatus = "running"
	Stopped      DatabaseStatus = "stopped"
	// essentially not a status
	Restart DatabaseStatus = "restart"

	// failed and unknown is a real status will be happen, but user can't set itz
	Failed  DatabaseStatus = "failed"
	Unknown DatabaseStatus = "unknown"

	// middle status isn't need in the DatabaseStatusMap
	Starting   DatabaseStatus = "starting"
	Installing DatabaseStatus = "installing"
	Stopping   DatabaseStatus = "stopping"
	Restarting DatabaseStatus = "restarting"
)

// all action
const (
	AStart     DatabaseAction = "start"
	AStop      DatabaseAction = "stop"
	AInstall   DatabaseAction = "install"
	ARestart   DatabaseAction = "restart"
	AUninstall DatabaseAction = "uninstall"
)

// all status user can set
var DatabaseStatusMap = map[DatabaseStatus]struct{}{
	NotInstalled: {},
	Running:      {},
	Stopped:      {},
	Restart:      {}, // it's a action, but we need use it as a status
	// Failed:       {}, // user won't set a app to failed
	// Unknown:      {}, // user won't set a app to unknown
}

// all action
var DatabaseActionMap = map[DatabaseAction]struct{}{
	AStart:     {},
	AStop:      {},
	AInstall:   {},
	ARestart:   {},
	AUninstall: {},
}

// Databases is used to store all database
type Databases interface {
	// Add a db to databases; If the db is already exist, return error
	Add(name string, db Database, ctx iris.Context) error
	// Get a db from databases; If the db is not exist, return {}, false
	Get(name string, ctx iris.Context) (Database, bool)
	// Delete a db from databases; If the db is exist, return the db and nil, else return nil and nil
	// if some error occur, return nil and the error
	// 1. db exist, delete success -> return (db, nil)
	// 2. db not exist -> return (nil, nil)
	// 3. some error occur -> return (nil, err)
	Delete(name string, ctx iris.Context) (Database, error)
}

// Database specify a database resource
type Database interface {
	UpdateStatus(action DatabaseAction, ctx iris.Context)
	GetStatus() *Statusx
	GetName() string
	GetApp() *Appx
	GetHosts() []Hostx
}

type EventLog interface {
	// GetEvents returns all events with a resource
	GetEvents() []map[string]string
	// AddEvent add a event with a resource to events
	AddEvent(map[string]string, iris.Context) (bool, error)
}
