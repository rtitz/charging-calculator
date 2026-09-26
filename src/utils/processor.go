package utils

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

	// 1. Daytime Household Totals aus Solar-Timeline integrieren
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
		metrics.HouseConsumptionkWh += currSolar.Record.TotalHouseComsumptionW * timeFactor
	}

	// 2. 24/7 Wallbox-Timeline verarbeiten (nur wenn Daten vorhanden sind)
	if len(wb) > 0 {
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
			wbWatts := (deltaWh / wbDuration) * 3600.0

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
			solarIdx = bestSolarIdx

			wbSolarRatio := 0.0

			if bestDiff <= 600 {
				solarRec := solar[bestSolarIdx].Record
				houseBaseW := solarRec.TotalHouseComsumptionW - wbWatts
				if houseBaseW < 0 {
					houseBaseW = 0
				}

				solarW := solarRec.TotalSolarPowerW
				gridW := solarRec.TotalGridPowerW

				var wbSolarW float64

				if variables.SolarPrioritization == "car" {
					if solarW >= wbWatts {
						wbSolarW = wbWatts
					} else {
						wbSolarW = solarW
					}
				} else {
					solarLeftForWb := solarW - houseBaseW
					if solarLeftForWb > wbWatts {
						wbSolarW = wbWatts
					} else {
						if solarLeftForWb < 0 {
							solarLeftForWb = 0
						}
						wbSolarW = solarLeftForWb
					}
				}

				if gridW <= 0 && solarW > 0 {
					wbSolarW = wbWatts
				}

				if wbWatts > 0 {
					wbSolarRatio = wbSolarW / wbWatts
				}
			} else {
				wbSolarRatio = 0.0
			}

			wbTotalkWh := deltaWh / 1000.0
			metrics.SolarChargedkWh += wbTotalkWh * wbSolarRatio
			metrics.GridChargedkWh += wbTotalkWh * (1.0 - wbSolarRatio)
		}
	}

	// 3. Systembilanzierung ausführen
	metrics.TotalChargedkWh = metrics.SolarChargedkWh + metrics.GridChargedkWh
	if metrics.TotalChargedkWh > 0 {
		metrics.SolarPercentage = (metrics.SolarChargedkWh / metrics.TotalChargedkWh) * 100
	}

	metrics.HouseConsumptionkWh = metrics.HouseConsumptionkWh - metrics.TotalChargedkWh
	if metrics.HouseConsumptionkWh < 0 {
		metrics.HouseConsumptionkWh = 0
	}

	totalSolarSelfConsumed := metrics.TotalSolarGenerationkWh - metrics.TotalGridFeedInkWh
	metrics.HouseSourcedFromSolarkWh = totalSolarSelfConsumed - metrics.SolarChargedkWh
	if metrics.HouseSourcedFromSolarkWh < 0 {
		metrics.HouseSourcedFromSolarkWh = 0
	}
	if metrics.HouseSourcedFromSolarkWh > metrics.HouseConsumptionkWh {
		metrics.HouseSourcedFromSolarkWh = metrics.HouseConsumptionkWh
	}

	metrics.HouseSourcedFromGridkWh = metrics.HouseConsumptionkWh - metrics.HouseSourcedFromSolarkWh

	if metrics.HouseConsumptionkWh > 0 {
		metrics.HouseSolarPercentage = (metrics.HouseSourcedFromSolarkWh / metrics.HouseConsumptionkWh) * 100
	}

	metrics.TotalGridConsumptionkWh = metrics.HouseSourcedFromGridkWh + metrics.GridChargedkWh
	metrics.OverallTotalConsumptionkWh = metrics.HouseConsumptionkWh + metrics.TotalChargedkWh

	if metrics.OverallTotalConsumptionkWh > 0 {
		totalCombinedSolarPowerkWh := metrics.HouseSourcedFromSolarkWh + metrics.SolarChargedkWh
		metrics.TotalSolarPercentage = (totalCombinedSolarPowerkWh / metrics.OverallTotalConsumptionkWh) * 100
	}

	return metrics
}

