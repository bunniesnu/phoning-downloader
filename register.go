package main

import (
	"encoding/json"
	"fmt"
)

func register() (map[string]string, error) {
	cookieData, client, err := getCookie()
	if err != nil {
		return nil, fmt.Errorf("error getting cookies: %v", err)
	}
	email, err := genEmail(client, cookieData)
	if err != nil {
		return nil, fmt.Errorf("error generating email: %v", err)
	}
	password := generatePassword(16)
	nickname := generateNickname()
	_, err = signUp(email, password, nickname)
	if err != nil {
		return nil, fmt.Errorf("error signing up: %v", err)
	}
	res, err := messageLoop(client, cookieData, email)
	if err != nil {
		return nil, fmt.Errorf("error in message loop: %v", err)
	}
	err = clickLink(res)
	if err != nil {
		return nil, fmt.Errorf("error clicking link: %v", err)
	}
	val, err := check_verification(email)
	if err != nil {
		return nil, fmt.Errorf("error checking verification: %v", err)
	}
	var body map[string]any
	err = json.Unmarshal(val, &body)
	if err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}
	result, ok := body["emailVerified"].(bool)
	if !ok || !result {
		return nil, fmt.Errorf("verification failed: %v", result)
	}
	return map[string]string{
		"email": email,
		"password": password,
	}, nil
}