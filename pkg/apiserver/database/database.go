package database

import (
	"github.com/kataras/iris"
)

type DatabaseStatus string // eg. running

type DatabaseAction string // eg. install

const (
	// user can ask db's status to not-installed/running/stoped
	NotInstalled DatabaseStatus = "not-installed"
	Running      DatabaseStatus = "running"
	Stopped      DatabaseStatus = "stoped"

	// program set
	Restart DatabaseStatus = "restart"

	Failed  DatabaseStatus = "failed"
	Unknown DatabaseStatus = "unknown"

	// mid status is not in the DatabaseStatusMap
	Starting   DatabaseStatus = "starting"
	Installing DatabaseStatus = "installing"
	Stopping   DatabaseStatus = "stopping"
	Restarting DatabaseStatus = "restarting"
)

const (
	AStart     DatabaseAction = "start"
	AStop      DatabaseAction = "stop"
	AInstall   DatabaseAction = "install"
	ARestart   DatabaseAction = "restart"
	AUninstall DatabaseAction = "uninstall"
)

var DatabaseStatusMap = map[DatabaseStatus]struct{}{
	NotInstalled: {},
	Running:      {},
	Stopped:      {},
	Restart:      {}, // it's a action, but we need it
	// Failed:       {}, // user won't set a app to failed
	// Unknown:      {}, // user won't set a app to unknown
}

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
	Add(name string, db Database) error
	// Get a db from databases; If the db is not exist, return {}, false
	Get(name string) (Database, bool)
}

// Database specify a database resource
type Database interface {
	UpdateStatus(action DatabaseAction, ctx iris.Context)
	GetStatus() *Statusx
	GetApp() *Appx
	GetName() string
	GetHosts() []Hostx
}

type Hostx struct {
	IP       string `json:"ip"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Appx struct {
	RepoURL   string            `json:"repo_url"`  // http://192.168.19.101:8080/
	Install   string            `json:"install"`   // install.sh
	Start     string            `json:"start"`     // start.sh
	Stop      string            `json:"stop"`      // stop.sh
	Restart   string            `json:"restart"`   //restart.sh
	Uninstall string            `json:"uninstall"` // uninstall.sh
	Package   string            `json:"package"`   // mysql-5.7.tar.gz
	Metadata  map[string]string `json:"metadata"`
	Status    Statusx           `json:"status"`
}

type Statusx struct {
	Expect   DatabaseStatus `json:"expect"`   // running
	Realtime DatabaseStatus `json:"realtime"` // failed
}