func ProcessMonthlyData(dirPath string) error {
	solarFiles, wallboxFiles, err := discoverMonthlyFiles(dirPath)
	if err != nil {
		return err
	}

	// 💡 Master-Liste basiert jetzt rein auf den vorhandenen Solar-Dateien
	var months []string
	for m := range solarFiles {
		months = append(months, m)
	}

	sort.Slice(months, func(i, j int) bool {
		if variables.OutputOrder == "asc" {
			return months[i] < months[j]
		}
		return months[i] > months[j]
	})

	for _, month := range months {
		solarPath := solarFiles[month]
		wallboxPath, hasWallbox := wallboxFiles[month]

		solarTimeline, _ := loadSolarTimeline(solarPath)

		// 💡 Wallbox-Timeline bleibt leer, wenn keine Datei existiert
		var wbTimeline []TimedWallboxRecord
		if hasWallbox {
			wbTimeline, _ = loadWallboxTimeline(wallboxPath)
			sort.Slice(wbTimeline, func(i, j int) bool { return wbTimeline[i].Time.Before(wbTimeline[j].Time) })
		}

		sort.Slice(solarTimeline, func(i, j int) bool { return solarTimeline[i].Time.Before(solarTimeline[j].Time) })

		checkTimeOverlap(month, solarTimeline, wbTimeline, hasWallbox)

		metrics := calculateMetrics(month, solarTimeline, wbTimeline)

		// Finanzberechnungen für die Wallbox
		wallboxTheoreticalCostEUR := (metrics.TotalChargedkWh * variables.GridPriceCents) / 100.0
		wallboxActualGridCostEUR := (metrics.GridChargedkWh * variables.GridPriceCents) / 100.0
		netSavingsPerkWhCents := variables.GridPriceCents - variables.SolarExportCreditCents
		wallboxNetSavingsEUR := (metrics.SolarChargedkWh * netSavingsPerkWhCents) / 100.0

		fmt.Printf("\nMonth:                               %s (Priority: %s)\n", metrics.Month, variables.SolarPrioritization)
		fmt.Printf("Total Solar Production:              %.2f kWh\n", metrics.TotalSolarGenerationkWh)
		fmt.Printf("Total Surplus Grid Feed-In (Export): %.2f kWh\n", metrics.TotalGridFeedInkWh)
		fmt.Printf("-----------------------------------------------------------------\n")
		fmt.Printf("HOUSEHOLD BASE METRICS (Daytime Only - Inverter Active):\n")
		fmt.Printf("  ├── Base House Consumption:        %.2f kWh\n", metrics.HouseConsumptionkWh)
		fmt.Printf("  ├── Sourced from Solar:            %.2f kWh\n", metrics.HouseSourcedFromSolarkWh)
		fmt.Printf("  └── Sourced from Grid:             %.2f kWh\n", metrics.HouseSourcedFromGridkWh)
		fmt.Printf("  └── House Base Solar Share:        %.1f%%\n", metrics.HouseSolarPercentage)
		fmt.Printf("-----------------------------------------------------------------\n")
		fmt.Printf("WALLBOX METRICS (24/7 Accurate Tracking):\n")
		fmt.Printf("  ├── Wallbox Total Charged:         %.2f kWh (Kosten ohne Solar: %.2f EUR)\n", metrics.TotalChargedkWh, wallboxTheoreticalCostEUR)
		fmt.Printf("  ├── Sourced from Solar:            %.2f kWh (Netto-Ersparnis:   %.2f EUR)\n", metrics.SolarChargedkWh, wallboxNetSavingsEUR)
		fmt.Printf("  └── Sourced from Grid:             %.2f kWh (Tatsächliche Kosten: %.2f EUR)\n", metrics.GridChargedkWh, wallboxActualGridCostEUR)
		fmt.Printf("  └── Wallbox Solar Share:           %.1f%%\n", metrics.SolarPercentage)
		fmt.Printf("-----------------------------------------------------------------\n")
		fmt.Printf("COMBINED SYSTEM METRICS (Daytime + Wallbox):\n")
		fmt.Printf("  ├── TOTAL CONSUMPTION FROM GRID:   %.2f kWh\n", metrics.TotalGridConsumptionkWh)
		fmt.Printf("  └── OVERALL TOTAL CONSUMPTION:     %.2f kWh\n", metrics.OverallTotalConsumptionkWh)
		fmt.Printf("  └── TOTAL SYSTEM SOLAR SHARE:      %.1f%%\n", metrics.TotalSolarPercentage)
		fmt.Printf("=================================================================\n")
	}
	return nil
}
