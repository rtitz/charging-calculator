package variables

import (
	"runtime"
	"time"
)

// App information
const (
	AppName       = "Charging-Calculator"
	AppVersion    = "1.0.0"
	DataSourceDir = "./data"
)

// System information
var (
	GOOS   = runtime.GOOS
	GOARCH = runtime.GOARCH
)

// WallboxRecord maps the raw JSON data fields required for metrics.
type WallboxRecord struct {
	Date string  `json:"Date"`
	Time string  `json:"Time"`
	Wh   float64 `json:"Wh"`  // Accumulated Watt-hours charged into the car
	Amp  float64 `json:"Amp"` // Charging current (used to detect active charging)
}

// SolarRecord maps the raw JSON data fields required for source balancing.
type SolarRecord struct {
	Date                   string  `json:"Date"`
	Time                   string  `json:"Time"`
	TotalSolarPowerW       float64 `json:"TotalSolarPowerW"`       // Current solar generation (Watts)
	TotalHouseComsumptionW float64 `json:"TotalHouseComsumptionW"` // Total house usage (Watts)
	TotalGridPowerW        float64 `json:"TotalGridPowerW"`        // Grid draw (Watts, positive if importing)
}

type MonthlyChargeMetrics struct {
	Month           string
	TotalChargedkWh float64 // Wallbox Total
	SolarChargedkWh float64 // Wallbox from Solar
	GridChargedkWh  float64 // Wallbox from Grid
	SolarPercentage float64 // Wallbox Solar Share %

	// HOUSE SPECIFIC (EXCLUDING WALLBOX)
	HouseConsumptionkWh      float64 // House base load (No car)
	HouseSourcedFromSolarkWh float64 // House base run on Solar
	HouseSourcedFromGridkWh  float64 // House base run on Grid

	// MACRO TOTALS
	TotalSolarGenerationkWh    float64 // Total PV production
	TotalGridFeedInkWh         float64 // Total PV exported to grid
	TotalGridConsumptionkWh    float64 // Combined Grid Import (House + Wallbox)
	OverallTotalConsumptionkWh float64 // Combined Total Energy Used (House + Wallbox)
}

// Helper to parse Go timestamps out of log strings
func (w *WallboxRecord) Timestamp() (time.Time, error) {
	return time.Parse("2006/01/02 15:04:05", w.Date+" "+w.Time)
}
