package application

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kataras/iris"

	"github.com/farmer-hutao/paas-operator/pkg/agent"
	"github.com/farmer-hutao/paas-operator/pkg/apiserver/utils"
	"github.com/farmer-hutao/paas-operator/pkg/apiserver/utils/sshcli"
)

var (
	WORK_DIR       = os.Getenv("APISERVER_WORK_DIR")
	AGENT_ZIP_NAME = os.Getenv("AGENT_ZIP_NAME")
	AGENT_PORT     = os.Getenv("AGENT_PORT")
	OPERATOR_IP    = os.Getenv("OPERATOR_IP")
	OPERATOR_PORT  = os.Getenv("OPERATOR_PORT")
)

func init() {
	if WORK_DIR == "" {
		WORK_DIR = "/opt/app/"
		log.Printf("Warning: %s is unset, use default value: %s", "APISERVER_WORK_DIR", WORK_DIR)
	}
	if AGENT_ZIP_NAME == "" {
		AGENT_ZIP_NAME = "agent.tar.gz"
		log.Printf("Warning: %s is unset, use default value: %s", "AGENT_ZIP_NAME", AGENT_ZIP_NAME)
	}
	if AGENT_PORT == "" {
		AGENT_PORT = "3335"
		log.Printf("Warning: %s is unset, use default value: %s", "AGENT_PORT", AGENT_PORT)
	}
	if OPERATOR_IP == "" {
		OPERATOR_IP = "127.0.0.1"
		log.Printf("Warning: %s is unset, use default value: %s", "OPERATOR_IP", OPERATOR_IP)
	}
	if OPERATOR_PORT == "" {
		OPERATOR_PORT = "3334"
		log.Printf("Warning: %s is unset, use default value: %s", "OPERATOR_PORT", OPERATOR_PORT)
	}
}

// GenericApplication is a generic implement of Application interface
type GenericApplication struct {
	// mysql-5.7-xxx-192.168.19.100
	Name  string  `json:"name"`
	Type  string  `json:"type"`
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
	Check     string            `json:"check"`     // check.sh
	Package   string            `json:"package"`   // mysql-5.7.tar.gz
	Metadata  map[string]string `json:"metadata"`
	Status    Statusx           `json:"status"`
}

type Statusx struct {
	Expect   ApplicationStatus `json:"expect"`   // running
	Realtime ApplicationStatus `json:"realtime"` // failed
}

// event saves all key log print with script in vm.
// like:
// [
//   {"some error":"some error occur, out of memory"},
//   {"xxx":"xxxxx"},
//   ……
// ]
type Eventx []map[string]string

func (a *GenericApplication) UpdateStatus(action ApplicationAction, ctx iris.Context) {
	ctx.Application().Logger().Infof("The application with name <%s> start update status; "+
		"expect status: <%s>; realtime status: <%s>;", a.Name, a.App.Status.Expect, a.App.Status.Realtime)

	appType := AppType(a.Type)

	// TODO(ht) consider concurrency
	updateFn := func() {
		err := GetETCDApplications(appType).Update(a.GetName(), a, ctx)
		if err != nil {
			ctx.Application().Logger().Errorf("Got some error: %s", err.Error())
		}
	}
	updateFn()
	defer updateFn()

	switch action {
	case AInstall:
		defer func() {
			err := CallToAgent(ACheck, a, ctx)
			if err != nil {
				ctx.Application().Logger().Errorf("Call to agent to start check failed: <%s>", err)
			}
			ctx.Application().Logger().Info("Call to agent to start check success")
		}()

		if err := InitAgent(a.Host[0].IP, a.Host[0].Auth, ctx); err != nil {
			ctx.Application().Logger().Errorf("Init agent failed: <%s>", err.Error())
			a.App.Status.Realtime = Failed
			return
		}
		// wait the agent starting
		time.Sleep(5 * time.Second)
		if err := CallToAgent(AInstall, a, ctx); err != nil {
			a.App.Status.Realtime = Failed
			return
		}
		a.App.Status.Realtime = Running
	case AStart:
		if err := CallToAgent(AStart, a, ctx); err != nil {
			a.App.Status.Realtime = Failed
			return
		}
		a.App.Status.Realtime = Running
	case AStop:
		if err := CallToAgent(AStop, a, ctx); err != nil {
			a.App.Status.Realtime = Failed
			return
		}
		a.App.Status.Realtime = Stopped
	case ARestart:
		if err := CallToAgent(ARestart, a, ctx); err != nil {
			a.App.Status.Realtime = Failed
			return
		}
		a.App.Status.Realtime = Running
	case AUninstall:
		if err := CallToAgent(AUninstall, a, ctx); err != nil {
			a.App.Status.Realtime = Failed
			return
		}
		a.App.Status.Realtime = NotInstalled
	}
}

