package main

import (
	"charging-calculator/variables"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func discoverMonthlyFiles(dirPath string) (map[string]string, map[string]string, error) {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, nil, err
	}
	solarFiles := make(map[string]string)
	wallboxFiles := make(map[string]string)
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}
		parts := strings.Split(file.Name(), "-")
		if len(parts) < 3 {
			continue
		}
		monthPrefix := fmt.Sprintf("%s-%s", parts[0], parts[1])
		fullPath := filepath.Join(dirPath, file.Name())
		if strings.Contains(file.Name(), "solar") {
			solarFiles[monthPrefix] = fullPath
		} else if strings.Contains(file.Name(), "wallbox") {
			wallboxFiles[monthPrefix] = fullPath
		}
	}
	return solarFiles, wallboxFiles, nil
}

func calculateMetrics(month string, solar []TimedSolarRecord, wb []TimedWallboxRecord) variables.MonthlyChargeMetrics {
	metrics := variables.MonthlyChargeMetrics{Month: month}

	// 1. Calculate Core Household Statistics from Solar Timeline (Read-Only)
	for i := 1; i < len(solar); i++ {
		prevSolar := solar[i-1]
		currSolar := solar[i]

		durationSec := currSolar.Time.Sub(prevSolar.Time).Seconds()
		if durationSec <= 0 || durationSec > 600 {
			continue
		}
		timeFactor := durationSec / (3600.0 * 1000.0)

		metrics.TotalSolarGenerationkWh += currSolar.Record.TotalSolarPowerW * timeFactor
		if currSolar.Record.TotalGridPowerW < 0 {
			metrics.TotalGridFeedInkWh += (-currSolar.Record.TotalGridPowerW) * timeFactor
		}

		// Temporary accumulation of raw home usage (we will subtract clean car totals later)
		metrics.HouseConsumptionkWh += currSolar.Record.TotalHouseComsumptionW * timeFactor
	}

	// 2. Calculate Wallbox Sourcing Mix using Wallbox as the Master Loop (Read-Only)
	solarIdx := 0
	for i := 1; i < len(wb); i++ {
		prevWb := wb[i-1]
		currWb := wb[i]

		deltaWh := currWb.Record.Wh - prevWb.Record.Wh
		if deltaWh <= 0 {
			continue
		}

		wbDuration := currWb.Time.Sub(prevWb.Time).Seconds()
		if wbDuration <= 0 {
			wbDuration = 1
		}

		// Derive the real, actual wattage draw of the charger
		wbWatts := (deltaWh / wbDuration) * 3600.0

		// Find the solar condition matching the exact moment of this charging event
		bestSolarIdx := solarIdx
		bestDiff := math.Abs(solar[bestSolarIdx].Time.Sub(currWb.Time).Seconds())

		for j := solarIdx + 1; j < len(solar); j++ {
			diff := math.Abs(solar[j].Time.Sub(currWb.Time).Seconds())
			if diff < bestDiff {
				bestDiff = diff
				bestSolarIdx = j
			} else if solar[j].Time.After(currWb.Time) {
				break
			}
		}
		solarIdx = bestSolarIdx // Update pointer to keep lookups fast

		wbSolarRatio := 0.0

		// Check if the closest solar record falls within a reasonable 5-minute window
		if bestDiff <= 300 {
			solarRec := solar[bestSolarIdx].Record

			// Separate base house load from charger load to prioritize home appliances
			houseBaseW := solarRec.TotalHouseComsumptionW - wbWatts
			if houseBaseW < 0 {
				houseBaseW = 0
			}

			solarLeftForWb := solarRec.TotalSolarPowerW - houseBaseW
			if solarLeftForWb < 0 {
				solarLeftForWb = 0
			}

			if wbWatts > 0 {
				wbSolarRatio = solarLeftForWb / wbWatts
				if wbSolarRatio > 1.0 {
					wbSolarRatio = 1.0
				}
			}

			// If back-exporting to the grid, the car runs entirely on solar power
			if solarRec.TotalGridPowerW <= 0 && solarRec.TotalSolarPowerW > 0 {
				wbSolarRatio = 1.0
			}
		} else {
			// No matching daytime solar record means it occurred during an overnight inverter shutdown
			wbSolarRatio = 0.0
		}

		wbTotalkWh := deltaWh / 1000.0
		metrics.SolarChargedkWh += wbTotalkWh * wbSolarRatio
		metrics.GridChargedkWh += wbTotalkWh * (1.0 - wbSolarRatio)
	}

	// 3. Aggregate Final Structural Splits cleanly
	metrics.TotalChargedkWh = metrics.SolarChargedkWh + metrics.GridChargedkWh
	if metrics.TotalChargedkWh > 0 {
		metrics.SolarPercentage = (metrics.SolarChargedkWh / metrics.TotalChargedkWh) * 100
	}

	// Subtract the car's energy from total household load to separate the house base load metrics
	metrics.HouseConsumptionkWh = metrics.HouseConsumptionkWh - metrics.TotalChargedkWh
	if metrics.HouseConsumptionkWh < 0 {
		metrics.HouseConsumptionkWh = 0
	}

	// Deduct the car's solar power to isolate the house's solar coverage share
	// (House has absolute priority on generation)
	totalSolarUsedBySystem := metrics.TotalSolarGenerationkWh - metrics.TotalGridFeedInkWh
	metrics.HouseSourcedFromSolarkWh = totalSolarUsedBySystem - metrics.SolarChargedkWh
	if metrics.HouseSourcedFromSolarkWh < 0 {
		metrics.HouseSourcedFromSolarkWh = 0
	}
	if metrics.HouseSourcedFromSolarkWh > metrics.HouseConsumptionkWh {
		metrics.HouseSourcedFromSolarkWh = metrics.HouseConsumptionkWh
	}

	metrics.HouseSourcedFromGridkWh = metrics.HouseConsumptionkWh - metrics.HouseSourcedFromSolarkWh
	metrics.TotalGridConsumptionkWh = metrics.HouseSourcedFromGridkWh + metrics.GridChargedkWh
	metrics.OverallTotalConsumptionkWh = metrics.HouseConsumptionkWh + metrics.TotalChargedkWh

	return metrics
}

