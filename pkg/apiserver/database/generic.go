package database

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/kataras/iris"

	"github.com/farmer-hutao/k6s/pkg/agent"
	"github.com/farmer-hutao/k6s/pkg/apiserver/utils/sshcli"
)

const (
	WORK_DIR       = "/opt/app/"
	AGENT_ZIP_NAME = "agent.tar.gz"
)

// GenericDatabase is a generic implement of Database interface
type GenericDatabase struct {
	// mysql-5.7-xxx-192.168.19.100
	Name string  `json:"name"`
	Host []Hostx `json:"host"`
	App  Appx    `json:"app"`
}

func (d *GenericDatabase) UpdateStatus(action DatabaseAction, ctx iris.Context) {
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

func InitAgent(ip, username, password string, ctx iris.Context) error {
	ctx.Application().Logger().Info("start to init agent!!!")

	agentTarPath := filepath.Join(WORK_DIR, AGENT_ZIP_NAME)
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
	cmd := fmt.Sprintf("tar -xzvf %s -C %s && sh %sagent/agent.sh", agentTarPath, WORK_DIR, WORK_DIR)
	result, err := sshCli.ExecCmd(cmd)
	ctx.Application().Logger().Infof("Exec cmd: <%s> get result: <%s>", cmd, result)
	if err != nil {
		ctx.Application().Logger().Errorf("Exec cmd: <%s> get error: <%s>", cmd, err.Error())
		return err
	}
	return nil
}

func CallToAgent(action DatabaseAction, db *GenericDatabase, ctx iris.Context) error {
	var agentUrlPrefix = "http://" + db.GetHosts()[0].IP + ":3335/"
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

	if _, ok := db.GetApp().Metadata["REPO_URL"]; !ok {
		appInfo.Metadata["REPO_URL"] = appInfo.RepoURL
	}

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
