package apiserver

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/kataras/iris"

	"github.com/farmer-hutao/k6s/pkg/apiserver/application"
	"github.com/farmer-hutao/k6s/pkg/apiserver/utils"
)

func CreateDatabase(ctx iris.Context) {
	appType := application.APP_DATABASE
	createApplication(appType, ctx)
}

func UpdateDatabaseStatus(ctx iris.Context) {
	appType := application.APP_DATABASE
	updateApplicationStatus(appType, ctx)
}

func GetDatabaseStatus(ctx iris.Context) {
	appType := application.APP_DATABASE
	getApplicationStatus(appType, ctx)
}

func DeleteDatabase(ctx iris.Context) {
	appType := application.APP_DATABASE
	deleteApplication(appType, ctx)
}

func SetDatabaseRealtimeStatus(ctx iris.Context) {
	appType := application.APP_DATABASE
	setApplicationRealtimeStatus(appType, ctx)
}

func CreateMiddleware(ctx iris.Context) {
	appType := application.APP_MIDDLEWARE
	createApplication(appType, ctx)
}

func UpdateMiddlewareStatus(ctx iris.Context) {
	appType := application.APP_MIDDLEWARE
	updateApplicationStatus(appType, ctx)
}

func GetMiddlewareStatus(ctx iris.Context) {
	appType := application.APP_MIDDLEWARE
	getApplicationStatus(appType, ctx)
}

func DeleteMiddleware(ctx iris.Context) {
	appType := application.APP_MIDDLEWARE
	deleteApplication(appType, ctx)
}

func SetMiddlewareRealtimeStatus(ctx iris.Context) {
	appType := application.APP_MIDDLEWARE
	setApplicationRealtimeStatus(appType, ctx)
}

func createApplication(appType application.AppType, ctx iris.Context) {
	var app application.GenericApplication

	// apply request body to app
	if err := ctx.ReadJSON(&app); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString(err.Error())
		ctx.Application().Logger().Errorf("CreateApplication Error, json is illegal: %s", err)
		return
	}

	// validate application's name
	if len(app.GetName()) < 1 {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString("CreateApplication Error, GenericApplication name is illegal")
		ctx.Application().Logger().Error("CreateApplication Error, GenericApplication name is illegal")
		return
	}

	// validate application is already exist
	if _, ok := application.GetETCDApplications(appType).Get(app.GetName(), ctx); ok {
		ctx.StatusCode(iris.StatusBadRequest)
		msg := fmt.Sprintf("CreateApplication Failed, the app with name <%s> is already exist.", app.GetName())
		ctx.WriteString(msg)
		ctx.Application().Logger().Error(msg)
		return
	}

	// init application type
	app.Type = string(appType)

	// init application status if it is empty
	if app.GetApp().Status.Expect == "" {
		app.GetApp().Status.Expect = application.NotInstalled
	}
	if app.GetApp().Status.Realtime == "" {
		app.GetApp().Status.Realtime = application.NotInstalled
	}

	app.GetApp().Metadata["CreateAt"] = time.Now().Format("2006-01-02 15:04:05")

	// add a application to etcd applications
	err := application.GetETCDApplications(appType).Add(app.GetName(), &app, ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.WriteString(err.Error())
		ctx.Application().Logger().Errorf("Add app to application list failed: %s", err)
		return
	}

	ctx.StatusCode(iris.StatusCreated)
	ctx.JSON(app)

	appBytes, err := json.MarshalIndent(app, "", "  ")
	if err != nil {
		ctx.Application().Logger().Errorf("Json Marshal application failed: %s", err.Error())
		return
	}
	ctx.Application().Logger().Infof("Created a application: %s", string(appBytes))
}

