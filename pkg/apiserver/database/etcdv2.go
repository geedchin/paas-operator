package database

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/coreos/etcd/client"
)

var globalKapi client.KeysAPI
var etcdEndpoint = "http://127.0.0.1:2379"
var dbPrefix = "/k6s/database"

func init() {
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

var etcdDatabases = &ETCDDatabases{
	kapi:   globalKapi,
	prefix: dbPrefix,
}

func GetETCDDatabases() *ETCDDatabases {
	return etcdDatabases
}

func (dbs *ETCDDatabases) Add(name string, db Database) error {
	dbBytes, err := json.MarshalIndent(db, "", " ")
	if err != nil {
		return err
	}
	// eg. key==/k6s/database/mysql-xxx
	resp, err := dbs.kapi.Set(context.Background(), fmt.Sprintf("%s/%s", dbs.prefix, name), string(dbBytes), nil)
	if err != nil {
		return err
	}
	// TODO(ht) print some log
	_ = resp
	return nil
}

func (dbs *ETCDDatabases) Get(name string) (Database, bool) {
	resp, err := dbs.kapi.Get(context.Background(), fmt.Sprintf("%s/%s", dbs.prefix, name), nil)
	if err != nil {
		// TODO(ht) add log
		return &GenericDatabase{}, false
	}

	var retDb *GenericDatabase
	dbStr := resp.Node.Value
	dbBytes := []byte(dbStr)
	err = json.Unmarshal(dbBytes, retDb)
	if err != nil {
		return &GenericDatabase{}, false
	}
	return retDb, true
}
