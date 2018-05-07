package utils

import (
    "math/rand"
    "os"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
    b := make([]byte, n)
    for i := range b {
        b[i] = letterBytes[rand.Int63() % int64(len(letterBytes))]
    }
    return string(b)
}

func GetEnv(name, defaultValue string) string {
    if env := os.Getenv(name); env != "" {
        return env
    }
    return defaultValue
}