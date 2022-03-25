package main

import (
    "fmt"
	"encoding/json"
	"log"
	"net/http"
    "io/ioutil"
	"gopkg.in/mgo.v2/bson"
	. "myapp/config"
	. "myapp/dao"
	. "myapp/models"
	"github.com/labstack/echo/v4"
	"time"
	_ "myapp/docs"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// Config variable
var config = Config{}
var dao = CovidDAO{}

// Global Variables
var ACCESS_KEY = "bb0fe31b2d1f4f3434f1a7c3d0e024c5"
var LATITUDE string
var LONGITUDE string

// Parse the configuration file 'config.toml', and establish a connection to DB
func init() {
	config.Read()
	dao.Server = config.Server
	dao.Database = config.Database
	dao.Connect()
}

// Log Error Information
func CheckError(err error) {
     if err != nil {
      log.Fatalln(err)
     }
}

// Fetch the Covid Data from api and Store into Mongo Db
func fetchAndStoreCovidData(c echo.Context) error {
    apiUrl := "https://data.covid19india.org/v4/min/data.min.json"
    log.Print(apiUrl)
    resp, err := http.Get(apiUrl)
    CheckError(err)
    body, err1 := ioutil.ReadAll(resp.Body)
    CheckError(err1)
    var data map[string]interface{}
    err2 := json.Unmarshal([]byte(string(body)), &data)
    CheckError(err2)
    for key, val := range data {
         // Checking whether data already exists in db, if not present then storing it.
         resp1, err3 := dao.FindAll(key)
         CheckError(err3)
         if len(resp1) == 0 {
             state:= val.(map[string]interface{})
             result := state["total"].(map[string]interface{})
             confirmed := result["confirmed"].(float64)
             deceased := result["deceased"].(float64)
             recovered := result["recovered"].(float64)
             log.Print(key)
             log.Print(confirmed - deceased - recovered)
             currentTime := time.Now()
             covidData := Covid{ID: bson.NewObjectId(), State: key, PatientCount: confirmed - deceased - recovered, Date: currentTime.String()}
             err4 := dao.Insert(covidData)
             CheckError(err4)
         }
    }
    return c.JSON(http.StatusOK, map[string]interface{}{"msg":"Covid Info Stored SuccessFully!!"})
}

// Get Covid Patient Count from Co-ordinates of Region
func getPatientsCount(c echo.Context) error {
     // Retriving Inputs from Get Request - Lattitude and Longitude values
     err_ := echo.QueryParamsBinder(c).
     String("latitude", &LATITUDE).
     String("longitude", &LONGITUDE).
     BindError()
     CheckError(err_)

     // Retriving Reverse GeoCoding data from API using respective Lattitude and Longitude values
     apiUrl := "http://api.positionstack.com/v1/reverse?access_key=" + ACCESS_KEY + "&query=" + LATITUDE + "," + LONGITUDE + "&limit=1"
     log.Printf(apiUrl)
     resp, err := http.Get(apiUrl)
     CheckError(err)
     body, err1 := ioutil.ReadAll(resp.Body)
     CheckError(err1)
     log.Print(body)
     var data map[string]interface{}
     err2 := json.Unmarshal([]byte(string(body)), &data)
     CheckError(err2)
     log.Print(data)
     regionInfo := data["data"]

     // Validation of Inputs
     if regionInfo == nil {
         return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Invalid Input Parameters"})
     }

     regionData := regionInfo.([]interface{})
     var regionCode string
     var patientCount string
     var dateTime string

     // Validation of Inputs and Region info from DB
     if len(regionData) > 0{
        var tmpRegionCode =  regionData[0].(map[string]interface{})["region_code"]
        log.Print(tmpRegionCode)
        if tmpRegionCode != nil {
            // Retriving RegionCode info from Reverse Geocoding API
            regionCode = tmpRegionCode.(string)
        } else {
            return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Invalid Input Parameters"})
        }
        // Finding PatientCount and Timestamp using Region Id.
        resp1, err3 := dao.FindAll(regionCode)
        CheckError(err3)
        if len(resp1) > 0 {
            patientCount = fmt.Sprintf("%f", resp1[0].PatientCount)
            dateTime = resp1[0].Date
        } else {
            return c.JSON(http.StatusNotFound , map[string]interface{}{"msg": "Region Information not Found in DB"})
        }
     }

     log.Print(regionCode)
     finalResult := map[string]interface{}{"state":regionCode, "patient_count":patientCount, "date_time":dateTime}
     return c.JSON(http.StatusOK, finalResult)
}

func healthCheck(c echo.Context) error {
   return c.JSON(http.StatusOK, map[string]interface{}{
      "data": "Server is up and running",
   })
}

// Start our Go Echo Application
func main() {
	e := echo.New()
    e.GET("/v1/fetch_and_store_covid_data", fetchAndStoreCovidData)
    e.GET("/v1/get_covid_patients_count_for_region", getPatientsCount)
    e.GET("/v1/health", healthCheck)
    e.GET("/v1/swagger/*", echoSwagger.WrapHandler)
	e.Logger.Fatal(e.Start(":8185"))
}