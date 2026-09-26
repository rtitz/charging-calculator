# 🚗 Charging Calculator & Solar Intelligence Dashboard

A high-performance, privacy-respecting Go utility designed to parse millions of rows of high-frequency telemetry logs from your solar inverter and EV Wallbox. It matches time-series data streams dynamically to calculate the exact sourcing breakdown of your energy consumption down to the second.

## 📊 Sample Output Dashboard

```text
=================================================================
Month:                               2026-08 (Priority: car)
Total Solar Production:              1164.27 kWh
Total Surplus Grid Feed-In (Export): 777.14 kWh
-----------------------------------------------------------------
HOUSEHOLD BASE METRICS (Excluding Wallbox):
  ├── Base House Consumption:        312.86 kWh
  ├── Sourced from Solar:            251.44 kWh
  └── Sourced from Grid:             61.42 kWh
  └── House Base Solar Share:        80.4%
-----------------------------------------------------------------
WALLBOX METRICS:
  ├── Wallbox Total Charged:         388.79 kWh
  ├── Sourced from Solar:            135.68 kWh
  └── Sourced from Grid:             253.11 kWh
  └── Wallbox Solar Share:           34.9%
-----------------------------------------------------------------
COMBINED SYSTEM METRICS:
  ├── TOTAL CONSUMPTION FROM GRID:   314.52 kWh
  └── OVERALL TOTAL CONSUMPTION:     701.65 kWh
  └── TOTAL SYSTEM SOLAR SHARE:      55.2%
=================================================================
```

## ✨ Core Features

- **Fuzzy Timeline Synchronization:** Automatically aligns independent logger timelines using a rolling look-ahead pointer with configurable time windows (bridges uneven logging variations smoothly).
- **Overnight Gap Bridging:** Detects automated overnight inverter shutdowns (e.g., when PV production drops to zero) and automatically assigns overnight car charging events to 100% grid power seamlessly without throwing out records.
- **Dynamic Prioritization Matrix:** Toggle load priority profiles via configuration flags (`"house"` to favor standard home appliances first, or `"car"` to direct pure solar margins directly to your EV).
- **Flexible Ordering Logic:** Read and print historical month sets sequentially in ascending (`"asc"`) or descending (`"desc"`) chronological order.
- **Timeline Range Filtering:** Trim metric computation scopes gracefully on the fly via execution args (`-start YYYY-MM`) without deleting historical data sets.
- **Ultra-light Memory Profile:** Utilizes read-only data streams and linear step indexing to process files containing over **1,300,000+ entries in seconds** on a standard MacBook Air.

## 📂 Project Architecture

Organize your workspace files according to the standard Go package blueprint below:

```text
. (Project Root)
├── data/
│   ├── 2026-07-solar.json
│   ├── 2026-07-wallbox.json
│   ├── 2026-08-solar.json
│   └── 2026-08-wallbox.json
└── src/
    ├── go.mod
    ├── main.go            # Entry point orchestration execution & flag registration
    ├── processor.go       # Core metrics compilation engine loops
    ├── parser.go          # High-speed file log decoding scanner
    ├── matching.go        # Time-based multi-timeline pointer pairing
    └── variables/
        └── variables.go   # Mutable global configurations and runtime flags
```

## ⚙️ Configuration (`src/variables/variables.go`)

Default code behaviors can be adjusted inside your configuration file using global mutable variables:

```go
var (
	DataSourceDir = "../data" // Location of raw telemetry JSON arrays

	// PRIORITIZATION SWITCH
	// Options: "house" (Priority to home appliances) or "car" (Priority to EV charging)
	SolarPrioritization = "car" 

	// OUTPUT ORDER SWITCH 
	// Options: "asc" (Oldest month first) or "desc" (Newest month first)
	OutputOrder = "asc" 

	// CHRONOLOGICAL FILTER BOUNDARY
	// Holds runtime boundary from CLI parameter (e.g., "2026-08"). Blank evaluates all history.
	StartDate = ""
)
```

## 🚀 Quick Start

1. Ensure your telemetry files match the chronological name standard inside your data folder: `YYYY-MM-solar.json` and `YYYY-MM-wallbox.json`.
2. Open your terminal window and step directly inside your logic development engine tree:
   ```bash
   cd src
   ```
3. Run the application bundle using one of the following compilation strategies:

   * **Process all historical logs:**
     ```bash
     go run .
     ```
   * **Filter processing from a specific month onwards:**
     ```bash
     go run . -start 2026-08
     ```

## 🧠 Data Processing Logic

1. **Instantaneous Power Derivation:** Wallbox power draw (W) is dynamically computed per interval segment relative to absolute log duration:
   \[\text{Wallbox Power} = \frac{\Delta\text{Wh}}{\Delta\text{Time in Seconds}} \times 3600\]
2. **Priority Apportionment:** Sourcing matrices split power segments. If configuration is set to `"car"`, the wallbox captures pure solar margins first, shifting residual household margins onto the grid whenever solar generation thresholds drop below current load requirements.
3. **Surplus Grid Export Guard:** Whenever the property is actively back-exporting energy to the public network (`TotalGridPowerW <= 0`), the system automatically forces a 100% pure solar score allocation to all active consumption units to bypass sensor rounding margins.
