package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

func main() {
	outputDir := flag.String("o", "Downloads_tmp", "Directory to save downloaded videos")
	concurrency := flag.Int("c", 10, "Concurrency")
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
	println("Fetching calls...")
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
	println("Found", len(callsData), "calls. Fetching informations...")
	p := mpb.New(mpb.WithWidth(64), mpb.PopCompletedMode())
	num := len(callsData)
	sizes := make(map[int]int64, num)
	bar := p.New(int64(num),
		mpb.BarStyle().Lbound("[").Filler("=").Tip(">").Padding(" ").Rbound("]"),
		mpb.PrependDecorators(
			decor.Name("Fetching...", decor.WC{W: 5, C: decor.DindentRight}),
		),
		mpb.AppendDecorators(
			decor.NewPercentage("%.2f", decor.WC{W: 7}),
		),
	)
	fetchFunction := func (call any, ctx context.Context, cancel context.CancelFunc, errCh chan error) {
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
		headReq, _ := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
		resp, err := http.DefaultClient.Do(headReq)
		if err != nil || resp.StatusCode != http.StatusOK {
			log.Printf("HEAD failed for live ID %d: %v", liveId, err)
			cancel()
			return
		}
		length := resp.ContentLength
		resp.Body.Close()
		sizes[liveId] = length
		bar.IncrInt64(1)
	}
	ok, err := concurrentExecute(fetchFunction, callsData, *concurrency)
	if err != nil {
		log.Fatalf("Error during concurrent execution: %v", err)
	}
	if !ok {
		log.Fatal("Some fetches failed, check the error log for details.")
	}
	p.Wait()
	fmt.Printf("Finished fetching %d calls.\n", len(callsData))
	totalSize := int64(0)
	for _, size := range sizes {
		if size <= 0 {
			log.Fatal("Some calls have invalid sizes, please check the error log for details.")
		}
		totalSize += size
	}
	if totalSize > (1024 * 1024 * 1024) {
		fmt.Printf("Total size of all calls: %.2f GB\n", float64(totalSize)/1024/1024/1024)
	} else if totalSize > (1024 * 1024) {
		fmt.Printf("Total size of all calls: %.2f MB\n", float64(totalSize)/1024/1024)
	} else {
		fmt.Printf("Total size of all calls: %.2f KB\n", float64(totalSize)/1024)
	}
	println("Downloading...")
	p = mpb.New(mpb.WithWidth(64), mpb.PopCompletedMode())
	downloadFunction := func(call any, ctx context.Context, cancel context.CancelFunc, errCh chan error) {
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
		liveIdStr := strconv.Itoa(liveId)
		filepath := *outputDir + "/" + liveIdStr + ".mp4"
		// HEAD request to get content-length
		headReq, _ := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
		resp, err := http.DefaultClient.Do(headReq)
		if err != nil || resp.StatusCode != http.StatusOK {
			log.Printf("HEAD failed for live ID %d: %v", liveId, err)
			cancel()
			return
		}
		length := resp.ContentLength
		resp.Body.Close()
		bar := p.New(length,
			mpb.BarStyle().Lbound("[").Filler("=").Tip(">").Padding(" ").Rbound("]"),
			mpb.PrependDecorators(
				decor.Name(liveIdStr, decor.WC{W: 5, C: decor.DindentRight}),
				decor.Current(decor.SizeB1024(0), "% .1f", decor.WC{W: 11}),
				decor.TotalKibiByte(" / % .1f", decor.WC{W: 14, C: decor.DindentRight}),
				decor.AverageSpeed(decor.SizeB1024(0), "% .1f", decor.WC{W: 13}),
				decor.Elapsed(decor.ET_STYLE_MMSS, decor.WC{W: 10}),
				decor.Name(" ETA: ", decor.WC{W: 6}),
				decor.AverageETA(decor.ET_STYLE_MMSS, decor.WC{W: 9, C: decor.DindentRight}),
			),
			mpb.AppendDecorators(
				decor.NewPercentage("%.2f", decor.WC{W: 7}),
			),
		)
		err = DownloadVideo(ctx, url, filepath, *outputDir, 10, bar)
		if err != nil {
			errCh <- fmt.Errorf("download failed for live ID %d: %w", liveId, err)
			cancel()
			return
		}
		os.Remove(filepath)
	}
	ok, err = concurrentExecute(downloadFunction, callsData, *concurrency)
	if err != nil {
		log.Fatalf("Error during concurrent execution: %v", err)
	}
	if !ok {
		log.Fatal("Some downloads failed, check the error log for details.")
	}
	p.Wait()
	fmt.Printf("Finished downloading %d calls.\n", len(callsData))
}