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
	mwRouter := versionRouter.Party("/middleware")

	// Query a database status
	dbRouter.Get("/{a_name}/status", GetDatabaseStatus)
	// Create a database
	dbRouter.Post("/create", CreateDatabase)
	// Update a database's expect status, status -> [ running、stopped、not-installed ]
	dbRouter.Put("/{a_name:string}/{status:string}", UpdateDatabaseStatus)
	// Delete a database by name
	dbRouter.Delete("/{a_name}", DeleteDatabase)

	// Query a middleware status
	mwRouter.Get("/{a_name}/status", GetMiddlewareStatus)
	// Create a middleware
	mwRouter.Post("/create", CreateMiddleware)
	// Update a middleware's expect status, status -> [ running、stopped、not-installed ]
	mwRouter.Put("/{a_name:string}/{status:string}", UpdateMiddlewareStatus)
	// Delete a middleware by name
	mwRouter.Delete("/{a_name}", DeleteMiddleware)
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
