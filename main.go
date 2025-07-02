package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"strconv"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
)

func main() {
	outputDir := flag.String("o", "Downloads", "Directory to save downloaded videos")
	help := flag.Bool("h", false, "Show help message")
	flag.Parse()
	if *help {
		flag.Usage()
		os.Exit(0)
	}
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	access_token := os.Getenv("ACCESS_TOKEN")
	generatingAccount := false
	if access_token == "" {
		generatingAccount = true
		println("Access Token not found. Generating...")
		body, err := register()
		if err != nil {
			log.Fatal(err)
		}
		email := body["email"]
		password := body["password"]
		if email == "" || password == "" {
			log.Fatal("Email or password not found in registration response")
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
		print("Access token fetch: ")
		color.Green("success")
	}
	godotenv.Load()
	println("Checking configurations...")
	api_key := os.Getenv("API_KEY")
	print("API key: ")
	if api_key == "" {
		color.Red("not found")
	} else {
		color.Green("found")
	}
	print("Access token: ")
	access_token = os.Getenv("ACCESS_TOKEN")
	if access_token == "" {
		color.Red("not found")
	} else {
		color.Green("found")
	}
	if api_key == "" || access_token == "" {
		color.Red("Please check your configurations in the .env file.")
		os.Exit(1)
	}
	if generatingAccount {
		_, err := phoning("POST", api_key, "", "/fan/v1.0/login", map[string]string{
			"wevAccessToken": access_token,
			"tokenType": "APNS",
			"deviceToken": "",
		})
		if err != nil {
			color.Red("failed\nYou do not have access to the Phoning API. Please check your network connection and API key.")
			os.Exit(1)
		}
	}
	print("Checking access to Phoning API... ")
	_, err = phoning("GET", api_key, access_token, "/fan/v1.0/users/me")
	if err != nil {
		color.Red("failed\nYou do not have access to the Phoning API. Please check your network connection, API key, and access token.")
	} else {
		color.Green("success")
		println("You have access to the Phoning API.")
	}
	// All ready, safe to proceed
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Failed to create Downloads directory: %v", err)
	}
	var callsData []any = make([]any, 0)
	nextCursor := ""
	cnt := 0
	for {
		cnt++
		if cnt > 10 {
			log.Fatal("Too many iterations, stopping to prevent infinite loop")
		}
		params := map[string]string{"limit": "100"}
		if nextCursor != "" {
			params["cursor"] = nextCursor
		}
		calls, err := phoning("GET", api_key, access_token, "/fan/v1.0/lives", params)
		if err != nil {
			log.Fatalf("%v", err)
		}
		callsMap, ok := calls["data"].([]any)
		if !ok {
			log.Fatalf("Unexpected data format: %T", calls["data"])
		}
		callsData = append(callsData, callsMap...)
		cursors, ok := calls["cursors"].(map[string]any)
		if !ok {
			log.Fatalf("Unexpected cursors format: %T", calls["cursors"])
		}
		next, ok := cursors["next"].(string)
		if !ok {
			break
		}
		nextCursor = next
	}
	for _, call := range callsData {
		callMap := call.(map[string]any)
		liveId := int(callMap["liveId"].(float64))
		pnxml, err := getPNXML(api_key, access_token, liveId)
		if err != nil {
			log.Fatalf("Error getting PNXML for live ID %d: %v", liveId, err)
		}
		url, ok := pnxml["url"].(string)
		if !ok {
			log.Fatalf("PNXML for live ID %d does not contain a valid URL", liveId)
		}
		ctx := context.Background()
		DownloadVideo(ctx, url, *outputDir + "/" + strconv.Itoa(liveId) + ".mp4", *outputDir, 10)
	}
}