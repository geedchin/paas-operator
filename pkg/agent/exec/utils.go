package exec

import (
	"fmt"
	"math/rand"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	FILEPATH        = "/tmp/"
	INSTALL         = "install.sh"
	START           = "start.sh"
	RESTART         = "restart.sh"
	STOP            = "stop.sh"
	DELETE          = "delete.sh"
	INSTALLFILENAME = FILEPATH + INSTALL
	STARTFILENAME   = FILEPATH + START
	RESTARTFILENAME = FILEPATH + RESTART
	STOPFILENAME    = FILEPATH + STOP
	DELETEFILENAME  = FILEPATH + DELETE
	SEP             = "/"
	USERTYPE        = "-u"
	GROUPTYPE       = "-g"
)

func splitUrl(url string) string {
	strArr := strings.Split(url, SEP)
	return strArr[len(strArr)-1]
}

func wgetScript(url string) (string, error) {
	rand.Seed(time.Now().UnixNano())
	r := rand.Int()
	rs := strconv.Itoa(r)
	fileName := FILEPATH + rs + splitUrl(url)
	cmd := exec.Command("wget", url, "-O", fileName)
	return fileName, cmd.Run()
}

func execScript(user, filePath, args string) error {
	command := fmt.Sprintf("/usr/bin/sh %s %s", filePath, args)
	cmd := exec.Command("/usr/bin/su", "-", user, "-c", command)

	return cmd.Run()
}
