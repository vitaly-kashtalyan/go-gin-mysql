package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	SensorsHost = "SENSORS_HOST"
	RelaysHost  = "RELAYS_HOST"
)

func init() {
	GetDB().Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&Sensors{}, &SensorsHistory{}, &RelayStateHistory{})
}

func main() {
	dataReadingService()

	r := gin.Default()
	r.Use(cors.Default())
	r.GET("/", func(c *gin.Context) {
		var data []Sensors

		//err := GetDB().Find(&data).Error
		err := GetDB().Raw("select id,pin,dec_sensor,round(avg(temperature), 2) as temperature,humidity,round(avg(humidity), 2) as humidity, created_at from sensors_history where created_at > (select updated_at from sensors order by updated_at desc limit 1) - INTERVAL 10 MINUTE group by pin, dec_sensor;").Scan(&data).Error
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

	_ = r.Run()
}

// Call this function in a goroutine.
func dataReadingService() {
	nextTime := time.Now().Truncate(time.Minute)
	nextTime = nextTime.Add(time.Minute)
	time.Sleep(time.Until(nextTime))
	scanSensors()
	scanRelays()
	go dataReadingService()
}

func scanRelays() {
	relayStatus := RelayStatus{}
	if err := getJSON("http://"+getRelaysHost()+"/status", &relayStatus); err == nil {
		for _, relay := range relayStatus.Relay {
			relayStateHistory := RelayStateHistory{}
			GetDB().Where(RelayStateHistory{RelayId: sql.NullInt32{Int32: relay.Id, Valid: true}}).
				Order("created_at desc").
				Limit(1).Find(&relayStateHistory)

			var newRecord = RelayStateHistory{
				RelayId:   sql.NullInt32{Int32: relay.Id, Valid: true},
				State:     sql.NullInt32{Int32: relay.State, Valid: true},
				CreatedAt: time.Now()}
			if err := GetDB().Create(&newRecord).Error; err != nil {
				log.Println("error creating relay history record: ", err)
			}
		}
	} else {
		log.Println("error getting json object: ", err)
	}
}

func scanSensors() {
	response := Response{}
	if err := getJSON("http://"+getSensorsHost(), &response); err == nil {
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

func getJSON(url string, result interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("cannot fetch URL %q: %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected http GET status: %s", resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		return fmt.Errorf("cannot decode JSON: %v", err)
	}
	return nil
}

func getRelaysHost() string {
	return os.Getenv(RelaysHost)
}

func getSensorsHost() string {
	return os.Getenv(SensorsHost)
}
