package agent

import (
	"encoding/json"
	"log"
	"os"
)

type Action string

const (
	Install   Action = "install"
	Start     Action = "start"
	Stop      Action = "stop"
	Restart   Action = "restart"
	Uninstall Action = "uninstall"
	Check     Action = "check"
)

var ActionMap = map[Action]struct{}{
	Install:   {},
	Start:     {},
	Stop:      {},
	Restart:   {},
	Uninstall: {},
	Check:     {},
}

var WorkDir = "/opt/app/"

func init() {
	// TODO This environment is in the vm
	if os.Getenv("AGENT_WORK_DIR") != "" {
		WorkDir = os.Getenv("AGENT_WORK_DIR")
	}
}

// AppInfo include all info the agent will use to control a app.
type AppInfo struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	OperatorIp   string `json:"operator_ip"`
	OperatorPort string `json:"operator_port"`
	RepoURL      string `json:"repo_url"`
	Install      string `json:"install"`
	Start        string `json:"start"`
	Stop         string `json:"stop"`
	Restart      string `json:"restart"`
	Uninstall    string `json:"uninstall"`
	Check        string `json:"check"`
	Package      string `json:"package"`
	// all metadata will inject to script as a param, like:
	// for k, v := range appInfo.Metadata {
	//	  args += k + "=" + v + " "
	// }
	// sh xxx.sh args
	Metadata map[string]string `json:"metadata"`
}

// Print return a string desc with AppInfo; If some error occur, return err.Error()
func (ai *AppInfo) Print() string {
	bytes, err := json.MarshalIndent(ai, "", " ")
	if err != nil {
		log.Println(err.Error())
		return err.Error()
	}
	return string(bytes)
}
