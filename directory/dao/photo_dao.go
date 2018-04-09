package dao

import (
	. "Haystack/directory/models"
	"gopkg.in/mgo.v2/bson"
)

type PhotoMetaDAO struct{}

const (
	photoCollection = "photoMeta"
)

func (m *PhotoMetaDAO) Insert(pm PhotoMeta) error {
	err := DS.database.C(photoCollection).Insert(&pm)
	return err
}

func (m *PhotoMetaDAO) FindById(id string) (PhotoMeta, error) {
	var photo PhotoMeta
	err := DS.database.C(photoCollection).Find(bson.M{"photoid": id}).One(&photo)
	return photo, err
}


func (m *PhotoMetaDAO) Update(photo PhotoMeta) error {
	err := DS.database.C(photoCollection).Update(bson.M{"photoid": photo.PhotoID}, bson.M{"$set": bson.M{"state": photo.State}})
	return err
}