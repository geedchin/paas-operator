package application

import (
	"github.com/kataras/iris"
)

type ApplicationStatus string // eg. [ running, stopped, ... ]

type ApplicationAction string // eg. [ install, stop, ... ]

type AppType string

const (
	APP_DATABASE   AppType = "database"
	APP_MIDDLEWARE AppType = "middleware"
)

const (
	// user can asks app's status to be not-installed/running/stopped and restart
	// restart means running -> stopped -> running
	NotInstalled ApplicationStatus = "not-installed"
	Running      ApplicationStatus = "running"
	Stopped      ApplicationStatus = "stopped"
	// essentially not a status
	Restart ApplicationStatus = "restart"

	// failed and unknown is a real status will be happen, but user can't set itz
	Failed  ApplicationStatus = "failed"
	Unknown ApplicationStatus = "unknown"

	// middle status isn't need in the ApplicationStatusMap
	Starting   ApplicationStatus = "starting"
	Installing ApplicationStatus = "installing"
	Stopping   ApplicationStatus = "stopping"
	Restarting ApplicationStatus = "restarting"
)

// all action
const (
	AStart     ApplicationAction = "start"
	AStop      ApplicationAction = "stop"
	AInstall   ApplicationAction = "install"
	ARestart   ApplicationAction = "restart"
	AUninstall ApplicationAction = "uninstall"
	ACheck     ApplicationAction = "check"
)

// all status user can set
var ApplicationStatusMap = map[ApplicationStatus]struct{}{
	NotInstalled: {},
	Running:      {},
	Stopped:      {},
	Restart:      {}, // it's a action, but we need use it as a status
	// Failed:       {}, // user won't set a app to failed
	// Unknown:      {}, // user won't set a app to unknown
}

// all action
var ApplicationActionMap = map[ApplicationAction]struct{}{
	AStart:     {},
	AStop:      {},
	AInstall:   {},
	ARestart:   {},
	AUninstall: {},
	ACheck:     {},
}

// Applications is used to store all Application
type Applications interface {
	// Add an app to Applications; If the app is already exist, return error
	Add(name string, app Application, ctx iris.Context) error
	// Update an app in Applications; If some error occur, return the error
	Update(name string, app Application, ctx iris.Context) error
	// Get an app from Applications; If the app is not exist, return {}, false
	Get(name string, ctx iris.Context) (Application, bool)
	// Delete an app from Applications; If the app is exist, return the app and nil, else return nil and nil
	// if some error occur, return nil and the error
	// 1. app exist, delete success -> return (app, nil)
	// 2. app not exist -> return (nil, nil)
	// 3. some error occur -> return (nil, err)
	Delete(name string, ctx iris.Context) (Application, error)
}

// Application specify a Application resource
type Application interface {
	UpdateStatus(action ApplicationAction, ctx iris.Context)
	SetStatus(expect, realtime ApplicationStatus, ctx iris.Context)
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
