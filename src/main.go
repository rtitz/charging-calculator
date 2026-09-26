package main

import (
	"charging-calculator/utils"
	"charging-calculator/variables"
	"flag"
	"fmt"
)

func main() {
	fmt.Printf("%s %s (%s/%s)\n\n", variables.AppName, variables.AppVersion, variables.GOOS, variables.GOARCH)

	flag.StringVar(&variables.StartDate, "start", "", "Starting month in YYYY-MM format (e.g., 2026-08)")
	flag.Parse()

	if variables.StartDate != "" {
		fmt.Printf("Filtering telemetry logs. Processing data from %s onwards...\n", variables.StartDate)
	} else {
		fmt.Println("No start date specified. Processing all available telemetry data.")
	}

	if err := utils.ProcessMonthlyData(variables.DataSourceDir); err != nil {
		fmt.Printf("Application Error: %v\n", err)
	}
}