func (a *GenericApplication) GetStatus() *Statusx {
	return &a.App.Status
}

func (a *GenericApplication) SetStatus(expect, realtime ApplicationStatus, ctx iris.Context) {
	appType := AppType(a.Type)

	// TODO(ht) consider concurrency
	updateFn := func() {
		err := GetETCDApplications(appType).Update(a.GetName(), a, ctx)
		if err != nil {
			ctx.Application().Logger().Errorf("Got some error: %s", err.Error())
		}

		err = GetETCDApplications(appType).AddChangedApp(a.Name, ctx)
		if err != nil {
			ctx.Application().Logger().Errorf("Got some error: %s", err.Error())
		}
	}
	defer updateFn()

	if len(expect) > 0 {
		a.App.Status.Expect = expect
	}
	if len(realtime) > 0 {
		a.App.Status.Realtime = realtime
	}
}

func (a *GenericApplication) GetName() string {
	return a.Name
}

func (a *GenericApplication) GetApp() *Appx {
	return &a.App
}

func (a *GenericApplication) GetHosts() []Hostx {
	return a.Host
}

func (a *GenericApplication) AddEvent(event map[string]string, ctx iris.Context) (bool, error) {
	return false, nil
}

func (a *GenericApplication) GetEvents() []map[string]string {
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

	localAgentTarPath := filepath.Join(WORK_DIR, AGENT_ZIP_NAME)
	remoteTmpTarPath := filepath.Join(tmpDir, AGENT_ZIP_NAME)

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
	if err := sshCli.UploadFile(localAgentTarPath, remoteTmpTarPath); err != nil {
		ctx.Application().Logger().Error(err)
		return err
	}

	// start agent
	doWithSuCmd := fmt.Sprintf("echo '%s' > dowithsu.sh && chmod +x dowithsu.sh && ./dowithsu.sh %s %s", utils.DoWithSu, agentUser, agentPasswd)
	cmd := fmt.Sprintf("tar -xzvf %s -C %s && %s 'sh %sagent/agent.sh'", remoteTmpTarPath, tmpDir, doWithSuCmd, tmpDir)
	ctx.Application().Logger().Infof("Prepare to exec cmd: %s", cmd)
	result, err := sshCli.ExecCmd(cmd)
	ctx.Application().Logger().Infof("Exec cmd: <%s> get result: <%s>", cmd, result)
	if err != nil {
		ctx.Application().Logger().Errorf("Exec cmd: <%s> get error: <%s>", cmd, err.Error())
		return err
	}
	return nil
}

func CallToAgent(action ApplicationAction, app *GenericApplication, ctx iris.Context) error {
	var agentUrlPrefix = fmt.Sprintf("http://%s:%s/", app.GetHosts()[0].IP, AGENT_PORT)
	var agentUrl = agentUrlPrefix + string(action)

	ctx.Application().Logger().Infof("call to agent: %s", agentUrl)

	var appInfo agent.AppInfo

	appInfo.Name = app.GetName()
	appInfo.Type = app.Type
	appInfo.OperatorIp = OPERATOR_IP
	appInfo.OperatorPort = OPERATOR_PORT
	appInfo.RepoURL = app.GetApp().RepoURL
	appInfo.Install = app.GetApp().Install
	appInfo.Start = app.GetApp().Start
	appInfo.Stop = app.GetApp().Stop
	appInfo.Restart = app.GetApp().Restart
	appInfo.Uninstall = app.GetApp().Uninstall
	appInfo.Check = app.GetApp().Check
	appInfo.Package = app.GetApp().Package
	appInfo.Metadata = app.GetApp().Metadata

	// repo_url and package environment is needed by scripts.
	if _, ok := app.GetApp().Metadata["REPO_URL"]; !ok {
		appInfo.Metadata["REPO_URL"] = appInfo.RepoURL
	}
	if _, ok := app.GetApp().Metadata["PACKAGE"]; !ok {
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

		if !strings.Contains(err.Error(), "connection refused") || !strings.Contains(err.Error(), "timeout") {
			return err
		}

		waitTime := 5 * time.Second
		retry := 5 // 5s;10s;20s;40s;80s
		for i := 0; i < retry; i++ {
			ctx.Application().Logger().Infof("wait for agent start, retry %d/%d", i+1, retry)
			time.Sleep(waitTime)
			waitTime = waitTime * 2
			resp, err = http.Post(agentUrl, "application/json;charset=utf-8", bytes.NewBuffer(jsonBody))
			if err != nil {
				if !strings.Contains(err.Error(), "connection refused") || !strings.Contains(err.Error(), "timeout") {
					return err
				}
				ctx.Application().Logger().Infof("Failed again: %s", err)
				continue
			}
			break
		}
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
