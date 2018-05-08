package utils

import (
	"bytes"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}

func CalAvgTime(cycles int, cycleTimes []time.Duration) time.Duration {
	var avgTime time.Duration
	for _, ct := range cycleTimes {
		avgTime += ct
	}
	avgTime = avgTime / time.Duration(cycles)
	logrus.Debugf("Average SGD time for 20 cycles is : %s", avgTime)
	return avgTime
}

func SendResponse(payload []byte) {
	url := os.Getenv("LOGURL")
	logrus.Infof("Sending response to URL:>", url)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logrus.Errorf("Error in sending error log", err)
	}
	defer resp.Body.Close()

	logrus.Infof("response Status:", resp.Status)
}
