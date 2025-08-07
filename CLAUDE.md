# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a **PDU (Power Distribution Unit) data processor** built in Go that processes electrical current measurements from Excel files and generates statistical reports. The codebase has two main applications:

1. **PDU Data Parser** (`main.go`) - Processes raw PDU measurement files (A1.xlsx, B1.xlsx, etc.) into statistical summaries
2. **Monthly Report Filler** (`pkg/report/report.go`) - Merges processed PDU statistics into monthly template reports

## Development Commands

### Basic Usage
```bash
# Process a single PDU file
go run main.go <input_file> [output_file]
go run main.go C3.xlsx                    # Output: total_c3.csv
go run main.go A1.xlsx total_a1.csv       # Custom output name

# Fill monthly template with PDU data
go run pkg/report/report.go total_a1.csv monthly-june-2025.xlsx result.csv

# Process multiple PDUs at once
./script/report.sh
```

### Go Commands
```bash
# Build the main application
go build -o pdu-parser main.go

# Build the report filler
go build -o monthly-filler pkg/report/report.go

# Install dependencies
go mod tidy

# Run with module
go run main.go <args>
```

## Architecture

### Core Data Flow
1. **Raw Excel Input** → **PDU Parser** → **Statistics CSV** → **Monthly Filler** → **Filled Report**

### Key Components

#### PDU Parser (`main.go`)
- **DataProcessor struct**: Handles Excel file parsing and statistical calculations
- Processes headers like "A1 Q1 Current : l1" to extract PDU name, rack, and line type
- Calculates min/max/avg statistics for each rack (Q1-Q18) and line (l1/l2/l3)
- Outputs structured CSV with measurements: "l1 min", "l1 avg", "l1 max", etc.

#### Monthly Report Filler (`pkg/report/report.go`)
- **MonthlyFiller struct**: Merges PDU statistics into Excel/CSV templates
- **PDUData struct**: Holds processed statistics for all line types (L1/L2/L3)
- **PDUSection struct**: Maps PDU locations in monthly template (rows, columns)
- Supports preserving existing data when adding new PDUs to reports
- Auto-calculates summary statistics per rack (min/avg/max across all lines)

### Data Structures
```go
// PDU measurement data organized by rack and line
map[string][]float64  // key: "Q1_l1", "Q2_l2", etc.

// Monthly template sections
type PDUSection struct {
    Name      string  // "A1", "B2", etc.
    HeaderRow int     // Template row with "PDU A1"
    StartRow  int     // First rack row (Q1)
    EndRow    int     // Last rack row (Q18)
}
```

### File Patterns
- **Input files**: `{PDU}.xlsx` (A1.xlsx, B2.xlsx, C3.xlsx)
- **Processed files**: `total_{pdu}.csv` (total_a1.csv, total_b2.csv)
- **Templates**: `monthly*.xlsx` or existing filled reports as CSV
- **Final output**: CSV format with all PDU sections filled

### Column Mapping (Monthly Template)
- Columns 3-11: Current L1/L2/L3 Min/Avg/Max per rack
- Columns 12-14: Summary Current Min/Avg/Max per rack (calculated automatically)

## Processing Workflow

### Single PDU Processing
```bash
go run main.go A1.xlsx          # → total_a1.csv
```

### Multiple PDU Monthly Report
```bash
go run pkg/report/report.go total_a1.csv monthlyjune2025.xlsx filled_monthly_report.csv
go run pkg/report/report.go total_a2.csv                      # Preserves A1 data, adds A2
go run pkg/report/report.go total_b1.csv                      # Preserves A1+A2, adds B1
```

### Command Line Options (Monthly Filler)
- `-t, --template <file>`: Template file (default: monthlyjune2025.xlsx)  
- `-o, --output <file>`: Output file (default: filled_monthly_report.csv)
- `-c, --clean`: Use clean template (don't preserve existing data)
- `-p, --preserve`: Preserve existing data (default)

## Dependencies
- **github.com/xuri/excelize/v2**: Excel file processing
- Go 1.23+ required
- No testing framework currently configured