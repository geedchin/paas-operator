package database

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/farmer-hutao/k6s/pkg/apiserver/utils"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/kataras/iris"

	"github.com/farmer-hutao/k6s/pkg/agent"
	"github.com/farmer-hutao/k6s/pkg/apiserver/utils/sshcli"
)

var (
	WORK_DIR       = os.Getenv("APISERVER_WORK_DIR")
	AGENT_ZIP_NAME = os.Getenv("AGENT_ZIP_NAME")
	AGENT_PORT     = os.Getenv("AGENT_PORT")
)

func init() {
	if WORK_DIR == "" {
		log.Printf("Warning: %s is unset, use default value: %s", "APISERVER_WORK_DIR", WORK_DIR)
		WORK_DIR = "/opt/app/"
	}
	if AGENT_ZIP_NAME == "" {
		log.Printf("Warning: %s is unset, use default value: %s", "AGENT_ZIP_NAME", AGENT_ZIP_NAME)
		AGENT_ZIP_NAME = "agent.tar.gz"
	}
	if AGENT_PORT == "" {
		log.Printf("Warning: %s is unset, use default value: %s", "AGENT_PORT", AGENT_PORT)
		AGENT_PORT = "3335"
	}
}

// GenericDatabase is a generic implement of Database interface
type GenericDatabase struct {
	// mysql-5.7-xxx-192.168.19.100
	Name  string  `json:"name"`
	Host  []Hostx `json:"host"`
	App   Appx    `json:"app"`
	Event Eventx  `json:"event"`
}

type Hostx struct {
	IP   string  `json:"ip"`
	Auth []Authx `json:"auth"`
}

type Authx struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Appx struct {
	RepoURL   string            `json:"repo_url"`  // http://192.168.19.101:8080/
	Install   string            `json:"install"`   // install.sh
	Start     string            `json:"start"`     // start.sh
	Stop      string            `json:"stop"`      // stop.sh
	Restart   string            `json:"restart"`   // restart.sh
	Uninstall string            `json:"uninstall"` // uninstall.sh
	Package   string            `json:"package"`   // mysql-5.7.tar.gz
	Metadata  map[string]string `json:"metadata"`
	Status    Statusx           `json:"status"`
}

type Statusx struct {
	Expect   DatabaseStatus `json:"expect"`   // running
	Realtime DatabaseStatus `json:"realtime"` // failed
}

// event saves all key log print with script in vm.
// like:
// [
//   {"some error":"some error occur, out of memory"},
//   {"xxx":"xxxxx"},
//   ……
// ]
type Eventx []map[string]string

