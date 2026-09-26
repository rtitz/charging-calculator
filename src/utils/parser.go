package utils

import (
	"bufio"
	"charging-calculator/variables"
	"encoding/json"
	"os"
	"strings"
	"time"
)

// parseLogTime converts "YYYY/MM/DD HH:MM:SS" or "YYYY-MM-DD HH:MM:SS" into a real time.Time object cleanly
func parseLogTime(dateStr, timeStr string) (time.Time, error) {
	normalizedDate := strings.ReplaceAll(dateStr, "/", "-")
	return time.Parse("2006-01-02 15:04:05", normalizedDate+" "+timeStr)
}

// loadSolarTimeline reads a solar data JSON file and parses it into chronological order
func loadSolarTimeline(path string) ([]TimedSolarRecord, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var timeline []TimedSolarRecord
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.Trim(strings.TrimSuffix(strings.TrimSpace(scanner.Text()), ","), "[]")
		if line == "" {
			continue
		}

		var sRec variables.SolarRecord
		if err := json.Unmarshal([]byte(line), &sRec); err == nil {
			if parsedTime, err := parseLogTime(sRec.Date, sRec.Time); err == nil {
				timeline = append(timeline, TimedSolarRecord{Time: parsedTime, Record: sRec})
			}
		}
	}
	return timeline, scanner.Err()
}

// loadWallboxTimeline reads a wallbox data JSON file and parses it into chronological order
func loadWallboxTimeline(path string) ([]TimedWallboxRecord, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var timeline []TimedWallboxRecord
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.Trim(strings.TrimSuffix(strings.TrimSpace(scanner.Text()), ","), "[]")
		if line == "" {
			continue
		}

		var wbRec variables.WallboxRecord
		if err := json.Unmarshal([]byte(line), &wbRec); err == nil {
			if parsedTime, err := parseLogTime(wbRec.Date, wbRec.Time); err == nil {
				timeline = append(timeline, TimedWallboxRecord{Time: parsedTime, Record: wbRec})
			}
		}
	}
	return timeline, scanner.Err()
}
