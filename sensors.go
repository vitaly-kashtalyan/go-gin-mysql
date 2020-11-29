package main

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"log"
	"net/http"
	"time"
)

type Sensors struct {
	Pin         int       `gorm:"default:NULL" sql:"type:tinyint(3);" binding:"required" json:"pin"`
	DecSensor   string    `gorm:"default:NULL" sql:"type:char(16);" binding:"min=16,max=16" json:"dec"`
	Temperature float32   `gorm:"default:NULL" sql:"type:decimal(4,2);" json:"temperature"`
	Humidity    float32   `gorm:"default:NULL" sql:"type:decimal(4,2);" json:"humidity"`
	CreatedAt   time.Time `json:"create_at"`
	UpdatedAt   time.Time `json:"update_at"`
}

type SensorsHistory struct {
	ID          uint      `gorm:"primary_key" json:"id"`
	Pin         int       `gorm:"default:NULL" sql:"type:tinyint(3);index:idx_pin" binding:"required" json:"pin"`
	DecSensor   string    `gorm:"default:NULL" sql:"type:char(16);" binding:"min=16,max=16" json:"dec"`
	Temperature float32   `gorm:"default:NULL" sql:"type:decimal(4,2);" json:"temperature"`
	Humidity    float32   `gorm:"default:NULL" sql:"type:decimal(4,2);" json:"humidity"`
	CreatedAt   time.Time `sql:"index" json:"date"`
}

type RelayStateHistory struct {
	ID        uint          `gorm:"primary_key" json:"id"`
	RelayId   sql.NullInt32 `gorm:"default:NULL" sql:"type:tinyint(3);" binding:"required" json:"relay_id"`
	State     sql.NullInt32 `gorm:"default:NULL" sql:"type:tinyint(3);" binding:"required" json:"state"`
	CreatedAt time.Time     `json:"create_at"`
}

type RelayStatus struct {
	Relay []Relay `json:"relays"`
}

type Relay struct {
	Id    int32 `json:"id"`
	State int32 `json:"state"`
}

func (Sensors) TableName() string {
	return "sensors"
}

func (SensorsHistory) TableName() string {
	return "sensors_history"
}

func (RelayStateHistory) TableName() string {
	return "relay_history"
}

type dht22 struct {
	Pin         int     `json:"pin"`
	Temperature float32 `json:"temperature"`
	Humidity    float32 `json:"humidity"`
	Status      string  `json:"status"`
}
type ds18b20 struct {
	Pin         int     `json:"pin"`
	Temperature float32 `json:"temperature"`
	Dec         string  `json:"dec"`
	Status      string  `json:"status"`
}

type Response struct {
	Dht22   []dht22   `json:"dht22"`
	Ds18b20 []ds18b20 `json:"ds18b20"`
}

type HttpOkData struct {
	Status  int         `json:"status" example:"200"`
	Message string      `json:"message" example:"OK"`
	Data    interface{} `json:"data" example:"interface{}"`
}

type HttpError struct {
	Status  int    `json:"status" example:"400"`
	Message string `json:"message" example:"Bad Request"`
}

func ResOkData(ctx *gin.Context, data interface{}) {
	ok := HttpOkData{
		Status:  http.StatusOK,
		Message: http.StatusText(http.StatusOK),
		Data:    data,
	}
	ctx.JSON(http.StatusOK, ok)
}

func ResStatus(ctx *gin.Context, code int) {
	er := HttpError{
		Status:  code,
		Message: http.StatusText(code),
	}
	ctx.JSON(http.StatusOK, er)
}

func (s *Sensors) AfterSave(scope *gorm.Scope) (err error) {
	sensorsHistory := SensorsHistory{}
	GetDB().Where(SensorsHistory{Pin: s.Pin, DecSensor: s.DecSensor}).
		Order("created_at desc").
		Limit(1).Find(&sensorsHistory)

	if sensorsHistory.ID == 0 || sensorsHistory.ID > 0 && sensorsHistory.Temperature != s.Temperature {
		var newRecord = SensorsHistory{
			Pin:         s.Pin,
			DecSensor:   s.DecSensor,
			Temperature: s.Temperature,
			Humidity:    s.Humidity}
		if err := GetDB().Create(&newRecord).Error; err != nil {
			log.Println("error creating sensor history record: ", err)
		}
	}
	return
}
