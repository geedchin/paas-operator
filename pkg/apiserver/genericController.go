package apiserver

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/kataras/iris"

	"github.com/farmer-hutao/k6s/pkg/apiserver/database"
)

func CreateDatabase(ctx iris.Context) {
	var db database.GenericDatabase

	// apply request body to db
	if err := ctx.ReadJSON(&db); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString(err.Error())
		ctx.Application().Logger().Errorf("CreateDatabase Error, json is illegal: %s", err)
		return
	}

	// validate database's name
	if len(db.GetName()) < 1 {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString("CreateDatabase Error, GenericDatabase name is illegal")
		ctx.Application().Logger().Error("CreateDatabase Error, GenericDatabase name is illegal")
		return
	}

	// validate database is already exist
	if _, ok := database.GetETCDDatabases().Get(db.GetName(), ctx); ok {
		ctx.StatusCode(iris.StatusBadRequest)
		msg := fmt.Sprintf("CreateDatabase Failed, the db with name <%s> is already exist.", db.GetName())
		ctx.WriteString(msg)
		ctx.Application().Logger().Error(msg)
		return
	}

	// init database status if it is empty
	if db.GetApp().Status.Expect == "" {
		db.GetApp().Status.Expect = database.NotInstalled
	}
	if db.GetApp().Status.Realtime == "" {
		db.GetApp().Status.Realtime = database.NotInstalled
	}

	db.GetApp().Metadata["CreateAt"] = time.Now().Format("2006-01-02 15:04:05")

	// add a database to etcd databases
	err := database.GetETCDDatabases().Add(db.GetName(), &db, ctx)
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
	//db, ok := database.GetMemoryDatabases().Get(dbName)
	db, ok := database.GetETCDDatabases().Get(dbName, ctx)
	if !ok {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString("d_name is not exist: " + dbName)
		ctx.Application().Logger().Errorf("d_name is not exist: %s", dbName)
		return
	}

	// validate whether the expect status is illegal
	if _, ok := database.DatabaseStatusMap[expectStatus]; !ok {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString("database status is illegal: " + status)
		ctx.Application().Logger().Errorf("database status is illegal: " + status)
		return
	}

	ctx.Application().Logger().Infof("UpdataDatabaseStatus: the database with name <%s> expect status is <%s> "+
		"and realtime status is <%s>;", db.GetName(), db.GetStatus().Expect, db.GetStatus().Realtime)

	// update db resource status
	switch expectStatus {

	// TODO(ht): uninstall
	case database.NotInstalled: // uninstall a database

	case database.Running: // install or start a database
		// install a database
		if db.GetStatus().Expect == database.NotInstalled {
			db.GetApp().Status.Expect = database.Running
			db.GetApp().Status.Realtime = database.Installing
			// update db's real status
			go db.UpdateStatus(database.AInstall, ctx)

			// start a database
		} else if db.GetStatus().Expect == database.Stopped {
			db.GetApp().Status.Expect = database.Running
			db.GetApp().Status.Realtime = database.Starting
			// update db's real status
			go db.UpdateStatus(database.AStart, ctx)
		}
		// stop a database
	case database.Stopped:
		if db.GetStatus().Expect == database.Running {
			db.GetApp().Status.Expect = database.Stopped
			db.GetApp().Status.Realtime = database.Stopping
			// update db's real status
			go db.UpdateStatus(database.AStop, ctx)
		}
		// restart a database
	case database.Restart:
		if db.GetStatus().Expect == database.Running {
			db.GetApp().Status.Expect = database.Running
			db.GetApp().Status.Realtime = database.Restarting
			// update db's real status
			go db.UpdateStatus(database.ARestart, ctx)
		}
	}

	ctx.StatusCode(iris.StatusAccepted)
	_, _ = ctx.JSON(iris.Map{
		"name":   db.GetName(),
		"status": db.GetStatus(),
	})
	return
}

func GetDatabaseStatus(ctx iris.Context) {
	dbName := ctx.Params().GetString("d_name")
	//db, ok := database.GetMemoryDatabases().Get(dbName)
	db, ok := database.GetETCDDatabases().Get(dbName, ctx)
	if !ok {
		ctx.StatusCode(iris.StatusBadRequest)
		msg := fmt.Sprintf("GenericDatabase with name <%s> is not exist: ", dbName)
		ctx.WriteString(msg)
		ctx.Application().Logger().Error(msg)
		return
	}

	status := db.GetStatus()

	_, err := ctx.JSON(iris.Map{
		"name":   db.GetName(),
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

func DeleteDatabase(ctx iris.Context) {
	dbName := ctx.Params().GetString("d_name")
	ctx.Application().Logger().Infof("Prepare to delete a db named <%s>", dbName)

	db, err := database.GetETCDDatabases().Delete(dbName, ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.WriteString("got some error")
		ctx.Application().Logger().Errorf("Failed to delete a db <%s>.", dbName)
		return
	}

	if db == nil {
		ctx.StatusCode(iris.StatusOK)
		ctx.WriteString("db not exist")
		return
	}

	ctx.StatusCode(iris.StatusOK)
	ctx.WriteString(db.GetName())
}
