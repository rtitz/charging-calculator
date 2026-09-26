package generalUtils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Function go get current time and date, return a string for date and a string for time
func GetDateAndTime() (string, string, string, string, string) {
	currentTime := time.Now()
	dateNow := currentTime.Format("2006/01/02")
	timeNow := currentTime.Format("15:04:05")
	timezoneNow, _ := currentTime.Zone()
	month := currentTime.Format("01")
	year := currentTime.Format("2006")
	return dateNow, timeNow, timezoneNow, month, year
}

func CreateJSONReaderKeyValue(key, value string) *strings.Reader {
	jsonStr := fmt.Sprintf(`{
        	"%s": "%s"
    	}`, key, value)
	return strings.NewReader(jsonStr)
}

// WriteJsonToFile writes JSON data to a file, creates file if it doesn't exist
func WriteJsonToFile(data []byte, filename string) error {
	// Ensure directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Create file if not exists, or open for writing
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write JSON data to file
	data = append(data, ',')
	data = append(data, '\n')
	_, err = file.Write(data)
	return err
}
