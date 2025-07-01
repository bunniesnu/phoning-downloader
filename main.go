package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	api_key := os.Getenv("API_KEY")
	if api_key == "" {
		log.Fatal("Please set API_KEY environment variable.")
	}
	access_token := os.Getenv("ACCESS_TOKEN")
	if access_token == "" {
		email := os.Getenv("EMAIL")
		password := os.Getenv("PASSWORD")
		if email == "" || password == "" {
			log.Fatal("Please set EMAIL and PASSWORDenvironment variables.")
		}
		respBody, err := getToken(email, password)
		if err != nil {
			log.Fatal(err)
		}
		decodedResponse := make(map[string]interface{})
		if err := json.Unmarshal(respBody, &decodedResponse); err != nil {
			log.Fatalf("Error decoding response: %v", err)
		}
		accessToken, ok := decodedResponse["accessToken"].(string)
		if !ok {
			log.Fatal("Access token not found in response")
		}
		appendEnv("ACCESS_TOKEN", accessToken)
	}
	godotenv.Load()
	access_token = os.Getenv("ACCESS_TOKEN")
	res, err := phoning(api_key, access_token, "/fan/v1.0/users/me")
	if err != nil {
		log.Fatalf("Error calling phoning API: %v", err)
	}
	log.Printf("Response: %v", res)
}