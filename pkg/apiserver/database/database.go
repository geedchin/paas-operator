package database

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/farmer-hutao/k6s/pkg/agent"
	"github.com/farmer-hutao/k6s/pkg/apiserver/utils/sshcli"
	"log"
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
	NotInstalled DatabaseStatus = "not-installed"
	Running      DatabaseStatus = "running"
	Stoped       DatabaseStatus = "stoped"
	Failed       DatabaseStatus = "failed"
	Unknown      DatabaseStatus = "unknown"
)

const (
	Start     DatabaseAction = "start"
	Stop      DatabaseAction = "stop"
	Install   DatabaseAction = "install"
	Restart   DatabaseAction = "restart"
	Uninstall DatabaseAction = "uninstall"
)

var DatabaseStatusMap = map[DatabaseStatus]struct{}{
	NotInstalled: {},
	Running:      {},
	Stoped:       {},
	// Failed:       {}, // user won't set a app to failed
	// Unknown:      {}, // user won't set a app to unknown
}

var DatabaseActionMap = map[DatabaseAction]struct{}{
	Start:   {},
	Stop:    {},
	Install: {},
	Restart: {},
}

type Interface interface {
	// keep the realtime status == expect status
	UpdateStatus()
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

func (d *Database) UpdateStatus() {
	expectStatus := d.App.Status.Expect
	realStatus := d.App.Status.Realtime

	// need init agent
	if expectStatus == Running && realStatus == NotInstalled {
		err := initAgent(d.Host[0].IP, d.Host[0].Username, d.Host[0].Password)
		if err != nil {
			log.Println("Error: " + err.Error())
			return
		}
		time.Sleep(5 * time.Second)
		callAgent(Install, d)
	}
	if expectStatus == Running && realStatus != NotInstalled {
		callAgent(Start, d)
	}
	if expectStatus == Stoped {
		callAgent(Stop, d)
	}
}

func (d *Database) Status() *Statusx {
	return &d.App.Status
}

func initAgent(ip, username, password string) error {
	agentTarPath := "/opt/app/agent.tar.gz"
	sshCli := sshcli.New(ip, username, password, "22")
	if err := sshCli.ValidateConn(); err != nil {
		return err
	}
	defer sshCli.Cli.Close()

	// upload
	if err := sshCli.UploadFile(agentTarPath, agentTarPath); err != nil {
		return err
	}

	cmd := fmt.Sprintf("tar -xzvf %s -C /opt/app/ && sh /opt/app/agent/agent.sh", agentTarPath)
	result, err := sshCli.ExecCmd(cmd)

	log.Println("Exec result: " + result)
	return err
}

func callAgent(action DatabaseAction, db *Database) {
	var agentUrlPrefix = "http://" + db.Host[0].IP + ":3335/"
	var agentUrl = agentUrlPrefix + string(action)

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
		log.Println(err)
		return
	}

	resp, err := http.Post(agentUrl, "application/json;charset=utf-8", bytes.NewBuffer([]byte(jsonBody)))
	if err != nil {
		log.Println(err)
		return
	}

	if resp.StatusCode != 200 {
		log.Println("Call agent got not 200: " + resp.Status)
	}

	log.Println("Call agent: " + resp.Status)

	// success, then update realtime status
	switch action {
	case Start:
		db.App.Status.Realtime = Running
	case Stop:
		db.App.Status.Realtime = Stoped
	case Install:
		db.App.Status.Realtime = Running
	case Restart:
	// TODO(ht): how to restart?
	case Uninstall:
		// TODO
	}
}
