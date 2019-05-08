package apiserver

import (
	"fmt"
	"github.com/farmer-hutao/k6s/pkg/apiserver/database"
	"github.com/golang/glog"
	"github.com/kataras/iris"
	"time"
)

func CreateDatabase(ctx iris.Context) {
	var db database.Database

	if err := ctx.ReadJSON(&db); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString(err.Error())
		glog.Error("Got request json error: " + err.Error())
		return
	}

	if len(db.Name) < 1 {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString("Name must not nil, got: " + db.Name)
		glog.Error("Name must not nil, got: " + db.Name)
		return
	}

	db.App.Status.Expect = database.NotInstalled
	db.App.Status.Realtime = database.NotInstalled
	db.App.Metadata["CreateAt"] = time.Now().Format("2006-01-02 15:04:05")

	err := database.DatabaseList.Add(db.Name, db)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString(err.Error())
		glog.Error("Create database failed: ", err)
		return
	}

	ctx.StatusCode(iris.StatusCreated)
	ctx.JSON(db)
}

func UpdateDatabaseStatus(ctx iris.Context) {
	dbName := ctx.Params().GetString("d_name")
	action := ctx.Params().GetString("action")

	// validate dbname
	if len(dbName) < 1 {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString("Name must not nil, got: " + dbName)
		glog.Error("Name must not nil, got: " + dbName)
		return
	}

	// validate whether the db exists
	db, ok := database.DatabaseList.Get(dbName)
	if !ok {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString("d_name is not exist: " + dbName)
		glog.Error("d_name is not exist: " + dbName)
		return
	}

	// validate whether the action is illegal
	if _, ok := database.DatabaseActionMap[database.DatabaseAction(action)]; !ok {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString("database action is illegal: " + action)
		glog.Error("database action is illegal: " + action)
		return
	}

	dbAction := database.DatabaseAction(action)
	// TODO call agent
	switch dbAction {
	case database.Start:
		db.App.Status.Expect = database.Running
	case database.Stop:
		db.App.Status.Expect = database.Stoped
	case database.Install:
		db.App.Status.Expect = database.Running
	case database.Restart:
		// TODO(ht): how to restart?
	}

	ctx.StatusCode(iris.StatusAccepted)
	ctx.JSON(iris.Map{
		"name":   db.Name,
		"status": db.App.Status,
	})
	return
}

func GetDatabaseStatus(ctx iris.Context) {
	dbName := ctx.Params().GetString("d_name")
	db, ok := database.DatabaseList.Get(dbName)
	if !ok {
		ctx.StatusCode(iris.StatusBadRequest)
		msg := fmt.Sprintf("Database with name <%s> is not exist: ", dbName)
		ctx.WriteString(msg)
		glog.Error(msg)
		return
	}

	status := db.App.Status

	_, err := ctx.JSON(iris.Map{
		"name":   db.Name,
		"status": status,
	})
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.WriteString("got some error")
		glog.Error("get db status failed: " + dbName)
		return
	}

	ctx.StatusCode(iris.StatusOK)
	return
}
