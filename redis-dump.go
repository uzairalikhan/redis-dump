package main

import (
	"bufio"
	"fmt"
	"os"
	"crypto/sha256"
	"net/http"
	"bytes"
	"io/ioutil"
	"encoding/json"
	"math/rand"
	"time"
	"strconv"
	
	"github.com/go-redis/redis"
	"github.com/Sirupsen/logrus"
	"github.com/uzairalikhan/redis-dump/utils"
)

type stats struct {
	ContainerId string
	ErrorMsg error
}

var client *redis.Client

func init() {
	// log level hierarchy anything that is info or above (debug, info, warn, error, fatal, panic). Default Info.
	switch loglevel := os.Getenv("LOGLEVEL"); loglevel {
		case "debug":
			logrus.SetLevel(logrus.DebugLevel)
		case "info":
			logrus.SetLevel(logrus.InfoLevel)
		case "warn":
			logrus.SetLevel(logrus.WarnLevel)
		case "error":
			logrus.SetLevel(logrus.ErrorLevel)
		case "fatal":
			logrus.SetLevel(logrus.FatalLevel)
		case "panic":
			logrus.SetLevel(logrus.PanicLevel)
		default:
			logrus.SetLevel(logrus.InfoLevel)
	}

	var host = os.Getenv("HOST")
	rand.Seed(time.Now().UnixNano())	
	
	client = redis.NewClient(&redis.Options{
		Addr:     host,
		Password: "", // no password set
		DB:       0,  // use default DB
		MaxRetries: 2,
		DialTimeout: 10 * time.Second, //default timeout is 5sec
		ReadTimeout: 5 * time.Second, //Default 3sec
		WriteTimeout: 5 * time.Second, //Default 3sec
	})	
	_, err := client.Ping().Result()
	if err != nil {
		logrus.Panicf("Error while connecting to redis: \n %s", err)
		panic(err)
	}
	logrus.Debugf("Redis connected to %s", "0.0.0.0:6379")
}

func main() {
	var freq = os.Getenv("FREQ")
	freqInt, err := strconv.Atoi(freq)
	if err != nil {
		logrus.Errorf("Invalid Frequency provided")
		panic(err)
	}
		// Read binary data
		// bytes, err := readBinary("dockerd")
		// if err !=nil {
		// 	logrus.Errorf("Error while reading binary data")
		// 	panic(err)
		// }
		// logrus.Debugf("Binary read success")
			
		for {
			//Run each sgd iteration after mentioned frequency
			time.Sleep(time.Duration(freqInt) * time.Second)
			randValue := utils.RandStringBytes(20)
			start := time.Now()
			sgd([]byte(randValue))
			elapsed := time.Since(start)
			logrus.Infof("SGD operation took %s", elapsed)
		}						
}

func readBinary(filename string) ([]byte, error){
	file, err := os.Open(filename)

    if err != nil {
        return nil, err
    }
    defer file.Close()

    stats, statsErr := file.Stat()
    if statsErr != nil {
        return nil, statsErr
    }

    var size int64 = stats.Size()
    bytes := make([]byte, size)

    bufr := bufio.NewReader(file)
    _,err = bufr.Read(bytes)

    return bytes, err
}

func sendResponse(err error) {
	//url := "http://172.16.23.248:4000/node/log"
	url := os.Getenv("LOGURL")
	nodeID := os.Getenv("NODEID")
	logrus.Debugf("Sending response to URL:>", url)

	data := stats{nodeID, err} 
	logJson, err := json.Marshal(data)
	if err != nil{
		logrus.Errorf("Error during marshal", err)
		//panic(err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(logJson))
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
		logrus.Errorf("Error in sending error log", err)	
        //panic(err)
    }
    defer resp.Body.Close()

    fmt.Println("response Status:", resp.Status)
    body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
}

func sgd(value []byte) {
		// Random 10 character string key for data to be stored in redis
		randString := utils.RandStringBytes(10)
		//logrus.Debugf("Storing value against key : %s", randString)
		err := client.Set(randString, value, 10*time.Minute).Err()
		if err != nil {
			logrus.Errorf("Error while saving data in redis for key: %s", randString)
			logrus.Errorf("Error is : %v", err)
			//sendResponse(err)
			//panic(err)
		} else {
			logrus.Debugf("Data saved for key: %s  is   %s", randString, value)

		// Check if data integrity is maintained or not		
		fetchedData, err := client.Get(randString).Result()
		if err != nil {
			logrus.Errorf("Error while reading data from redis for key: %s", randString)
			logrus.Errorf("Error is : %v", err)
		}
		logrus.Debugf("Data fetched for key: %s    is     %s", randString, fetchedData)
		if (sha256.Sum256(value) != sha256.Sum256([]byte(fetchedData))){
			logrus.Errorf("Data is not same for key: %s , expected value: %s  , recieved value  %s", randString, value, fetchedData)
		}
		
		err = client.Del(randString).Err()
		if err != nil {
			logrus.Errorf("Error while deleting binary data from redis for key: %s", randString)
			logrus.Errorf("Error is : %v", err)
			//sendResponse(err)
		}
		logrus.Debugf("Data deleted for key: %s", randString)
		}
}