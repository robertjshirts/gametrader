package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func defaultSaramaConfig() map[string]string {
	return map[string]string{
		"brokers": "kafka:9092",
		"offerTopic": "offer",
		"userTopic": "user",
	}
}

func ReadSaramaConfig(configFile string) map[string]string {
	m := defaultSaramaConfig()

	file, err := os.Open(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening sarama config file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "#") && len(line) > 0 {
			before, after, found := strings.Cut(line, "=")
			if found {
				parameter := strings.TrimSpace(before)
				value := strings.TrimSpace(after)
				m[parameter] = value
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Failed to read file: %s", err)
		os.Exit(1)
	}

	return m
}

func defaultDatabaseConfig() map[string]string {
	return map[string]string{
		"host": "database",
		"protocol": "tcp",
		"port": "3306",
		"user": "root",
		"password": "password",
		"database": "retro-games",
	}
}

func ReadDatabaseConfig(configFile string) map[string]string {
	m := defaultDatabaseConfig()

	file, err := os.Open(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database config file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "#") && len(line) > 0 {
			before, after, found := strings.Cut(line, "=")
			if found {
				parameter := strings.TrimSpace(before)
				value := strings.TrimSpace(after)
				m[parameter] = value
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Failed to read file: %s", err)
		os.Exit(1)
	}

	return m
}