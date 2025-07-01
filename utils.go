package main

import "github.com/joho/godotenv"

func appendEnv(key, value string) error {
	envFile := ".env"
	envMap, err := godotenv.Read(envFile)
	if err != nil {
		envMap = make(map[string]string)
	}
	envMap[key] = value
	return godotenv.Write(envMap, envFile)
}