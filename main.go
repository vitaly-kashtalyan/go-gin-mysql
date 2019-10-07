package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"log"
	"net/http"
	"os"
	"time"
)

func init() {
	GetDB().Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&Sensors{}, &SensorsHistory{})
}

func main() {
	dataReadingService()

	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		var data []Sensors

		err := GetDB().Find(&data).Error
		if err == gorm.ErrRecordNotFound {
			ResStatus(c, http.StatusNotFound)
			return
		}

		if err != nil {
			ResStatus(c, http.StatusBadRequest)
			log.Println(err)
			return
		}

		ResOkData(c, data)
	})

	_ = r.Run(":" + os.Getenv("PORT_ENV"))
}

// Call this function in a goroutine.
func dataReadingService() {
	nextTime := time.Now().Truncate(time.Minute)
	nextTime = nextTime.Add(time.Minute)
	time.Sleep(time.Until(nextTime))
	scanSensors()
	go dataReadingService()
}

func scanSensors() {
	response := Response{}

	if err := getJSON(&response); err == nil {
		tx := GetDB().Begin()
		for _, v := range response.Dht22 {
			if v.Status == http.StatusText(http.StatusOK) {
				if err := tx.Where(Sensors{Pin: v.Pin}).
					Assign(Sensors{Pin: v.Pin, Temperature: v.Temperature, Humidity: v.Humidity, UpdatedAt: time.Now()}).
					FirstOrCreate(&Sensors{}).Error; err != nil {
					tx.Rollback()
					log.Println(err)
				}
			}
		}
		for _, v := range response.Ds18b20 {
			if v.Status == http.StatusText(http.StatusOK) {
				if err := tx.Where(Sensors{Pin: v.Pin, DecSensor: v.Dec}).
					Assign(Sensors{Pin: v.Pin, Temperature: v.Temperature, DecSensor: v.Dec, UpdatedAt: time.Now()}).
					FirstOrCreate(&Sensors{}).Error; err != nil {
					tx.Rollback()
					log.Println(err)
				}
			}
		}
		tx.Commit()
	} else {
		log.Println("error getting json object: ", err)
	}
}

func getJSON(result interface{}) error {
	resp, err := http.Get("http://" + os.Getenv("HOST_SENSORS"))
	if err != nil {
		return fmt.Errorf("cannot fetch URL %q: %v", os.Getenv("HOST_SENSORS"), err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected http GET status: %s", resp.Status)
	}
	// We could check the resulting content type
	// here if desired.
	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		return fmt.Errorf("cannot decode JSON: %v", err)
	}
	return nil
}
