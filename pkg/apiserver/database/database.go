package database

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/farmer-hutao/k6s/pkg/agent"
	"github.com/farmer-hutao/k6s/pkg/apiserver/utils/sshcli"
	"github.com/kataras/iris"
	"net/http"
	"sync"
	"time"
)

//TODO(ht): use etcd

// databases holds all created database
type Databases struct {
	lock  sync.RWMutex
	dbMap map[string]Database
}

// add a db to databases;if the db is already exist, return error
func (dbs *Databases) Add(name string, db Database) error {
	dbs.lock.Lock()
	defer dbs.lock.Unlock()
	if _, ok := dbs.dbMap[name]; ok {
		return errors.New("database is already exist: " + name)
	}
	dbs.dbMap[name] = db
	return nil
}

// get a db from dabatases; if exist,bool->true;else bool->false
func (dbs *Databases) Get(name string) (Database, bool) {
	dbs.lock.RLock()
	defer dbs.lock.RUnlock()
	if db, ok := dbs.dbMap[name]; ok {
		return db, true
	}
	return Database{}, false
}

// global databases store
var DatabaseList = &Databases{
	dbMap: make(map[string]Database, 0),
}

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

type Interface interface {
	// keep the realtime status == expect status
	UpdateStatus(action DatabaseAction, ctx iris.Context)
	Status() *Statusx
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

func (d *Database) UpdateStatus(action DatabaseAction, ctx iris.Context) {
	ctx.Application().Logger().Infof("The database with name <%s> start update status; "+
		"expect status: <%s>; realtime status: <%s>;", d.Name, d.App.Status.Expect, d.App.Status.Realtime)

	switch action {
	case AInstall:
		if err := InitAgent(d.Host[0].IP, d.Host[0].Username, d.Host[0].Password, ctx); err != nil {
			ctx.Application().Logger().Errorf("Init agent failed: ", err)
			d.App.Status.Realtime = Failed
			return
		}
		// wait the agent starting
		time.Sleep(5 * time.Second)
		if err := CallToAgent(AInstall, d, ctx); err != nil {
			d.App.Status.Realtime = Failed
			return
		}
		d.App.Status.Realtime = Running
	case AStart:
		if err := CallToAgent(AStart, d, ctx); err != nil {
			d.App.Status.Realtime = Failed
			return
		}
		d.App.Status.Realtime = Running
	case AStop:
		if err := CallToAgent(AStop, d, ctx); err != nil {
			d.App.Status.Realtime = Failed
			return
		}
		d.App.Status.Realtime = Stopped
	case ARestart:
		if err := CallToAgent(ARestart, d, ctx); err != nil {
			d.App.Status.Realtime = Failed
			return
		}
		d.App.Status.Realtime = Running
	case AUninstall:
		if err := CallToAgent(AUninstall, d, ctx); err != nil {
			d.App.Status.Realtime = Failed
			return
		}
		d.App.Status.Realtime = NotInstalled
	}
}

func (d *Database) Status() *Statusx {
	return &d.App.Status
}

func InitAgent(ip, username, password string, ctx iris.Context) error {
	ctx.Application().Logger().Info("start to init agent!!!")

	agentTarPath := "/opt/app/agent.tar.gz"
	sshCli := sshcli.New(ip, username, password, "22")
	if err := sshCli.ValidateConn(); err != nil {
		ctx.Application().Logger().Error(err)
		return err
	}
	defer sshCli.Cli.Close()

	// upload
	if err := sshCli.UploadFile(agentTarPath, agentTarPath); err != nil {
		ctx.Application().Logger().Error(err)
		return err
	}

	// start agent
	cmd := fmt.Sprintf("tar -xzvf %s -C /opt/app/ && sh /opt/app/agent/agent.sh", agentTarPath)
	result, err := sshCli.ExecCmd(cmd)
	ctx.Application().Logger().Infof("Exec cmd: <%s> get result: <%s>", cmd, result)
	if err != nil {
		ctx.Application().Logger().Errorf("Exec cmd: <%s> get error: <%s>", cmd, err.Error())
		return err
	}
	return nil
}

func CallToAgent(action DatabaseAction, db *Database, ctx iris.Context) error {
	var agentUrlPrefix = "http://" + db.Host[0].IP + ":3335/"
	var agentUrl = agentUrlPrefix + string(action)

	ctx.Application().Logger().Infof("call to agent: %s", agentUrl)

	var appInfo agent.AppInfo

	appInfo.RepoURL = db.App.RepoURL
	appInfo.Install = db.App.Install
	appInfo.Start = db.App.Start
	appInfo.Stop = db.App.Stop
	appInfo.Restart = db.App.Restart
	appInfo.Uninstall = db.App.Uninstall
	appInfo.Package = db.App.Package
	appInfo.Metadata = db.App.Metadata

	jsonBody, err := json.Marshal(appInfo)
	if err != nil {
		ctx.Application().Logger().Error(err)
		return err
	}

	resp, err := http.Post(agentUrl, "application/json;charset=utf-8", bytes.NewBuffer([]byte(jsonBody)))
	if err != nil {
		ctx.Application().Logger().Error(err)
		return err
	}

	if resp.StatusCode != 200 {
		ctx.Application().Logger().Errorf("Result code != 200: %s", resp.Status)
		return errors.New(resp.Status)
	}
	return nil
}