func processMonthlyData(dirPath string) error {
	solarFiles, wallboxFiles, err := discoverMonthlyFiles(dirPath)
	if err != nil {
		return err
	}

	var months []string
	for m := range solarFiles {
		if _, found := wallboxFiles[m]; found {
			months = append(months, m)
		}
	}

	sort.Slice(months, func(i, j int) bool {
		return months[i] > months[j]
	})

	for _, month := range months {
		solarPath := solarFiles[month]
		wallboxPath := wallboxFiles[month]

		solarTimeline, _ := loadSolarTimeline(solarPath)
		wbTimeline, _ := loadWallboxTimeline(wallboxPath)

		sort.Slice(solarTimeline, func(i, j int) bool { return solarTimeline[i].Time.Before(solarTimeline[j].Time) })
		sort.Slice(wbTimeline, func(i, j int) bool { return wbTimeline[i].Time.Before(wbTimeline[j].Time) })

		checkTimeOverlap(month, solarTimeline, wbTimeline)

		metrics := calculateMetrics(month, solarTimeline, wbTimeline)

		fmt.Printf("\nMonth:                               %s\n", metrics.Month)
		fmt.Printf("Total Solar Production:              %.2f kWh\n", metrics.TotalSolarGenerationkWh)
		fmt.Printf("Total Surplus Grid Feed-In (Export): %.2f kWh\n", metrics.TotalGridFeedInkWh)
		fmt.Printf("-----------------------------------------------------------------\n")
		fmt.Printf("HOUSEHOLD BASE METRICS (Excluding Wallbox):\n")
		fmt.Printf("  ├── Base House Consumption:        %.2f kWh\n", metrics.HouseConsumptionkWh)
		fmt.Printf("  ├── Sourced from Solar:            %.2f kWh\n", metrics.HouseSourcedFromSolarkWh)
		fmt.Printf("  └── Sourced from Grid:             %.2f kWh\n", metrics.HouseSourcedFromGridkWh)
		fmt.Printf("-----------------------------------------------------------------\n")
		fmt.Printf("WALLBOX METRICS:\n")
		fmt.Printf("  ├── Wallbox Total Charged:         %.2f kWh\n", metrics.TotalChargedkWh)
		fmt.Printf("  ├── Sourced from Solar:            %.2f kWh\n", metrics.SolarChargedkWh)
		fmt.Printf("  └── Sourced from Grid:             %.2f kWh\n", metrics.GridChargedkWh)
		fmt.Printf("  └── Wallbox Solar Share:           %.1f%%\n", metrics.SolarPercentage)
		fmt.Printf("-----------------------------------------------------------------\n")
		fmt.Printf("COMBINED SYSTEM METRICS:\n")
		fmt.Printf("  ├── TOTAL CONSUMPTION FROM GRID:   %.2f kWh\n", metrics.TotalGridConsumptionkWh)
		fmt.Printf("  └── OVERALL TOTAL CONSUMPTION:     %.2f kWh\n", metrics.OverallTotalConsumptionkWh)
		fmt.Printf("=================================================================\n")
	}
	return nil
}
