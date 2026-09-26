package main

import (
	"charging-calculator/variables"
	"fmt"
)

func main() {
	if err := processMonthlyData(variables.DataSourceDir); err != nil {
		fmt.Printf("Application Error: %v\n", err)
	}
}
