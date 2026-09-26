package utils

import (
	"charging-calculator/variables"
	"fmt"
	"math"
	"time"
)

// Declare package-level structs here so they are shared across all utils package files
type TimedSolarRecord struct {
	Time   time.Time
	Record variables.SolarRecord
}

type TimedWallboxRecord struct {
	Time   time.Time
	Record variables.WallboxRecord
}

// checkTimeOverlap prints out a clear comparison of the logging timelines to catch gaps
func checkTimeOverlap(month string, solar []TimedSolarRecord, wb []TimedWallboxRecord, hasWallbox bool) {
	if len(solar) == 0 {
		fmt.Printf("[%s] ⚠️ CRITICAL: Solar timeline tracking is completely empty!\n", month)
		return
	}

	// 💡 Only log critical emptiness alerts if Wallbox dataset was expected to be present
	if hasWallbox && len(wb) == 0 {
		fmt.Printf("[%s] ⚠️ CRITICAL: Wallbox file was found but timeline tracking is empty!\n", month)
		return
	}

	fmt.Printf("\n\n=================================================================\n")
	fmt.Printf("[%s] 🕰️ TIMELINE PROFILE CHECK:\n", month)
	fmt.Printf("  ├── Solar Data Range  : %v   UNTIL   %v\n", solar[0].Time.Format("2006-01-02 15:04:05"), solar[len(solar)-1].Time.Format("2006-01-02 15:04:05"))

	if len(wb) > 0 {
		fmt.Printf("  └── Wallbox Data Range: %v   UNTIL   %v\n", wb[0].Time.Format("2006-01-02 15:04:05"), wb[len(wb)-1].Time.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Printf("  └── Wallbox Data Range: No Wallbox Telemetry Data Available\n")
	}
}

// findClosestWallboxRecord searches the wallbox timeline to find the closest entry
func findClosestWallboxRecord(target time.Time, wb []TimedWallboxRecord, currentIdx int) (TimedWallboxRecord, int, bool) {
	if len(wb) == 0 {
		return TimedWallboxRecord{}, currentIdx, false
	}

	bestIdx := currentIdx
	if bestIdx >= len(wb) {
		bestIdx = len(wb) - 1
	}
	bestDiff := math.Abs(wb[bestIdx].Time.Sub(target).Seconds())

	for i := bestIdx + 1; i < len(wb); i++ {
		diff := math.Abs(wb[i].Time.Sub(target).Seconds())
		if diff < bestDiff {
			bestDiff = diff
			bestIdx = i
		} else if wb[i].Time.After(target) {
			break
		}
	}

	return wb[bestIdx], bestIdx, (bestDiff <= 300)
}
