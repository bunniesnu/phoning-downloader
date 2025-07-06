package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/bunniesnu/go-gmailnator"
)

func register() (map[string]string, error) {
	gmail, err := gmailnator.NewGmailnator()
	if err != nil {
		return nil, fmt.Errorf("error creating Gmailnator client: %v", err)
	}
	err = gmail.GenerateEmail()
	if err != nil {
		return nil, fmt.Errorf("error generating email: %v", err)
	}
	email := gmail.Email.Email
	password := generatePassword(16)
	nickname := generateNickname()
	_, err = signUp(email, password, nickname)
	if err != nil {
		return nil, fmt.Errorf("error signing up: %v", err)
	}
	res := ""
	for range 5 {
		email, err := gmail.GetMails()
		if err != nil {
			return nil, fmt.Errorf("error getting emails: %v", err)
		}
		if email == nil {
			return nil, fmt.Errorf("no emails found")
		}
		for _, mail := range email {
			messageId := mail.Mid
			mailDetails, err := gmail.GetMailBody(messageId)
			if err != nil {
				return nil, fmt.Errorf("error getting mail body for message ID %s: %v", messageId, err)
			}
			if mailDetails == "" {
				return nil, fmt.Errorf("mail body is empty for message ID %s", messageId)
			}
			if strings.Contains(mailDetails, "account.weverse.io/signup") {
				start := strings.Index(mailDetails, "https://account.weverse.io/signup")
				if start == -1 {
					return nil, fmt.Errorf("verification link not found in mail body")
				}
				end := strings.IndexAny(mailDetails[start:], " \"'<")
				if end == -1 {
					res = mailDetails[start:]
				} else {
					res = mailDetails[start : start+end]
				}
				break
			}
		}
		time.Sleep(5 * time.Second) // Wait for 5 seconds before the next iteration
	}
	if res == "" {
		return nil, fmt.Errorf("verification link not found in any emails")
	}
	res = strings.ReplaceAll(res, "&amp;", "&")
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