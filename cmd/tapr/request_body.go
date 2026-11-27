package main

import (
	"os"
	"strings"
)

var requestData string

func init() {
	rootCmd.Flags().StringVarP(&requestData, "data", "d", "", "Request body, or @filename to read from a file")
}

func loadRequestBody(data string) ([]byte, error) {
	if strings.HasPrefix(data, "@") {
		return os.ReadFile(data[1:])
	}
	return []byte(data), nil
}
