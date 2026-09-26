package variables

import (
	"runtime"
	"time"
)

const (
	AppName       = "Charging-Calculator"
	AppVersion    = "1.0.0"
	DataSourceDir = "./data"

	// PRIORITIZATION SWITCH ("house" or "car")
	SolarPrioritization = "car"

	// OUTPUT ORDER SWITCH ("asc" or "desc")
	OutputOrder = "asc"
)

var (
	GOOS   = runtime.GOOS
	GOARCH = runtime.GOARCH
)

type WallboxRecord struct {
	Date string  `json:"Date"`
	Time string  `json:"Time"`
	Wh   float64 `json:"Wh"`
	Amp  float64 `json:"Amp"`
}

type SolarRecord struct {
	Date                   string  `json:"Date"`
	Time                   string  `json:"Time"`
	TotalSolarPowerW       float64 `json:"TotalSolarPowerW"`
	TotalHouseComsumptionW float64 `json:"TotalHouseComsumptionW"`
	TotalGridPowerW        float64 `json:"TotalGridPowerW"`
}

type MonthlyChargeMetrics struct {
	Month           string
	TotalChargedkWh float64
	SolarChargedkWh float64
	GridChargedkWh  float64
	SolarPercentage float64 // Wallbox Solar Share %

	// HOUSE SPECIFIC (EXCLUDING WALLBOX)
	HouseConsumptionkWh      float64
	HouseSourcedFromSolarkWh float64
	HouseSourcedFromGridkWh  float64
	HouseSolarPercentage     float64 // NEW: House Base Solar Share %

	// MACRO TOTALS
	TotalSolarGenerationkWh    float64
	TotalGridFeedInkWh         float64
	TotalGridConsumptionkWh    float64
	OverallTotalConsumptionkWh float64
	TotalSolarPercentage       float64 // NEW: Total System Solar Share %
}

func (w *WallboxRecord) Timestamp() (time.Time, error) {
	return time.Parse("2006/01/02 15:04:05", w.Date+" "+w.Time)
}
