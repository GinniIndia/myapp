package main

import (
    "fmt"
    "os"
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
var ACCESS_KEY = "" // Create a ACCESS_KEY for Reverse Gecoding API and define in this string.
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
    log.Print("Covid Data API : " + apiUrl)
    resp, err := http.Get(apiUrl)
    CheckError(err)
    body, err1 := ioutil.ReadAll(resp.Body)
    CheckError(err1)
    var data map[string]interface{}
    err2 := json.Unmarshal([]byte(string(body)), &data)
    CheckError(err2)
    for key, val := range data {
         result := val.(map[string]interface{})["total"].(map[string]interface{})
         confirmed := result["confirmed"].(float64)
         deceased := result["deceased"].(float64)
         recovered := result["recovered"].(float64)
         currentTime := time.Now()
         covidData := Covid{ID: bson.NewObjectId(), State: key, PatientCount: confirmed - deceased - recovered, Date: currentTime.String()}

         // Checking whether data already exists in db, if not present then storing it.
         resp1, err3 := dao.FindAll(key)
         CheckError(err3)
         if len(resp1) == 0 {
            err4 := dao.Insert(covidData)
            CheckError(err4)
         } else {
            err5 := dao.Update(covidData)
            CheckError(err5)
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
     log.Printf("Reverse GeoCoding API Url : " + apiUrl)
     resp, err := http.Get(apiUrl)
     CheckError(err)
     body, err1 := ioutil.ReadAll(resp.Body)
     CheckError(err1)
     var data map[string]interface{}
     err2 := json.Unmarshal([]byte(string(body)), &data)
     CheckError(err2)
     log.Print("Data from Reverse GeoCoding API:")
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

     log.Print("Region Code : " + regionCode)
     log.Print("Patient Count : " + patientCount)
     log.Print("Date Time : " + dateTime)
     finalResult := map[string]interface{}{"state":regionCode, "patient_count":patientCount, "date_time":dateTime}
     return c.JSON(http.StatusOK, finalResult)
}

// Health Check of App
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

    port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}
	e.Logger.Fatal(e.Start(":"+port))
}