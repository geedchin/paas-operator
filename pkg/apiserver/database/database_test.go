package database

import (
	"fmt"
	"net/http"
	"testing"
)

var (
	agentIp       = "192.168.19.100"
	agentHostname = "root"
	agentPassword = "xxx"
	agentPort     = "3335"
)

func TestDatabase_UpdateStatus(t *testing.T) {

}

func TestInitAgent(t *testing.T) {
	hostIp := agentIp
	username := agentHostname
	password := agentPassword
	err := InitAgent(hostIp, username, password)
	if err != nil {
		t.Error(err)
	}

	hostIpError := "xxx"
	err = InitAgent(hostIpError, username, password)
	if err == nil {
		t.Error("err should be not nil!!!")
	}

	passwordError := "xxx"
	err = InitAgent(hostIp, username, passwordError)
	if err == nil {
		t.Error("err should be not nil!!!")
	}
}

func TestCallToAgent(t *testing.T) {

}

// agent is alive
func TestCallToAgent2(t *testing.T) {
	resp, err := http.Get(fmt.Sprintf("http://%s:%s/ping", agentIp, agentPort))
	if err != nil {
		t.Error(err)
	}
	if resp.StatusCode != 200 {
		t.Error(resp.Status)
	}
}
