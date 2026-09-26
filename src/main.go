package main

import (
	"charging-calculator/utils"
	"charging-calculator/variables"
	"fmt"
)

func main() {
	if err := utils.ProcessMonthlyData(variables.DataSourceDir); err != nil {
		fmt.Printf("Application Error: %v\n", err)
	}
}
