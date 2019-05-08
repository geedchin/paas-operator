package apiserver

import (
	"fmt"
	"github.com/kataras/iris"
)

func Run() {
	app := iris.Default()

	// for test only
	app.Get("/ping", func(ctx iris.Context) {
		_, _ = ctx.JSON(iris.Map{
			"msg": "pong",
		})
	})

	applyRoute(app)
	err := app.Run(iris.Addr(fmt.Sprintf("%s:%s", "", "3334")))
	if err != nil {
		panic(err)
	}
}

func applyRoute(app *iris.Application) {
	versionRouter := app.Party("/apis/v1alpha1")
	dbRouter := versionRouter.Party("/database")

	// 查询Database状态
	dbRouter.Get("/{d_name}/status", GetDatabaseStatus)
	// 创建Database资源
	dbRouter.Post("/create", CreateDatabase)
	// 修改Database的期望状态，status-> running/stoped/not-installed
	dbRouter.Put("/{d_name:string}/{status:string}", UpdateDatabaseStatus)
}
