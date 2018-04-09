package dao

import (
	. "Haystack/directory/models"
	"gopkg.in/mgo.v2/bson"
)

type LogicalMappingDAO struct {}

const (
	mappingCollection = "logicalMapping"
)

func (l *LogicalMappingDAO) FindAll() ([]LogicalMapping, error) {
	var mappings []LogicalMapping
	err := DS.database.C(mappingCollection).Find(bson.M{}).All(&mappings)
	return mappings, err
}

func (l *LogicalMappingDAO) Insert(mapping LogicalMapping)  error{
	err := DS.database.C(mappingCollection).Insert(&mapping)
	return err
}
