package database

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

type ETCDDatabases struct {
	kapi   client.KeysAPI
	prefix string
}

var etcdDatabases = &ETCDDatabases{}

func GetETCDDatabases() *ETCDDatabases {
	etcdDatabases.kapi = globalKapi
	etcdDatabases.prefix = dbPrefix
	return etcdDatabases
}

func (dbs *ETCDDatabases) Add(name string, db Database, ctx iris.Context) error {
	dbBytes, err := json.MarshalIndent(db, "", " ")
	if err != nil {
		return err
	}
	// eg. key==/k6s/database/mysql-xxx
	key := fmt.Sprintf("%s/%s", dbs.prefix, name)
	_, err = dbs.kapi.Set(context.Background(), key, string(dbBytes), nil)
	if err != nil {
		ctx.Application().Logger().Errorf("Add db <%s> to etcd failed. with error: <%s>", name, err.Error())
		return err
	}
	ctx.Application().Logger().Infof("Add database to ETCDDatabases success: <%s>", name)
	return nil
}

func (dbs *ETCDDatabases) Get(name string, ctx iris.Context) (Database, bool) {
	key := fmt.Sprintf("%s/%s", dbs.prefix, name)
	resp, err := dbs.kapi.Get(context.Background(), key, nil)
	if err != nil {
		ctx.Application().Logger().Errorf("Get database from ETCDDatabases failed: <%s>", err.Error())
		return &GenericDatabase{}, false
	}

	var retDb = new(GenericDatabase)
	dbStr := resp.Node.Value
	dbBytes := []byte(dbStr)
	err = json.Unmarshal(dbBytes, retDb)
	if err != nil {
		ctx.Application().Logger().Errorf("Get db <%s>, json unmarshal failed: <%s>", name, err.Error())
		return &GenericDatabase{}, false
	}
	return retDb, true
}

func (dbs *ETCDDatabases) Delete(name string, ctx iris.Context) (Database, error) {
	key := fmt.Sprintf("%s/%s", dbs.prefix, name)
	resp, err := dbs.kapi.Delete(context.Background(), key, nil)
	if err != nil {
		ctx.Application().Logger().Errorf("Delete database from ETCDDatabases failed: <%s>", err.Error())
		return nil, err
	}

	ctx.Application().Logger().Infof("Delete database <%s> successful.", name)

	var retDb = new(GenericDatabase)
	dbStr := resp.Node.Value
	dbBytes := []byte(dbStr)
	err = json.Unmarshal(dbBytes, retDb)
	if err != nil {
		ctx.Application().Logger().Errorf("Delete db <%s>, json unmarshal failed: <%s>", name, err.Error())
		return nil, err
	}

	return retDb, nil
}
