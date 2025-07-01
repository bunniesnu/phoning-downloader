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
		decodedResponse := make(map[string]any)
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
	_, err = phoning(api_key, access_token, "/fan/v1.0/users/me")
	if err != nil {
		log.Fatalf("%v", err)
	}
	// All ready, safe to proceed
	calls, err := phoning(api_key, access_token, "/fan/v1.0/lives", map[string]string{"limit": "100"})
	if err != nil {
		log.Fatalf("%v", err)
	}
	for _, call := range calls["data"].([]any) {
		callMap := call.(map[string]any)
		liveId := int(callMap["liveId"].(float64))
		pnxml, err := getPNXML(api_key, access_token, liveId)
		if err != nil {
			log.Printf("Error getting PNXML for live ID %d: %v", liveId, err)
			break
		}
		url, ok := pnxml["url"].(string)
		if !ok {
			log.Fatalf("PNXML for live ID %d does not contain a valid URL", liveId)
		}
		log.Printf("Live ID %d URL: %s", liveId, url)
	}
}