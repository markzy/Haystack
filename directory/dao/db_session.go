package dao

import (
	"gopkg.in/mgo.v2"
	"log"
)

type DataStore struct {
	session  *mgo.Session
	database *mgo.Database
}

var DS DataStore

func DBConnect(server string, database string) {
	session, err := mgo.Dial(server)
	if err != nil {
		log.Fatal(err)
	}
	DS.session = session
	DS.database = DS.session.DB(database)
}

func (ds *DataStore) SessionCopy() *mgo.Session {
	return ds.session.Copy()
}