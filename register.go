package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/bunniesnu/go-gmailnator"
	"github.com/bunniesnu/weverse-api"
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
	w, err := weverse.New(email, password, "", 0)
	if err != nil {
		return nil, fmt.Errorf("error creating Weverse client: %v", err)
	}
	nickname, err := w.GetAccountNicknameSuggestion()
	if err != nil {
		return nil, fmt.Errorf("error getting nickname suggestion")
	}
	w.Nickname = nickname
	err = w.CreateAccount()
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
	val, err := w.GetAccountStatus()
	if err != nil {
		return nil, fmt.Errorf("error checking verification: %v", err)
	}
	if !(val.EmailVerified) {
		return nil, fmt.Errorf("email verification failed")
	}
	return map[string]string{
		"email": email,
		"password": password,
	}, nil
}