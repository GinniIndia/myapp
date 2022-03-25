package models

import (
    "gopkg.in/mgo.v2/bson"
    )

// Represents a structure for Covid Info
type Covid struct {
   ID         bson.ObjectId `bson:"_id" json:"id"`
   State      string `json:"state" bson:"state"`
   PatientCount float64    `json:"patient_count" bson:"patient_count"`
   Date     string `json:"date" bson:"date"`
}