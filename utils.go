package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Checks and return params in expected format, exclude all params with incorrect format
func checkParams(params map[string]string) map[string]string {
	if params != nil {
		result := make(map[string]string)
		for k, v := range params {
			if envVariableFormat.MatchString(k) {
				result[strings.ToUpper(k)] = v
			}
		}
		return result
	}
	return nil
}

// Saves params into file in expected format
func saveParams(path string, params map[string]string) error {
	if params != nil {
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		defer file.Close()

		for k, v := range params {
			if _, err := file.WriteString(fmt.Sprintf("%s=%s\n", k, v)); err != nil {
				return err
			}
		}
	}
	return nil
}

// Loads params from file into map, if not found, leave method without drama
func loadParams(path string) map[string]string {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()

	result := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		splitIndex := strings.Index(line, "=")
		if splitIndex >= 0 {
			result[line[:splitIndex]] = line[splitIndex+1:]
		}
	}
	return result
}
