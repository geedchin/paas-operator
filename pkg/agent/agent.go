package agent

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

const (
	WORK_DIR = "/opt/app/"
)

type AppInfo struct {
	RepoURL   string            `json:"repo_url"`
	Install   string            `json:"install"`
	Start     string            `json:"start"`
	Stop      string            `json:"stop"`
	Restart   string            `json:"restart"`
	Uninstall string            `json:"uninstall"`
	Package   string            `json:"package"`
	Metadata  map[string]string `json:"metadata"`
}

func (ai *AppInfo) Print() string {
	bytes, err := json.MarshalIndent(ai, "", " ")
	if err != nil {
		log.Println(err.Error())
		return err.Error()
	}
	return string(bytes)
}

type Action string

const (
	Install   Action = "install"
	Start     Action = "start"
	Stop      Action = "stop"
	Restart   Action = "restart"
	Uninstall Action = "uninstall"
)

var ActionMap = map[Action]struct{}{
	Install:   {},
	Start:     {},
	Stop:      {},
	Restart:   {},
	Uninstall: {},
}

func NewGinEngine() *gin.Engine {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.POST("/:action", DoAction)
	return r
}

func DoAction(c *gin.Context) {
	// validate action
	action := c.Param("action")
	if _, ok := ActionMap[Action(action)]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "action can't be " + action,
		})
		return
	}

	// apply app info
	var appInfo AppInfo
	if err := c.ShouldBindJSON(&appInfo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	log.Println("Action: " + action)
	log.Println("AppInfo: " + appInfo.Print())

	do := func(action Action, appInfo *AppInfo) error {
		// eg. install.sh
		var scriptName string
		var repoUrl = appInfo.RepoURL
		switch action {
		case Install:
			scriptName = appInfo.Install
		case Start:
			scriptName = appInfo.Start
		case Stop:
			scriptName = appInfo.Stop
		case Restart:
			scriptName = appInfo.Restart
		case Uninstall:
			scriptName = appInfo.Uninstall
		}

		//validate the len(name) with script > 0
		if len(scriptName) < 1 {
			return errors.New("script name illegal: " + scriptName)
		}

		scriptPath, err := getScriptIfNotExist(scriptName, repoUrl)
		if err != nil {
			log.Println("wget script failed: " + err.Error())
			return err
		}

		// add args to script
		var args string
		for k, v := range appInfo.Metadata {
			args += k + "=" + v + " "
		}

		err = execInLinux("sh", WORK_DIR, []string{scriptPath, args})
		if err != nil {
			return err
		}
		return nil
	}

	err := do(Action(action), &appInfo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"msg": "ok",
	})
}

// local dir: /opt/app/; local script: /opt/app/xxx.sh
// origin script: http://xxx:nn/xxx/xxx.sh
// return /opt/app/xxx.sh, error
func getScriptIfNotExist(scriptName, repoUrl string) (string, error) {
	scriptPath := filepath.Join(WORK_DIR,scriptName)
	exist, err := pathExists(scriptPath)
	if err != nil {
		return "", err
	}

	if exist {
		return scriptPath, nil
	}

	// file not exist, do wget

	// if workdir is not exist
	err = os.MkdirAll(WORK_DIR, os.ModePerm)
	if err != nil {
		return "", err
	}
	err = execInLinux("wget", WORK_DIR, []string{repoUrl + scriptName})
	if err != nil {
		return "", err
	}
	return scriptPath, nil
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func execInLinux(cmdName, execPath string, params []string) error {
	cmd := exec.Command(cmdName, params...)
	cmd.Dir = execPath

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	// print log
	outReader := bufio.NewReader(stdout)
	errReader := bufio.NewReader(stderr)
	printLog := func(reader *bufio.Reader) {
		for {
			line, err := outReader.ReadString('\n')
			if err != nil || err == io.EOF {
				break
			}
			log.Println(line)
		}
	}
	go printLog(outReader)
	go printLog(errReader)

	cmd.Start()
	return cmd.Wait()
}
