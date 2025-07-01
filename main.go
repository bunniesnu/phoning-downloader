package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"golang.org/x/term"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	access_token := os.Getenv("ACCESS_TOKEN")
	if access_token == "" {
		println("Access Token not found.")
		email := ""
		password := ""
		for email == "" {
			print("Enter your email: ")
			fmt.Scanln(&email)
			if email == "" {
				println("Email cannot be empty. Please try again.")
			}
		}
		for password == "" {
			print("Enter your password: ")
			bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				log.Fatal("Error reading password:", err)
			}
			password = string(bytePassword)
			if password == "" {
				println("Password cannot be empty. Please try again.")
			}
		}
		println("\nFetching access token...")
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
		println("Access token fetched successfully.")
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
	print("Checking access to Phoning API... ")
	_, err = phoning(api_key, access_token, "/fan/v1.0/users/me")
	if err != nil {
		color.Red("failed\nYou do not have access to the Phoning API. Please check your network connection, API key, and access token.")
	} else {
		color.Green("success")
		println("You have access to the Phoning API.")
	}
	// All ready, safe to proceed
	downloadDir := "./Downloads"
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
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
		calls, err := phoning(api_key, access_token, "/fan/v1.0/lives", params)
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
		DownloadVideo(ctx, url, downloadDir + "/" + strconv.Itoa(liveId) + ".mp4", "./Downloads", 10)
	}
}