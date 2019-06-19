package utils

import (
	"encoding/json"
)

// TODO 结果判断
var DoWithSu = `#!/usr/bin/expect -f
spawn -noecho su [lindex $argv 0] -c [lindex $argv 2]
set password [lindex $argv 1]
expect "*assword:"
send "$password\r"
set timeout 60
expect eof
exit`

//{
//	  "code": "0",
//	  "msg": "some message"
//}
type AppHealthy struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
}

func ValidateAppHealthyJson(jsonStr string) bool {
	var ah AppHealthy
	err := json.Unmarshal([]byte(jsonStr), &ah)
	return err == nil
}
