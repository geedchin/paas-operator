package application

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kataras/iris"
	"log"
	"os"
	"time"

	"github.com/coreos/etcd/client"
)

var globalKapi client.KeysAPI
var etcdEndpoint = os.Getenv("ETCD_ENDPOINT")
var dbPrefix = os.Getenv("ETCD_DB_PREFIX")
var mwPrefix = os.Getenv("ETCD_MW_PREFIX")

func init() {
	if etcdEndpoint == "" {
		log.Printf("Warning: %s is unset, use default value: %s", "ETCD_ENDPOINT", etcdEndpoint)
		etcdEndpoint = "http://127.0.0.1:2379"
	} else {
		log.Printf("ETCD ENDPOINT: %s", etcdEndpoint)
	}
	if dbPrefix == "" {
		log.Printf("Warning: %s is unset, use default value: %s", "ETCD_DB_PREFIX", dbPrefix)
		dbPrefix = "/k6s/database"
	}
	if mwPrefix == "" {
		log.Printf("Warning: %s is unset, use default value: %s", "ETCD_MW_PREFIX", mwPrefix)
		mwPrefix = "/k6s/middleware"
	}
	cfg := client.Config{
		Endpoints: []string{etcdEndpoint},
		Transport: client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	globalKapi = client.NewKeysAPI(c)
}

type ETCDApplications struct {
	kapi   client.KeysAPI
	prefix string
}

var etcdDatabases = &ETCDApplications{}

var etcdMiddlewares = &ETCDApplications{}

func GetETCDApplications(appType AppType) *ETCDApplications {
	switch appType {
	case APP_DATABASE:
		etcdDatabases.kapi = globalKapi
		etcdDatabases.prefix = dbPrefix
		return etcdDatabases
	case APP_MIDDLEWARE:
		etcdMiddlewares.kapi = globalKapi
		etcdMiddlewares.prefix = mwPrefix
		return etcdMiddlewares
	default:
		log.Fatal("AppType illegal")
		return nil
	}
}

func (apps *ETCDApplications) Add(name string, app Application, ctx iris.Context) error {
	appBytes, err := json.MarshalIndent(app, "", " ")
	if err != nil {
		return err
	}
	// eg. key==/k6s/database/mysql-xxx
	key := fmt.Sprintf("%s/%s", apps.prefix, name)
	_, err = apps.kapi.Set(context.Background(), key, string(appBytes), nil)
	if err != nil {
		ctx.Application().Logger().Errorf("Add app <%s> to etcd failed. with error: <%s>", name, err.Error())
		return err
	}
	ctx.Application().Logger().Infof("Add application to ETCDApplications success: <%s>", name)
	return nil
}

func (apps *ETCDApplications) Update(name string, app Application, ctx iris.Context) error {
	appBytes, err := json.MarshalIndent(app, "", " ")
	if err != nil {
		return err
	}
	// eg. key==/k6s/database/mysql-xxx
	key := fmt.Sprintf("%s/%s", apps.prefix, name)
	_, err = apps.kapi.Update(context.Background(), key, string(appBytes))
	if err != nil {
		ctx.Application().Logger().Errorf("Update app <%s> to etcd failed. with error: <%s>", name, err.Error())
		return err
	}
	ctx.Application().Logger().Infof("Update application to ETCDApplications success: <%s>", name)
	return nil
}

func (apps *ETCDApplications) Get(name string, ctx iris.Context) (Application, bool) {
	key := fmt.Sprintf("%s/%s", apps.prefix, name)
	resp, err := apps.kapi.Get(context.Background(), key, nil)
	if err != nil {
		ctx.Application().Logger().Errorf("Get application from ETCDApplications failed: <%s>", err.Error())
		return &GenericApplication{}, false
	}

	var retApp = new(GenericApplication)
	appStr := resp.Node.Value
	appBytes := []byte(appStr)
	err = json.Unmarshal(appBytes, retApp)
	if err != nil {
		ctx.Application().Logger().Errorf("Get app <%s>, json unmarshal failed: <%s>", name, err.Error())
		return &GenericApplication{}, false
	}
	return retApp, true
}

func (apps *ETCDApplications) Delete(name string, ctx iris.Context) (Application, error) {
	key := fmt.Sprintf("%s/%s", apps.prefix, name)
	resp, err := apps.kapi.Get(context.Background(), key, nil)
	// not exist, return nil, nil
	if err != nil {
		ctx.Application().Logger().Errorf("Delete application from ETCDApplications failed when query app: <%s>", err.Error())
		return nil, nil
	}

	var retApp = new(GenericApplication)
	appStr := resp.Node.Value
	appBytes := []byte(appStr)
	err = json.Unmarshal(appBytes, retApp)
	if err != nil {
		ctx.Application().Logger().Errorf("Delete app <%s>, json unmarshal failed: <%s>", name, err.Error())
		return nil, err
	}

	resp, err = apps.kapi.Delete(context.Background(), key, nil)
	if err != nil {
		ctx.Application().Logger().Errorf("Delete application from ETCDApplications failed: <%s>", err.Error())
		return nil, err
	}

	ctx.Application().Logger().Infof("Delete application <%s> successful.", name)

	return retApp, nil
}
