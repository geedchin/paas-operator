package apiserver

import (
	"fmt"
	"github.com/farmer-hutao/k6s/pkg/apiserver/database"
	"github.com/kataras/iris"
)

func Run() {
	app := iris.Default()
	applyRoute(app)
	err := app.Run(iris.Addr(fmt.Sprintf("%s:%s", "", "3334")))
	if err != nil {
		panic(err)
	}
}

func applyRoute(app *iris.Application) {
	db := &database.Database{}

	versionRouter := app.Party("/apis/v1alpha1")
	dbRouter := versionRouter.Party("/database")

	dbRouter.Get("{d_name}/status", db.Status)
	dbRouter.Post("/create", db.Create)
	dbRouter.Put("/{d_name:string}/{action:string}", db.UpdateStatus)
}
