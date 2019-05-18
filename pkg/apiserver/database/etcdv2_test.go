package database

import "testing"

func TestETCDDatabases_Add(t *testing.T) {
	db := &GenericDatabase{}
	db.Name = "testdb"

	dbs := GetETCDDatabases()
	err := dbs.Add(db.GetName(), db)
	if err != nil {
		t.Error(err)
	}
}

func TestETCDDatabases_Get(t *testing.T) {

	dbs := GetETCDDatabases()
	db, ok := dbs.Get("testdb")
	if !ok {
		t.Error(ok)
	}

	t.Log(db.GetName())
}