func updateApplicationStatus(appType application.AppType, ctx iris.Context) {
	appName := ctx.Params().GetString("a_name")
	status := ctx.Params().GetString("status")
	expectStatus := application.ApplicationStatus(status)

	ctx.Application().Logger().Infof("UpdataApplicationStatus: got a_name <%s> and expect status <%s>;", appName, status)

	// validate appname
	if len(appName) < 1 {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString("CreateApplication Error, GenericApplication name is illegal")
		ctx.Application().Logger().Error("CreateApplication Error, GenericApplication name is illegal")
		return
	}

	// validate whether the app exists
	//app, ok := application.GetMemoryApplications().Get(appName)
	app, ok := application.GetETCDApplications(appType).Get(appName, ctx)
	if !ok {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString("a_name is not exist: " + appName)
		ctx.Application().Logger().Errorf("a_name is not exist: %s", appName)
		return
	}

	// validate whether the expect status is illegal
	if _, ok := application.ApplicationStatusMap[expectStatus]; !ok {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString("application status is illegal: " + status)
		ctx.Application().Logger().Errorf("application status is illegal: " + status)
		return
	}

	ctx.Application().Logger().Infof("UpdataApplicationStatus: the application with name <%s> expect status is <%s> "+
		"and realtime status is <%s>;", app.GetName(), app.GetStatus().Expect, app.GetStatus().Realtime)

	// update app resource status
	switch expectStatus {
	// TODO(ht): uninstall
	case application.NotInstalled: // uninstall a application

	case application.Running: // install or start a application
		// install a application
		if app.GetStatus().Expect == application.NotInstalled {
			app.GetApp().Status.Expect = application.Running
			app.GetApp().Status.Realtime = application.Installing
			// update app's real status
			go app.UpdateStatus(application.AInstall, ctx)

			// start a application
		} else if app.GetStatus().Expect == application.Stopped {
			app.GetApp().Status.Expect = application.Running
			app.GetApp().Status.Realtime = application.Starting
			// update app's real status
			go app.UpdateStatus(application.AStart, ctx)
		}
		// stop a application
	case application.Stopped:
		if app.GetStatus().Expect == application.Running {
			app.GetApp().Status.Expect = application.Stopped
			app.GetApp().Status.Realtime = application.Stopping
			// update app's real status
			go app.UpdateStatus(application.AStop, ctx)
		}
		// restart a application
	case application.Restart:
		if app.GetStatus().Expect == application.Running {
			app.GetApp().Status.Expect = application.Running
			app.GetApp().Status.Realtime = application.Restarting
			// update app's real status
			go app.UpdateStatus(application.ARestart, ctx)
		}
	}

	ctx.StatusCode(iris.StatusAccepted)
	_, _ = ctx.JSON(iris.Map{
		"name":   app.GetName(),
		"status": app.GetStatus(),
	})
	return
}

func getApplicationStatus(appType application.AppType, ctx iris.Context) {
	appName := ctx.Params().GetString("a_name")
	//app, ok := application.GetMemoryApplications().Get(appName)
	app, ok := application.GetETCDApplications(appType).Get(appName, ctx)
	if !ok {
		ctx.StatusCode(iris.StatusBadRequest)
		msg := fmt.Sprintf("GenericApplication with name <%s> is not exist: ", appName)
		ctx.WriteString(msg)
		ctx.Application().Logger().Error(msg)
		return
	}

	status := app.GetStatus()

	_, err := ctx.JSON(iris.Map{
		"name":   app.GetName(),
		"status": status,
	})
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.WriteString("got some error")
		ctx.Application().Logger().Errorf("get app status failed: %s" + appName)
		return
	}

	ctx.StatusCode(iris.StatusOK)
	return
}

func deleteApplication(appType application.AppType, ctx iris.Context) {
	appName := ctx.Params().GetString("a_name")
	ctx.Application().Logger().Infof("Prepare to delete a app named <%s>", appName)

	app, err := application.GetETCDApplications(appType).Delete(appName, ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.WriteString("got some error")
		ctx.Application().Logger().Errorf("Failed to delete a app <%s>.", appName)
		return
	}

	if app == nil {
		ctx.StatusCode(iris.StatusOK)
		ctx.WriteString("app not exist")
		return
	}

	ctx.StatusCode(iris.StatusOK)
	ctx.WriteString(app.GetName())
}

func setApplicationRealtimeStatus(appType application.AppType, ctx iris.Context) {
	appName := ctx.Params().GetString("a_name")

	var appHealthy utils.AppHealthy

	//reqBody, err := ctx.Request().GetBody()
	//if err != nil {
	//	ctx.StatusCode(iris.StatusBadRequest)
	//	ctx.WriteString(err.Error())
	//	ctx.Application().Logger().Errorf("Get body failed: %s", err)
	//	return
	//}

	// apply request body to app
	if err := ctx.ReadJSON(&appHealthy); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString(err.Error())
		ctx.Application().Logger().Errorf("CreateAppHealthy Error, json is illegal: %s", err)
		ctx.Application().Logger().Errorf("Request app info: %s", appName)

		//bodyBytes, err := ioutil.ReadAll(reqBody)
		//if err != nil {
		//	ctx.Application().Logger().Errorf("Get request body failed: %s", err)
		//	return
		//}
		//ctx.Application().Logger().Errorf("Get request body: %s", string(bodyBytes))
		return
	}

	var healthy = false
	if appHealthy.Code == "0" {
		healthy = true
	}

	app, ok := application.GetETCDApplications(appType).Get(appName, ctx)
	if !ok {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString("app not exist")
		return
	}

	expect := app.GetStatus().Expect

	if expect == application.Running && !healthy {
		app.SetStatus(application.ApplicationStatus(""), application.Failed, ctx)
	}
	ctx.StatusCode(iris.StatusAccepted)
}
