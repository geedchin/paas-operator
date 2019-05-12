package database

import (
	"errors"
	"sync"
)

// MemoryDatabases holds all created database in memory
type MemoryDatabases struct {
	lock  sync.RWMutex
	dbMap map[string]Database
}

// global databases store
var memoryDatabases = &MemoryDatabases{
	dbMap: make(map[string]Database, 0),
}

// return global memory databases store
func GetMemoryDatabases() *MemoryDatabases {
	return memoryDatabases
}

// Add adds a database to global memory databases
func (dbs *MemoryDatabases) Add(name string, db Database) error {
	dbs.lock.Lock()
	defer dbs.lock.Unlock()
	if _, ok := dbs.dbMap[name]; ok {
		return errors.New("database is already exist: " + name)
	}
	dbs.dbMap[name] = db
	return nil
}

// Get gets a database from memory databases by name
func (dbs *MemoryDatabases) Get(name string) (Database, bool) {
	dbs.lock.RLock()
	defer dbs.lock.RUnlock()
	if db, ok := dbs.dbMap[name]; ok {
		return db, true
	}
	return &GenericDatabase{}, false
}