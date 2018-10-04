package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/go-redis/redis"
	"github.com/uzairalikhan/redis-dump/utils"
)

// Stats ... Store dump information
type Stats struct {
	NodeID       string
	Errors       []error
	AvgCycleTime time.Duration
	Timestamp    time.Time
	Cycles       int
}

var client *redis.Client
var errors []error
var cycleTimes []time.Duration

func init() {
	// log level hierarchy anything that is info or above (debug, info, warn, error, fatal, panic). Default Info.
	level, err := logrus.ParseLevel(utils.GetEnv("LOGLEVEL", "info"))
	if err != nil {
		logrus.Errorf("Invalid Log Level")
		panic(err)
	}
	logrus.SetLevel(level)
	var host = utils.GetEnv("HOST", "0.0.0.0:6379")
	rand.Seed(time.Now().UnixNano())

	client = redis.NewClient(&redis.Options{
		Addr:         host,
		Password:     "", // no password set
		DB:           0,  // use default DB
		MaxRetries:   2,
		DialTimeout:  10 * time.Second, //default timeout is 5sec
		ReadTimeout:  5 * time.Second,  //Default 3sec
		WriteTimeout: 5 * time.Second,  //Default 3sec
	})
	_, err = client.Ping().Result()
	if err != nil {
		logrus.Errorf("Error while connecting to redis: \n %s", err)
		panic(err)
	}
	logrus.Debugf("Redis connected to %s", host)
}

func main() {
	defaultNode, _ := os.Hostname()
	freq, err := strconv.Atoi(utils.GetEnv("FREQ", "1"))
	if err != nil {
		logrus.Error("Invalid Frequency provided")
		panic(err)
	}
	cycles, err := strconv.Atoi(utils.GetEnv("CYCLES", "20"))
	if err != nil {
		logrus.Error("Invalid Cycles provided")
		panic(err)
	}
	for {
		//Run each sgd iteration after mentioned frequency
		time.Sleep(time.Duration(freq) * time.Second)
		randValue := utils.RandStringBytes(20)
		start := time.Now()
		sgd([]byte(randValue))
		elapsed := time.Since(start)
		logrus.Infof("SGD operation took %s", elapsed)

		cycleTimes = append(cycleTimes, elapsed)
		if len(cycleTimes) == cycles {
			avgTime := utils.CalAvgTime(cycles, cycleTimes)
			payload, err := json.Marshal(Stats{
				NodeID:       utils.GetEnv("NODEID", defaultNode),
				Errors:       errors,
				Timestamp:    time.Now(),
				AvgCycleTime: avgTime,
				Cycles:       cycles,
			})
			if err != nil {
				logrus.Errorf("Error while parsing json, Err is %s", err)
			}
			logrus.Debugf("Payload is %s", payload)
			go utils.SendResponse(payload)
			cycleTimes = cycleTimes[:0]
			errors = errors[:0]
		}
	}
}

// Keeping this for future use
func readBinary(filename string) ([]byte, error) {
	file, err := os.Open(filename)

	if err != nil {
		return nil, err
	}
	defer file.Close()

	stats, statsErr := file.Stat()
	if statsErr != nil {
		return nil, statsErr
	}

	var size = stats.Size()
	bytes := make([]byte, size)

	bufr := bufio.NewReader(file)
	_, err = bufr.Read(bytes)

	return bytes, err
}

//This function provides save, get and delete functionality for redis
func sgd(value []byte) {
	// Random 10 character string key for data to be stored in redis
	randString := utils.RandStringBytes(10)
	err := client.Set(randString, value, 10*time.Minute).Err()
	if err != nil {
		logrus.Errorf("Error while saving data in redis for key: %s", randString)
		logrus.Errorf("Error is : %v", err)
		errors = append(errors, err)
	} else {
		logrus.Debugf("Data saved for key: %s  is   %s", randString, value)

		// Check if data integrity is maintained or not
		fetchedData, err := client.Get(randString).Result()
		if err != nil {
			logrus.Errorf("Error while reading data from redis for key: %s", randString)
			logrus.Errorf("Error is : %v", err)
			errors = append(errors, err)
		}
		logrus.Debugf("Data fetched for key: %s    is     %s", randString, fetchedData)
		if sha256.Sum256(value) != sha256.Sum256([]byte(fetchedData)) {
			logrus.Errorf("Data is not same for key: %s , expected value: %s  , recieved value  %s", randString, value, fetchedData)
		}

		err = client.Del(randString).Err()
		if err != nil {
			logrus.Errorf("Error while deleting binary data from redis for key: %s", randString)
			logrus.Errorf("Error is : %v", err)
			errors = append(errors, err)
		}
		logrus.Debugf("Data deleted for key: %s", randString)
	}
}
