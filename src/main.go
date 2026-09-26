package main

import (
	"charging-calculator/utils"
	"charging-calculator/variables"
	"fmt"
)

func main() {
	fmt.Printf("%s %s (%s/%s)\n\n", variables.AppName, variables.AppVersion, variables.GOOS, variables.GOARCH)

	if err := utils.ProcessMonthlyData(variables.DataSourceDir); err != nil {
		fmt.Printf("Application Error: %v\n", err)
	}
}
