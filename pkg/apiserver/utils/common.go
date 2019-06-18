package utils

// TODO 结果判断
var DoWithSu = `#!/usr/bin/expect -f
spawn -noecho su [lindex $argv 0] -c [lindex $argv 2]
set password [lindex $argv 1]
expect "*assword:"
send "$password\r"
set timeout 60
expect eof
exit`
