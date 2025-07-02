package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const baseURL = "https://www.emailnator.com/"
const signupQuery = "account.weverse.io/signup"

type CookieData struct {
	XSRFToken         string
	GmailnatorSession string
}

func extractSignupLinks(html string) ([]string, error) {
    doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
    if err != nil {
        return nil, err
    }

    var links []string
    doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
        href, exists := s.Attr("href")
        if exists && strings.Contains(href, signupQuery) {
            links = append(links, href)
        }
    })

    return links, nil
}

func getCookie() (CookieData, *http.Client, error) {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	resp, err := client.Get(baseURL)
	if err != nil {
		return CookieData{}, nil, err
	}
	defer resp.Body.Close()

	cookies := jar.Cookies(resp.Request.URL)
	var data CookieData
	for _, cookie := range cookies {
		if cookie.Name == "XSRF-TOKEN" {
			data.XSRFToken = strings.ReplaceAll(cookie.Value, "%3D", "=")
		}
		if cookie.Name == "gmailnator_session" {
			data.GmailnatorSession = cookie.Value
		}
	}
	return data, client, nil
}

func genEmail(client *http.Client, cookieData CookieData) (string, error) {
	payload := map[string][]string{"email": {"dotGmail"}}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", baseURL + "generate-email", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Xsrf-Token", cookieData.XSRFToken)
	req.Header.Set("Origin", baseURL)
	req.AddCookie(&http.Cookie{Name: "XSRF-TOKEN", Value: cookieData.XSRFToken})
	req.AddCookie(&http.Cookie{Name: "gmailnator_session", Value: cookieData.GmailnatorSession})

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string][]string
	json.NewDecoder(resp.Body).Decode(&result)

	return result["email"][0], nil
}

func messageLoop(client *http.Client, cookieData CookieData, email string) (string, error) {
	headers := map[string]string{
		"User-Agent":       "Mozilla/5.0",
		"Accept":           "application/json, text/plain, */*",
		"X-Requested-With": "XMLHttpRequest",
		"Content-Type":     "application/json",
		"X-Xsrf-Token":     cookieData.XSRFToken,
		"Origin":           baseURL,
		"Referer":          baseURL,
	}

	seenIDs := loadSeenIDs()

	for {
		body, _ := json.Marshal(map[string]string{"email": email})
		req, _ := http.NewRequest("POST", baseURL + "message-list", bytes.NewBuffer(body))
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		req.AddCookie(&http.Cookie{Name: "XSRF-TOKEN", Value: cookieData.XSRFToken})
		req.AddCookie(&http.Cookie{Name: "gmailnator_session", Value: cookieData.GmailnatorSession})

		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		var data map[string][]map[string]string
		json.NewDecoder(resp.Body).Decode(&data)
		resp.Body.Close()

		for _, msg := range data["messageData"] {
			msgID := msg["messageID"]
			if len(msgID) > 12 && !seenIDs[msgID] {
				seenIDs[msgID] = true
				appendToFile("temp_id.txt", msgID+"\n")

				detailReqBody, _ := json.Marshal(map[string]string{"email": email, "messageID": msgID})
				detailReq, _ := http.NewRequest("POST", baseURL + "message-list", bytes.NewBuffer(detailReqBody))
				for k, v := range headers {
					detailReq.Header.Set(k, v)
				}
				detailReq.AddCookie(&http.Cookie{Name: "XSRF-TOKEN", Value: cookieData.XSRFToken})
				detailReq.AddCookie(&http.Cookie{Name: "gmailnator_session", Value: cookieData.GmailnatorSession})

				detailResp, _ := client.Do(detailReq)
				msgContent, _ := io.ReadAll(detailResp.Body)
				links, err := extractSignupLinks(string(msgContent))
				if err != nil {
					return "", fmt.Errorf("error extracting links: %v", err)
				} else {
					os.Remove("temp_id.txt")
					for _, link := range links {
						return link, nil
					}
				}
				detailResp.Body.Close()
			}
		}
		time.Sleep(5 * time.Second) // polling interval
	}
}

func loadSeenIDs() map[string]bool {
	seen := map[string]bool{}
	data, err := os.ReadFile("temp_id.txt")
	if err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if line != "" {
				seen[line] = true
			}
		}
	}
	return seen
}

func appendToFile(filename, content string) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		defer f.Close()
		f.WriteString(content)
	}
}