package main

import (
	"encoding/json"
	"fmt"
	"net/url"
)

func getAPIHeaders(accessToken string) map[string]string {
	return map[string]string{
		"Host": "apis.naver.com",
		"Content-Type": "application/json; charset=utf-8",
		"X-Client-Name": "IOS",
		"X-Client-Version": "2.1.2",
		"Connection": "keep-alive",
		"Accept": "application/json",
		"Accept-Language": "ko-KR,ko;q=0.9",
		"Authorization": "Bearer " + accessToken,
		"Accept-Encoding": "gzip, deflate, br",
		"User-Agent": "Phoning/20102019 CFNetwork/1496.0.7 Darwin/23.5.0",
	}
}

func phoning(apiKey, accessToken, endpoint string, params ...map[string]string) (map[string]interface{}, error) {
	var paramMap map[string]string
	if len(params) > 0 && params[0] != nil {
		paramMap = params[0]
	} else {
		paramMap = make(map[string]string)
	}
	values := url.Values{}
	for k, v := range paramMap {
		values.Set(k, v)
	}
	encodeUrl := "https://apis.naver.com/phoning/phoning-api/api" + endpoint
	if len(values) > 0 {
		encodeUrl += "?" + values.Encode()
	}
	h := hash(encodeUrl, apiKey)
	hashValues := url.Values{}
	hashValues.Set("msgpad", h["msgpad"].(string))
	hashValues.Set("md", h["md"].(string))
	queryUrl := encodeUrl
	if len(values) > 0 {
		queryUrl += "&" + hashValues.Encode()
	} else {
		queryUrl += "?" + hashValues.Encode()
	}
	respBody, err := CallAPI("GET", queryUrl, nil, getAPIHeaders(accessToken))
	if err != nil {
		return nil, err
	}
	var response map[string]interface{}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to decode phoning API response: %w, %s", err, string(respBody))
	}
	return response, nil
}