func (d *GenericDatabase) UpdateStatus(action DatabaseAction, ctx iris.Context) {
	ctx.Application().Logger().Infof("The database with name <%s> start update status; "+
		"expect status: <%s>; realtime status: <%s>;", d.Name, d.App.Status.Expect, d.App.Status.Realtime)

	// TODO(ht) consider concurrency
	updateFn := func() {
		err := GetETCDDatabases().Add(d.GetName(), d, ctx)
		if err != nil {
			ctx.Application().Logger().Errorf("Got some error: %s", err.Error())
		}
	}
	updateFn()
	defer updateFn()

	switch action {
	case AInstall:
		if err := InitAgent(d.Host[0].IP, d.Host[0].Auth, ctx); err != nil {
			ctx.Application().Logger().Errorf("Init agent failed: <%s>", err.Error())
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

func (d *GenericDatabase) GetStatus() *Statusx {
	return &d.App.Status
}

func (d *GenericDatabase) GetName() string {
	return d.Name
}

func (d *GenericDatabase) GetApp() *Appx {
	return &d.App
}

func (d *GenericDatabase) GetHosts() []Hostx {
	return d.Host
}

func (d *GenericDatabase) AddEvent(event map[string]string, ctx iris.Context) (bool, error) {
	return false, nil
}

func (d *GenericDatabase) GetEvents() []map[string]string {
	return nil
}

func InitAgent(ip string, auth []Authx, ctx iris.Context) error {
	var tmpDir = "/tmp/"
	var sshPort = "22"

	var (
		agentUser   string
		agentPasswd string
		sshUser     string
		sshPasswd   string
	)

	if len(auth) < 1 {
		return errors.New("Auth is nil")
	} else if len(auth) == 1 {
		agentUser = auth[0].Username
		agentPasswd = auth[0].Password
		sshUser = agentUser
		sshPasswd = agentPasswd
	} else {
		agentUser = auth[0].Username
		agentPasswd = auth[0].Password
		sshUser = auth[1].Username
		sshPasswd = auth[1].Password
	}

	ctx.Application().Logger().Info("start to init agent!!!")

	agentTarPath := filepath.Join(WORK_DIR, AGENT_ZIP_NAME)
	tmpTarPath := filepath.Join(tmpDir, AGENT_ZIP_NAME)

	sshCli := sshcli.New(ip, sshUser, sshPasswd, sshPort)
	// max -> 10 minutes = 30*20s
	retryTimes := 30
	retryGap := 20 * time.Second
	okFlag := false
	for i := 0; i < retryTimes; i++ {
		if err := sshCli.ValidateConn(); err != nil {
			ctx.Application().Logger().Errorf("ValidateConn failed: <%s>; retry <%d/%d>", err.Error(), i, retryTimes)
			time.Sleep(retryGap)
			continue
		}
		okFlag = true
		break
	}
	if !okFlag {
		return errors.New(fmt.Sprintf("ValidateConn failed, retry <%d/%d>", retryTimes, retryTimes))
	}

	defer sshCli.Cli.Close()

	// upload
	if err := sshCli.UploadFile(agentTarPath, tmpTarPath); err != nil {
		ctx.Application().Logger().Error(err)
		return err
	}

	// start agent
	doWithSuCmd := fmt.Sprintf("cat %s > dowithsu.sh && ./dowithsu.sh %s %s", utils.DoWithSu, agentUser, agentPasswd)
	cmd := fmt.Sprintf("mkdir %s && tar -xzvf %s -C %s && %s 'sh %sagent/agent.sh'", WORK_DIR, tmpTarPath, WORK_DIR, doWithSuCmd, WORK_DIR)
	ctx.Application().Logger().Infof("Prepare to exec cmd: %s", cmd)
	result, err := sshCli.ExecCmd(cmd)
	ctx.Application().Logger().Infof("Exec cmd: <%s> get result: <%s>", cmd, result)
	if err != nil {
		ctx.Application().Logger().Errorf("Exec cmd: <%s> get error: <%s>", cmd, err.Error())
		return err
	}
	return nil
}

func CallToAgent(action DatabaseAction, db *GenericDatabase, ctx iris.Context) error {
	var agentUrlPrefix = fmt.Sprintf("http://%s:%s/", db.GetHosts()[0].IP, AGENT_PORT)
	var agentUrl = agentUrlPrefix + string(action)

	ctx.Application().Logger().Infof("call to agent: %s", agentUrl)

	var appInfo agent.AppInfo

	appInfo.RepoURL = db.GetApp().RepoURL
	appInfo.Install = db.GetApp().Install
	appInfo.Start = db.GetApp().Start
	appInfo.Stop = db.GetApp().Stop
	appInfo.Restart = db.GetApp().Restart
	appInfo.Uninstall = db.GetApp().Uninstall
	appInfo.Package = db.GetApp().Package
	appInfo.Metadata = db.GetApp().Metadata

	// repo_url and package environment is needed by scripts.
	if _, ok := db.GetApp().Metadata["REPO_URL"]; !ok {
		appInfo.Metadata["REPO_URL"] = appInfo.RepoURL
	}
	if _, ok := db.GetApp().Metadata["PACKAGE"]; !ok {
		appInfo.Metadata["PACKAGE"] = appInfo.Package
	}

	jsonBody, err := json.Marshal(appInfo)
	if err != nil {
		ctx.Application().Logger().Error(err)
		return err
	}

	ctx.Application().Logger().Infof("Call to agent with body:\n%s", string(jsonBody))

	resp, err := http.Post(agentUrl, "application/json;charset=utf-8", bytes.NewBuffer(jsonBody))
	if err != nil {
		ctx.Application().Logger().Error(err)
		return err
	}

	if resp.StatusCode != 200 {
		// TODO(ht) add buf
		var bodyBytes = make([]byte, 512)
		n, err := resp.Body.Read(bodyBytes)
		if err != nil && err != io.EOF {
			ctx.Application().Logger().Errorf("Read resp failed: <%s>", err.Error())
			return err
		}
		errMsg := fmt.Sprintf("status: <%s>; msg: <%s>", resp.Status, string(bodyBytes[0:n]))
		ctx.Application().Logger().Errorf("Result code != 200: %s", errMsg)
		return errors.New(errMsg)
	}
	return nil
}
