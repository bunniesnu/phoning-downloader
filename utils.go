package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/vbauerster/mpb/v8"
)

func appendEnv(key, value string) error {
	envFile := ".env"
	envMap, err := godotenv.Read(envFile)
	if err != nil {
		envMap = make(map[string]string)
	}
	envMap[key] = value
	return godotenv.Write(envMap, envFile)
}

func hash(url, apikey string) map[string]string {
	apiKey := []byte(apikey)

	msgpad := int(time.Now().UnixNano() / int64(time.Millisecond))

	if len(url) > 255 {
		url = url[:255]
	}

	message := []byte(url + strconv.Itoa(msgpad))

	mac := hmac.New(sha1.New, apiKey)
	mac.Write(message)
	digest := mac.Sum(nil)

	md := base64.StdEncoding.EncodeToString(digest)
	return map[string]string{
		"msgpad": strconv.Itoa(msgpad),
		"md": md,
	}
}

func hookTotalProgress(bar, totalBar *mpb.Bar) {
    go func() {
        var lastVal int64 = 0
        for {
            time.Sleep(100 * time.Millisecond) // adjust if needed
            curr := bar.Current()
            delta := curr - lastVal
            if delta > 0 {
                totalBar.IncrBy(int(delta))
                lastVal = curr
            }
            if bar.Completed() {
                break
            }
        }
    }()
}
