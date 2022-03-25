package dao

import (
	"log"
    "gopkg.in/mgo.v2/bson"
	. "myapp/models"
	mgo "gopkg.in/mgo.v2"
)

type CovidDAO struct {
	Server   string
	Database string
}

var db *mgo.Database

const (
	COLLECTION = "covid_case7"
)

// Establish a connection to database
func (m *CovidDAO) Connect() {
	session, err := mgo.Dial(m.Server)
	if err != nil {
		log.Fatal(err)
	}
	db = session.DB(m.Database)
}

// Insert a State Covid Information
func (m *CovidDAO) Insert(covid Covid) error {
	err := db.C(COLLECTION).Insert(&covid)
	return err
}

// Find the infomation for specific state
func (m *CovidDAO) FindAll(region_id string) ([]Covid, error) {
	var covid []Covid
	err := db.C(COLLECTION).Find(bson.M{"state":region_id}).All(&covid)
	return covid, err
}

// Update an existing covid state data
func (m *CovidDAO) Update(covid Covid) error {
	err := db.C(COLLECTION).Update(bson.M{"state":covid.State}, bson.M{"$set": bson.M{"patient_count":covid.PatientCount,"date_time":covid.Date}})
	return err
}