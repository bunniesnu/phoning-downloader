package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

const warningConcurrency = 15

func main() {
	outputDir := flag.String("o", "Downloads", "Directory to save downloaded videos")
	concurrency := flag.Int("c", 10, "Concurrent downloads")
	chunk := flag.Int("d", 10, "Number of chunks to download in parallel")
	help := flag.Bool("h", false, "Show help message")
	flag.Parse()
	if *help {
		flag.Usage()
		os.Exit(0)
	}
	if *concurrency < 1 {
		log.Fatal("Concurrency must be at least 1")
	}
	if *concurrency > warningConcurrency {
		color.Yellow("Warning: High concurrency may cause issues. Consider using a lower value.")
	}
	if *chunk < 1 {
		log.Fatal("Chunk size must be at least 1")
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
	num := len(callsData)
	println("Found", num, "calls. Fetching informations...")
	liveIds := make([]int, num)
	callsMap := make(map[int](map[string]any), num)
	for i, call := range callsData {
		callMap, ok := call.(map[string]any)
		if !ok {
			log.Fatalf("Unexpected call format: %T", call)
		}
		liveId, ok := callMap["liveId"].(float64)
		if !ok {
			log.Fatalf("Live ID not found in call: %v", callMap)
		}
		liveIdInt := int(liveId)
		callsMap[liveIdInt] = callMap
		liveIds[i] = liveIdInt
	}
	p := mpb.New(mpb.WithWidth(64), mpb.PopCompletedMode())
	bar := p.New(int64(num),
		mpb.BarStyle().Lbound("[").Filler("=").Tip(">").Padding(" ").Rbound("]"),
		mpb.PrependDecorators(
			decor.Name("Fetching...", decor.WC{W: 5, C: decor.DindentRight}),
			decor.Current(0, "(%d", decor.WC{W: 5}),
			decor.Total(0, "/%d)", decor.WC{W: 5, C: decor.DindentRight}),
		),
		mpb.AppendDecorators(
			decor.NewPercentage("%.2f", decor.WC{W: 7}),
		),
	)
	fetchFunction := func (liveId int, ctx context.Context) (int64, error) {
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
			return 0, fmt.Errorf("HEAD request failed for live ID %d: %v", liveId, err)
		}
		length := resp.ContentLength
		resp.Body.Close()
		bar.IncrInt64(1)
		return length, nil
	}
	sizes, err := concurrentExecute(fetchFunction, liveIds, *concurrency)
	if err != nil {
		log.Fatalf("Error during concurrent execution: %v", err)
	}
	p.Wait()
	println("Finished fetching calls.")
	totalSize := int64(0)
	for _, size := range sizes {
		if size <= 0 {
			log.Fatal("Some calls have invalid sizes, please check the error log for details.")
		}
		totalSize += size
	}
	if totalSize > (1024 * 1024 * 1024) {
		fmt.Printf("Total size of all calls: %.2f GiB\n", float64(totalSize)/1024/1024/1024)
	} else if totalSize > (1024 * 1024) {
		fmt.Printf("Total size of all calls: %.2f MiB\n", float64(totalSize)/1024/1024)
	} else {
		fmt.Printf("Total size of all calls: %.2f KiB\n", float64(totalSize)/1024)
	}
	println("Downloading...")
	p = mpb.New(mpb.WithWidth(64), mpb.PopCompletedMode())
	totalbar := p.New(totalSize,
		mpb.BarStyle().Lbound("[").Filler("=").Tip(">").Padding(" ").Rbound("]"),
		mpb.BarPriority(1000),
		mpb.PrependDecorators(
			decor.Name("", decor.WC{W: 5, C: decor.DindentRight}),
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
	countbar := p.New(int64(num),
		mpb.BarStyle().Padding(" ").Lbound(" ").Filler(" ").Tip(" ").Lbound(" ").Rbound(" "),
		mpb.BarPriority(999),
		mpb.PrependDecorators(
			decor.Name(color.CyanString("Total"), decor.WC{W: 5, C: decor.DindentRight}),
			decor.Current(0, "(%d", decor.WC{W: 5}),
			decor.Total(0, "/%d)", decor.WC{W: 5, C: decor.DindentRight}),
		),
	)
	downloadFunction := func(liveId int, ctx context.Context) (bool, error) {
		pnxml, err := getPNXML(api_key, access_token, liveId)
		if err != nil {
			log.Fatalf("Error getting PNXML for live ID %d: %v", liveId, err)
		}
		url, ok := pnxml["url"].(string)
		if !ok {
			log.Fatalf("PNXML for live ID %d does not contain a valid URL", liveId)
		}
		liveIdStr := strconv.Itoa(liveId)
		downloadFilePath := filepath.Join(*outputDir, liveIdStr+".mp4")
		bar := p.New(sizes[liveId],
			mpb.BarStyle().Lbound("[").Filler("=").Tip(">").Padding(" ").Rbound("]"),
			mpb.PrependDecorators(
				decor.Name(liveIdStr, decor.WC{W: 5, C: decor.DindentRight}),
				decor.Current(decor.SizeB1024(0), "% .1f", decor.WC{W: 11}),
				decor.TotalKibiByte(" / % .1f", decor.WC{W: 14, C: decor.DindentRight}),
				decor.AverageSpeed(decor.SizeB1024(0), "% .1f", decor.WC{W: 13}),
				decor.Elapsed(decor.ET_STYLE_MMSS, decor.WC{W: 10}),
				decor.OnComplete(
					decor.Name(" ETA: "),
					color.GreenString(" Done"),
				),
				decor.OnComplete(
					decor.AverageETA(decor.ET_STYLE_MMSS, decor.WC{W: 9, C: decor.DindentRight}),
					"",
				),
			),
			mpb.BarFillerOnComplete(""),
			mpb.AppendDecorators(
				decor.OnComplete(
					decor.NewPercentage("%.2f", decor.WC{W: 7}),
					"",
				),
			),
		)
		hookTotalProgress(bar, totalbar)
		err = DownloadVideo(ctx, url, downloadFilePath, *outputDir, *chunk, bar)
		if err != nil {
			return false, fmt.Errorf("error downloading live ID %d: %v", liveId, err)
		}
		os.Remove(downloadFilePath)
		countbar.IncrInt64(1)
		return true, nil
	}
	_, err = concurrentExecute(downloadFunction, liveIds, *concurrency)
	if err != nil {
		log.Fatalf("Error during concurrent execution: %v", err)
	}
	p.Wait()
	fmt.Printf("Finished downloading %d calls.\n", num)
}