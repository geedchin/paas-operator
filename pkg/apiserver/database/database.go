package database

import (
	"fmt"
	"github.com/kataras/iris"
)

var DatabaseList = make(map[string]Database, 0)

type Database struct {
	// mysql-5.7-xxx-192.168.19.100
	Name string `json:"name"`
	Host []struct {
		IP       string `json:"ip"`
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"host"`
	App struct {
		RepoURL   string            `json:"repo_url"`  // http://192.168.19.101:8080/
		Install   string            `json:"install"`   // install.sh
		Start     string            `json:"start"`     // start.sh
		Stop      string            `json:"stop"`      // stop.sh
		Restart   string            `json:"restart"`   //restart.sh
		Uninstall string            `json:"uninstall"` // uninstall.sh
		Package   string            `json:"package"`   // mysql-5.7.tar.gz
		Metadata  map[string]string `json:"metadata"`
		Status    struct {
			Expect   string `json:"expect"`   // running
			Realtime string `json:"realtime"` // failed
		} `json:"status"`
	} `json:"app"`
}

func (d *Database) Create(ctx iris.Context) {
	var dataBase Database

	if err := ctx.ReadJSON(&dataBase); err != nil {
		fmt.Println("invalid json from the request body !")
		ctx.StatusCode(iris.StatusBadRequest)
		return
	}

	//if err := upload_agent(dataBase.Host[0].Username,
	//	dataBase.Host[0].Password, dataBase.Host[0].IP); err != nil {
	//	fmt.Printf("upload agent to remote host(%s) failed ", dataBase.Host[0].IP)
	//	ctx.StatusCode(iris.StatusInternalServerError)
	//	return
	//}

	// 1. validate ssh connection

	// 2. distribute agent

	// 3. call api - install

	// 4.
}

// action can be [start / stop / install / restart / uninstall]
func (d *Database) UpdateStatus(ctx iris.Context) {
	//switch action {
	//case "start":
	//
	//}

}

func (d *Database) Status(ctx iris.Context) {

}
