package agent

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/farmer-hutao/k6s/pkg/apiserver/utils"
	"github.com/gin-gonic/gin"
)

var once = &sync.Once{}

func NewGinEngine() *gin.Engine {
	r := gin.New()

	// LoggerWithFormatter middleware will write the logs to gin.DefaultWriter
	// By default gin.DefaultWriter = os.Stdout
	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {

		// your custom format
		return fmt.Sprintf("[%s] \"%s %s %d %s \n",
			param.TimeStamp.Format(time.StampMilli),
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency,
		)
	}))
	r.Use(gin.Recovery())

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
	doAction := func(action Action, appInfo *AppInfo) (*bytes.Buffer, error) {
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
		case Check:
			scriptName = appInfo.Check
		}

		//validate scriptName
		if len(scriptName) < 1 {
			return nil, errors.New("script name illegal: " + scriptName)
		}

		// prepare the script
		scriptPath, err := getScriptIfNotExist(scriptName, repoUrl)
		if err != nil {
			log.Println("Get script failed: " + err.Error())
			return nil, err
		}

		// prepare args for script
		var args string
		for k, v := range appInfo.Metadata {
			args += k + "=" + v + " "
		}

		if action == Check {
			check(appInfo.Name, appInfo.Type, appInfo.OperatorIp, appInfo.OperatorPort, WorkDir, scriptPath, args)
			return nil, nil
		}

		// exec the script
		err = execInLinux("sh", WorkDir, []string{scriptPath, args}, nil, true)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}

	_, err := doAction(Action(action), &appInfo)
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

type CheckArg struct {
	Name         string `json:"name"`
	AppType      string `json:"apptype"`
	OperatorIp   string `json:"operatorip"`
	OperatorPort string `json:"operatorport"`
	WorkDir      string `json:"workdir"`
	ScriptPath   string `json:"scriptpath"`
	Args         string `json:"args"`
}

func check(name, appType, operatorIp, operatorPort, workdir, scriptPath, args string) {
	// save all func args to checkInfo.json
	var ca = CheckArg{
		Name:         name,
		AppType:      appType,
		OperatorIp:   operatorIp,
		OperatorPort: operatorPort,
		WorkDir:      workdir,
		ScriptPath:   scriptPath,
		Args:         args,
	}
	caBytes, err := json.Marshal(ca)
	if err != nil {
		log.Printf("marshal checkInfo.json failed: %s", err)
	} else {
		err = ioutil.WriteFile(filepath.Join(WorkDir, "checkInfo.json"), caBytes, 0666)
		if err != nil {
			log.Printf("write checkInfo.json failed: %s", err)
		}
	}

	// prepare to check
	period := 5 * time.Second
	var c = &http.Client{}

	report := func(msg string) {
		// trim "xxx{xxx}xxx" to "{xxx}"
		trimMsg := func(msg string) string {
			start := strings.Index(msg, "{")
			end := strings.LastIndex(msg, "}")
			if start < 0 || end < 0 {
				return msg
			}
			return msg[start : end+1]
		}
		msg = trimMsg(msg)

		if !utils.ValidateAppHealthyJson(msg) {
			log.Printf("Error: Json illeagel:<%s>", msg)
			return
		}

		url := fmt.Sprintf("http://%s:%s/apis/v1alpha1/%s/%s/check", operatorIp, operatorPort, appType, name)
		payload := bytes.NewBufferString(msg)
		req, err := http.NewRequest("PUT", url, payload)
		if err != nil {
			log.Printf("Error: NewRequest failed: %s", err)
			return
		}
		req.Header.Add("Content-Type", "application/json;charset=utf-8")
		resp, err := c.Do(req)
		if err != nil {
			log.Printf("Error: Do Request failed: %s", err)
			return
		}
		if resp.StatusCode != http.StatusAccepted {
			log.Printf("Error: %s", resp.Status)
			if period < 1*time.Minute {
				period += 5 * time.Second
			}
			return
		}
		period = 5 * time.Second
	}

	var buf bytes.Buffer
	once.Do(func() {
		go func() {
			for {
				err := execInLinux("sh", workdir, []string{scriptPath, args}, &buf, false)
				if err != nil {
					if period < 1*time.Hour {
						period *= 2
					}
					log.Printf("Exec check cmd failed: %s, wait %ds", err, period/time.Second)
				} else {
					report(buf.String())
				}
				buf.Reset()
				time.Sleep(period)
			}
		}()
	})
}

func TryCheck() {
	infoBytes, err := ioutil.ReadFile(filepath.Join(WorkDir, "checkInfo.json"))
	if err != nil {
		log.Printf("read checkInfo.json failed: %s", err)
		return
	}
	var ca CheckArg
	err = json.Unmarshal(infoBytes, &ca)
	if err != nil {
		log.Printf("Unmarshal checkInfo.json failed: %s", err)
		return
	}
	check(ca.Name, ca.AppType, ca.OperatorIp, ca.OperatorPort, ca.WorkDir, ca.ScriptPath, ca.Args)
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
	err = execInLinux("wget", WorkDir, []string{repoUrl + scriptName}, nil, true)
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
// all log producted by script would be print to stdout and return at logsBuffer if logsBuffer is not nil
func execInLinux(cmdName, execPath string, params []string, logsBuffer *bytes.Buffer, print bool) error {
	var lock sync.Mutex
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
	printLog := func(reader *bufio.Reader, typex string) {
		for {
			line, err := reader.ReadString('\n')
			if print {
				log.Printf("%s: %s", typex, line)
			}
			if logsBuffer != nil {
				lock.Lock()
				logsBuffer.WriteString(line)
				lock.Unlock()
			}
			if err != nil || err == io.EOF {
				break
			}
		}
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		printLog(outReader, "Stdout")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		printLog(errReader, "Stderr")
	}()

	err = cmd.Start()
	if err != nil {
		return err
	}

	wg.Wait()
	return cmd.Wait()
}
