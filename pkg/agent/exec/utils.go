package exec

import (
	"github.com/golang/glog"
	"math/rand"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
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

func getID(user, uorg string) (string, error) {
	cmd := exec.Command("id", uorg, user)

	err := cmd.Run()
	if err != nil {
		glog.Errorln(err)
		return "", err
	}
	stdout, err := cmd.Output()
	if err != nil {
		glog.Errorln(err)
		return "", err
	}

	return string(stdout), nil
}

func convertID(id string) (uint32, error) {
	uid64, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		glog.Errorln(err)
		return 0, err
	}

	return uint32(uid64), nil
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
	uid, err := getID(user, USERTYPE)
	if err != nil {
		return err
	}
	uid32, err := convertID(uid)
	if err != nil {
		return err
	}
	gid, err := getID(user, GROUPTYPE)
	if err != nil {
		return err
	}
	gid32, err := convertID(gid)
	if err != nil {
		return err
	}
	cmd := exec.Command("/bin/sh", filePath, args)
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uid32, Gid: gid32}
	return cmd.Run()
}
