package apiserver

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/kataras/iris"
)

func Run() {
	f := newLogFile("./log")
	defer f.Close()

	app := iris.New()
	app.Logger().AddOutput(f)

	// for test only
	app.Get("/ping", func(ctx iris.Context) {
		_, _ = ctx.JSON(iris.Map{
			"msg": "pong",
		})
	})

	applyRoute(app)
	if err := app.Run(iris.Addr(fmt.Sprintf("%s:%s", "", "3334"))); err != nil {
		app.Logger().Fatal(err)
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

// eg. path=/var/log ->
func newLogFile(path string) *os.File {
	todayFilename := func() string {
		today := time.Now().Format("2006-0102-1504-05")
		return today + ".log"
	}

	filename := todayFilename()
	log.Println("logfile: " + filename)

	//create log dir
	if err := os.MkdirAll(path, 0666); err != nil {
		panic(err)
	}

	// Open the file, this will append to the today's file if server restarted.
	f, err := os.OpenFile(filepath.Join(path, filename), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}

	return f
}
