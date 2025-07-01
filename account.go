package main

import (
	"encoding/json"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

var DefaultHeaders = map[string]string{
	"Host": "sdk.weverse.io",
	"Accept": "*/*",
	"X-SDK-SERVICE-ID": "phoning",
	"X-SDK-LANGUAGE": "ko",
	"X-CLOG-USER-DEVICE-ID": "1",
	"X-SDK-PLATFORM": "iOS",
	"Accept-Language": "ko-KR,ko;q=0.9",
	"Accept-Encoding": "gzip, deflate, br",
	"Content-Type": "application/json",
	"X-SDK-VERSION": "3.4.2",
	"User-Agent": "Phoning/20201014 CFNetwork/3826.500.131 Darwin/24.5.0",
	"Connection": "keep-alive",
	"X-SDK-TRACE-ID": "1",
	"X-SDK-APP-VERSION": "2.2.1",
	"Pragma": "no-cache",
	"Cache-Control": "no-cache",
}

func getHeaders() map[string]string {
	godotenv.Load()
	headers := make(map[string]string)
	for k, v := range DefaultHeaders {
		headers[k] = v
	}
	headers["X-SDK-SERVICE-SECRET"] = os.Getenv("PHONING_SDK_SECRET")
	return headers
}

func signUp(email, password, nickname string) ([]byte, error) {
    body := map[string]interface{}{
        "idToken": nil,
        "email":    email,
        "password": password,
        "nickname": nickname,
        "termsAgreements": []map[string]interface{}{
            {
                "termsDocumentId": "ACC-1:ko:3",
                "agreed":          true,
            },
            {
                "termsDocumentId": "ACC-2:ko:4",
                "agreed":          true,
            },
            {
                "termsDocumentId": "phoning-1:ko:1",
                "agreed":          true,
            },
            {
                "termsDocumentId": "phoning-2:ko:1",
                "agreed":          true,
            },
            {
                "termsDocumentId": "phoning-3:ko:1",
                "agreed":          true,
            },
            {
                "termsDocumentId": "phoning-age14:ko:1",
                "agreed":          true,
            },
        },
    }
	encodedBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return CallAPI("POST", "https://sdk.weverse.io/api/v3/signup/by-credentials", encodedBody, getHeaders())
}

func check_verification(email string) ([]byte, error) {
	queryUrl := "https://sdk.weverse.io/api/v1/signup/email/status?email=" + url.QueryEscape(email)
	return CallAPI("GET", queryUrl, nil, getHeaders())
}

func getToken(email, password string) ([]byte, error) {
	body := map[string]interface{}{
		"email":    email,
		"password": password,
	}
	encodedBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return CallAPI("POST", "https://sdk.weverse.io/api/v2/auth/token/by-credentials", encodedBody, getHeaders())
}