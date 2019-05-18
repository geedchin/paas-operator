package agent

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func NewGinEngine() *gin.Engine {
	r := gin.Default()

	// for test only
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// action can only be [ install, start, stop, restart, uninstall ]
	r.POST("/:action", DoAction)
	return r
}

// DoAction knows how to judge the script to be executed
// according to the Action

// if ok, return 200 and {"msg":"ok"}
// if some error occer, return not200 and {"error":"detail info"}
func DoAction(c *gin.Context) {
	// validate action
	action := c.Param("action")
	if _, ok := ActionMap[Action(action)]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "action can't be " + action,
		})
		return
	}

	// apply request body to appInfo
	var appInfo AppInfo
	if err := c.ShouldBindJSON(&appInfo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	log.Println("Action: " + action)
	log.Println("AppInfo: " + appInfo.Print())

	// doAction get the Action & AppInfo, then exec a corresponding script.
	doAction := func(action Action, appInfo *AppInfo) error {
		// eg. [ install.sh, start.sh, ... ]
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

		//validate scriptName
		if len(scriptName) < 1 {
			return errors.New("script name illegal: " + scriptName)
		}

		// prepare the script
		scriptPath, err := getScriptIfNotExist(scriptName, repoUrl)
		if err != nil {
			log.Println("wget script failed: " + err.Error())
			return err
		}

		// prepare args for script
		var args string
		for k, v := range appInfo.Metadata {
			args += k + "=" + v + " "
		}

		// exec the script
		err = execInLinux("sh", WorkDir, []string{scriptPath, args})
		if err != nil {
			return err
		}
		return nil
	}

	err := doAction(Action(action), &appInfo)
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
	scriptPath := filepath.Join(WorkDir, scriptName)
	exist, err := pathExists(scriptPath)
	if err != nil {
		return "", err
	}

	if exist {
		return scriptPath, nil
	}

	// file not exist, do wget

	// ensure the workdir is exist
	err = os.MkdirAll(WorkDir, os.ModePerm)
	if err != nil {
		return "", err
	}
	err = execInLinux("wget", WorkDir, []string{repoUrl + scriptName})
	if err != nil {
		log.Println("Wget Failed!!!")
		return "", err
	}
	return scriptPath, nil
}

// exist -> true; else -> false; if some error occur, return the err
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

// execInLinux can exec a command with some params in linux system
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

	err = cmd.Start()
	if err != nil {
		return err
	}
	return cmd.Wait()
}
