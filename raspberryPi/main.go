package main

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	sensors, err := Sensors()
	if err != nil {
		panic(err)
	}

	now := time.Now()
	//fmt.Printf("sensor IDs: %v\n", sensors)

	for _, sensor := range sensors {
		t, err := Temperature(sensor)
		if err == nil {

			fmt.Printf("date: %s | sensor: %s | temperature: %.2f°C\n", now.String(), sensor, t)

			db, err := sql.Open("mysql",
				"ba5390dfde843d:@tcp(eu-cdbr-azure-west-b.cloudapp.net:3306)/baptistedixneufhome")
			if err != nil {
				fmt.Printf(err.Error())
			}
			defer db.Close()

			stmt, err := db.Prepare("INSERT INTO measure(SensorId, DateTime, Value) VALUES (184, NOW() , ?);")
			if err != nil {
				fmt.Printf(err.Error())
			}
			res, err := stmt.Exec(t)
			if err != nil {
				fmt.Printf(err.Error())
			}
			lastID, err := res.LastInsertId()
			if err != nil {
				log.Fatal(err)
			}
			rowCnt, err := res.RowsAffected()
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("ID = %d, affected = %d\n", lastID, rowCnt)
		}
	}

}

// Sensors get all connected sensor IDs as array
func Sensors() ([]string, error) {
	data, err := ioutil.ReadFile("/sys/bus/w1/devices/w1_bus_master1/w1_master_slaves")
	if err != nil {
		return nil, err
	}

	sensors := strings.Split(string(data), "\n")
	if len(sensors) > 0 {
		sensors = sensors[:len(sensors)-1]
	}

	return sensors, nil
}

// Temperature get the temperature of a given sensor
func Temperature(sensor string) (float64, error) {
	data, err := ioutil.ReadFile("/sys/bus/w1/devices/" + sensor + "/w1_slave")
	if err != nil {
		return 0.0, nil
	}

	if strings.Contains(string(data), "YES") {
		arr := strings.SplitN(string(data), " ", 3)

		switch arr[1][0] {
		case 'f': //-0.5 ~ -55°C
			x, err := strconv.ParseInt(arr[1]+arr[0], 16, 32)
			if err != nil {
				return 0.0, err
			}
			return float64(^x+1) * 0.0625, nil

		case '0': //0~125°C
			x, err := strconv.ParseInt(arr[1]+arr[0], 16, 32)
			if err != nil {
				return 0.0, err
			}
			return float64(x) * 0.0625, nil
		}
	}

	return 0.0, errors.New("can not read temperature for sensor " + sensor)
}
