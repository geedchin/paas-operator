package apiserver

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang/glog"
	"github.com/kataras/iris"

	"github.com/farmer-hutao/k6s/pkg/apiserver/database"
)

func CreateDatabase(ctx iris.Context) {
	var db database.GenericDatabase

	// use request body init database
	if err := ctx.ReadJSON(&db); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString(err.Error())
		ctx.Application().Logger().Errorf("CreateDatabase Error, json is illegal: %s", err)
		return
	}

	// validate database's name
	if len(db.Name()) < 1 {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString("CreateDatabase Error, GenericDatabase name is illegal")
		ctx.Application().Logger().Error("CreateDatabase Error, GenericDatabase name is illegal")
		return
	}

	// init database status if it is empty
	if db.App().Status.Expect == "" {
		db.App().Status.Expect = database.NotInstalled
	}
	if db.App().Status.Realtime == "" {
		db.App().Status.Realtime = database.NotInstalled
	}

	db.App().Metadata["CreateAt"] = time.Now().Format("2006-01-02 15:04:05")

	err := database.GetMemoryDatabases().Add(db.Name(), &db)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.WriteString(err.Error())
		ctx.Application().Logger().Errorf("Add db to database list failed: %s", err)
		return
	}

	ctx.StatusCode(iris.StatusCreated)
	ctx.JSON(db)

	dbBytes, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		ctx.Application().Logger().Errorf("Json Marshal database failed: %s", err.Error())
		return
	}
	ctx.Application().Logger().Infof("Created a database: %s", string(dbBytes))
}

func UpdateDatabaseStatus(ctx iris.Context) {
	dbName := ctx.Params().GetString("d_name")
	status := ctx.Params().GetString("status")
	expectStatus := database.DatabaseStatus(status)

	ctx.Application().Logger().Infof("UpdataDatabaseStatus: got d_name <%s> and expect status <%s>;", dbName, status)

	// validate dbname
	if len(dbName) < 1 {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString("CreateDatabase Error, GenericDatabase name is illegal")
		ctx.Application().Logger().Error("CreateDatabase Error, GenericDatabase name is illegal")
		return
	}

	// validate whether the db exists
	db, ok := database.GetMemoryDatabases().Get(dbName)
	if !ok {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString("d_name is not exist: " + dbName)
		glog.Error("d_name is not exist: " + dbName)
		return
	}

	// validate whether the expect status is illegal
	if _, ok := database.DatabaseStatusMap[expectStatus]; !ok {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString("database status is illegal: " + status)
		glog.Error("database status is illegal: " + status)
		return
	}

	ctx.Application().Logger().Infof("UpdataDatabaseStatus: the database with name <%s> expect status is <%s> "+
		"and realtime status is <%s>;", db.Name(), db.Status().Expect, db.Status().Realtime)

	// update db resource status
	switch expectStatus {

	// TODO(ht): uninstall
	case database.NotInstalled: // uninstall a database

	case database.Running: // install or start a database
		// install a database
		if db.Status().Expect == database.NotInstalled {
			db.App().Status.Expect = database.Running
			db.App().Status.Realtime = database.Installing
			// update db's real status
			go db.UpdateStatus(database.AInstall, ctx)

			// start a database
		} else if db.Status().Expect == database.Stopped {
			db.App().Status.Expect = database.Running
			db.App().Status.Realtime = database.Starting
			// update db's real status
			go db.UpdateStatus(database.AStart, ctx)
		}
		// stop a database
	case database.Stopped:
		if db.Status().Expect == database.Running {
			db.App().Status.Expect = database.Stopped
			db.App().Status.Realtime = database.Stopping
			// update db's real status
			go db.UpdateStatus(database.AStop, ctx)
		}
		// restart a database
	case database.Restart:
		if db.Status().Expect == database.Running {
			db.App().Status.Expect = database.Running
			db.App().Status.Realtime = database.Restarting
			// update db's real status
			go db.UpdateStatus(database.ARestart, ctx)
		}
	}

	ctx.StatusCode(iris.StatusAccepted)
	_, _ = ctx.JSON(iris.Map{
		"name":   db.Name(),
		"status": db.Status(),
	})
	return
}

func GetDatabaseStatus(ctx iris.Context) {
	dbName := ctx.Params().GetString("d_name")
	db, ok := database.GetMemoryDatabases().Get(dbName)
	if !ok {
		ctx.StatusCode(iris.StatusBadRequest)
		msg := fmt.Sprintf("GenericDatabase with name <%s> is not exist: ", dbName)
		ctx.WriteString(msg)
		ctx.Application().Logger().Error(msg)
		return
	}

	status := db.Status()

	_, err := ctx.JSON(iris.Map{
		"name":   db.Name(),
		"status": status,
	})
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.WriteString("got some error")
		ctx.Application().Logger().Errorf("get db status failed: %s" + dbName)
		return
	}

	ctx.StatusCode(iris.StatusOK)
	return
}
