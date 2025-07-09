// go.mod file - create this as a separate file
/*
module monthly-filler

go 1.19

require github.com/xuri/excelize/v2 v2.8.0

require (
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/richardlehane/mscfb v1.0.4 // indirect
	github.com/richardlehane/msoleps v1.0.3 // indirect
	github.com/xuri/efp v0.0.0-20230802181842-ad255f2331ca // indirect
	github.com/xuri/nfp v0.0.0-20230819163627-dc951e3ffe1a // indirect
	golang.org/x/crypto v0.12.0 // indirect
	golang.org/x/net v0.14.0 // indirect
	golang.org/x/text v0.12.0 // indirect
)
*/

package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// PDUData holds the processed statistics for a PDU
type PDUData struct {
	PDUName string
	L1Min   []float64 // Q1-Q18 values
	L1Avg   []float64
	L1Max   []float64
	L2Min   []float64
	L2Avg   []float64
	L2Max   []float64
	L3Min   []float64
	L3Avg   []float64
	L3Max   []float64
}

// PDUSection represents a PDU section in the template
type PDUSection struct {
	Name      string
	HeaderRow int
	StartRow  int // First Q1 row
	EndRow    int // Last Q18 row
}

// MonthlyFiller handles filling PDU data into monthly template
type MonthlyFiller struct {
	pduData     PDUData
	monthlyData [][]interface{}
	pduSections []PDUSection
}

// NewMonthlyFiller creates a new filler instance
func NewMonthlyFiller() *MonthlyFiller {
	return &MonthlyFiller{}
}

// ExtractPDUNameFromFilename extracts PDU name from filename like "total_a1.csv" -> "A1"
func (mf *MonthlyFiller) ExtractPDUNameFromFilename(filename string) (string, error) {
	basename := filepath.Base(filename)
	basename = strings.TrimSuffix(basename, filepath.Ext(basename))

	// Match patterns like "total_a1", "total_b2", "a3", "A1", etc.
	re := regexp.MustCompile(`(?i)([abc]\d+)`)
	matches := re.FindStringSubmatch(basename)

	if len(matches) < 2 {
		return "", fmt.Errorf("could not extract PDU name from filename: %s", filename)
	}

	pduName := strings.ToUpper(matches[1])
	return pduName, nil
}

// LoadPDUData loads the processed PDU statistics from CSV
func (mf *MonthlyFiller) LoadPDUData(filename string) error {
	// Extract PDU name from filename
	pduName, err := mf.ExtractPDUNameFromFilename(filename)
	if err != nil {
		return err
	}
	mf.pduData.PDUName = pduName

	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open PDU data file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV: %v", err)
	}

	if len(records) < 10 { // Header + 9 measurement types
		return fmt.Errorf("insufficient data in PDU CSV file")
	}

	// Initialize slices for Q1-Q18 (18 values each)
	mf.pduData.L1Min = make([]float64, 18)
	mf.pduData.L1Avg = make([]float64, 18)
	mf.pduData.L1Max = make([]float64, 18)
	mf.pduData.L2Min = make([]float64, 18)
	mf.pduData.L2Avg = make([]float64, 18)
	mf.pduData.L2Max = make([]float64, 18)
	mf.pduData.L3Min = make([]float64, 18)
	mf.pduData.L3Avg = make([]float64, 18)
	mf.pduData.L3Max = make([]float64, 18)

	// Parse each measurement type row
	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) < 19 { // Measurement Type + Q1-Q18
			continue
		}

		measurementType := strings.TrimSpace(record[0])

		// Parse Q1-Q18 values (columns 1-18)
		var targetSlice []float64
		switch measurementType {
		case "l1 min":
			targetSlice = mf.pduData.L1Min
		case "l1 avg":
			targetSlice = mf.pduData.L1Avg
		case "l1 max":
			targetSlice = mf.pduData.L1Max
		case "l2 min":
			targetSlice = mf.pduData.L2Min
		case "l2 avg":
			targetSlice = mf.pduData.L2Avg
		case "l2 max":
			targetSlice = mf.pduData.L2Max
		case "l3 min":
			targetSlice = mf.pduData.L3Min
		case "l3 avg":
			targetSlice = mf.pduData.L3Avg
		case "l3 max":
			targetSlice = mf.pduData.L3Max
		default:
			continue
		}

		// Fill the values for Q1-Q18
		for j := 1; j < 19; j++ { // columns 1-18 (Q1-Q18)
			if j < len(record) {
				value, err := strconv.ParseFloat(strings.TrimSpace(record[j]), 64)
				if err == nil {
					targetSlice[j-1] = value // j-1 because array is 0-indexed
				}
			}
		}
	}

	fmt.Printf("Loaded PDU %s data from %s\n", mf.pduData.PDUName, filename)
	return nil
}

// loadExcelFile loads an Excel file and returns rows as string arrays
func (mf *MonthlyFiller) loadExcelFile(filename string) ([][]string, error) {
	f, err := excelize.OpenFile(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("no sheets found")
	}

	return f.GetRows(sheets[0])
}

// loadCSVFile loads a CSV file and returns rows as string arrays
func (mf *MonthlyFiller) loadCSVFile(filename string) ([][]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	return reader.ReadAll()
}

// LoadMonthlyTemplate loads the monthly Excel template and identifies PDU sections
// If preserveExisting is true, it will load an existing filled template to preserve previous data
func (mf *MonthlyFiller) LoadMonthlyTemplate(filename string, preserveExisting bool, outputFile string) error {
	var targetFile string

	// Decide which file to load based on preserveExisting flag and file existence
	if preserveExisting && outputFile != "" {
		if _, err := os.Stat(outputFile); err == nil {
			// Output file exists, use it to preserve previous data
			targetFile = outputFile
			fmt.Printf("Loading existing filled template: %s (preserving previous data)\n", targetFile)
		} else {
			// Output file doesn't exist, use clean template
			targetFile = filename
			fmt.Printf("Output file doesn't exist, using clean template: %s\n", targetFile)
		}
	} else {
		// Use clean template
		targetFile = filename
		fmt.Printf("Using clean template: %s\n", targetFile)
	}

	// Handle different file formats
	var rows [][]string
	var err error

	if strings.HasSuffix(strings.ToLower(targetFile), ".csv") {
		// Load CSV file (existing filled template)
		rows, err = mf.loadCSVFile(targetFile)
		if err != nil {
			return fmt.Errorf("failed to load CSV file: %v", err)
		}
	} else {
		// Load Excel file (clean template)
		rows, err = mf.loadExcelFile(targetFile)
		if err != nil {
			return fmt.Errorf("failed to load Excel file: %v", err)
		}
	}

	// Convert to interface{} format
	mf.monthlyData = make([][]interface{}, len(rows))
	for i, row := range rows {
		mf.monthlyData[i] = make([]interface{}, len(row))
		for j, cell := range row {
			// Try to parse as float first, then keep as string
			if cell != "" {
				if val, err := strconv.ParseFloat(cell, 64); err == nil {
					mf.monthlyData[i][j] = val
				} else {
					mf.monthlyData[i][j] = cell
				}
			} else {
				mf.monthlyData[i][j] = nil
			}
		}

		// Ensure all rows have at least 15 columns (including summary columns)
		for len(mf.monthlyData[i]) < 15 {
			mf.monthlyData[i] = append(mf.monthlyData[i], nil)
		}
	}

	// Identify all PDU sections
	mf.identifyPDUSections()

	fmt.Printf("Loaded template from %s (%d rows, %d PDU sections)\n",
		targetFile, len(mf.monthlyData), len(mf.pduSections))
	return nil
}

// identifyPDUSections scans the template to find all PDU sections
func (mf *MonthlyFiller) identifyPDUSections() {
	mf.pduSections = []PDUSection{}

	for i := 0; i < len(mf.monthlyData); i++ {
		if len(mf.monthlyData[i]) == 0 {
			continue
		}

		firstCell := mf.monthlyData[i][0]
		if firstCell == nil {
			continue
		}

		cellStr := fmt.Sprintf("%v", firstCell)
		if strings.HasPrefix(cellStr, "PDU ") {
			// Found a PDU header
			pduName := strings.TrimPrefix(cellStr, "PDU ")
			startRow := i + 1       // Q1 starts on next row
			endRow := startRow + 17 // Q1-Q18 = 18 rows

			section := PDUSection{
				Name:      pduName,
				HeaderRow: i,
				StartRow:  startRow,
				EndRow:    endRow,
			}
			mf.pduSections = append(mf.pduSections, section)

			fmt.Printf("Found PDU %s: header at row %d, data rows %d-%d\n",
				pduName, section.HeaderRow, section.StartRow, section.EndRow)
		}
	}
}

// FindPDUSection finds the section for the specified PDU name
func (mf *MonthlyFiller) FindPDUSection(pduName string) (*PDUSection, error) {
	for i := range mf.pduSections {
		if mf.pduSections[i].Name == pduName {
			return &mf.pduSections[i], nil
		}
	}
	return nil, fmt.Errorf("PDU section '%s' not found in template", pduName)
}

// FillPDUData fills the PDU statistics into the correct section of the monthly template
func (mf *MonthlyFiller) FillPDUData() error {
	// Find the target PDU section
	section, err := mf.FindPDUSection(mf.pduData.PDUName)
	if err != nil {
		return err
	}

	fmt.Printf("Filling data into PDU %s section (rows %d-%d)\n",
		section.Name, section.StartRow, section.EndRow)

	// Column mapping based on the template structure
	columnMapping := map[string]int{
		"L1Min": 3,  // Current L1 Min
		"L1Avg": 4,  // Current L1 AVG
		"L1Max": 5,  // Current L1 Max
		"L2Min": 6,  // Current L2 Min
		"L2Avg": 7,  // Current L2 AVG
		"L2Max": 8,  // Current L2 Max
		"L3Min": 9,  // Current L3 Min
		"L3Avg": 10, // Current L3 AVG
		"L3Max": 11, // Current L3 Max
		// Summary per Rack columns
		"SummaryMin": 12, // Current Min (summary)
		"SummaryAvg": 13, // Current AVG (summary)
		"SummaryMax": 14, // Current Max (summary)
	}

	// Fill data for Q1-Q18
	for q := 0; q < 18; q++ {
		rowIndex := section.StartRow + q // Q1 at StartRow, Q2 at StartRow+1, etc.

		if rowIndex >= len(mf.monthlyData) {
			return fmt.Errorf("row index %d exceeds template size", rowIndex)
		}

		// Ensure row has enough columns
		for len(mf.monthlyData[rowIndex]) < 15 {
			mf.monthlyData[rowIndex] = append(mf.monthlyData[rowIndex], nil)
		}

		// Fill L1, L2, L3 data
		l1Min := mf.pduData.L1Min[q]
		l1Avg := mf.pduData.L1Avg[q]
		l1Max := mf.pduData.L1Max[q]
		l2Min := mf.pduData.L2Min[q]
		l2Avg := mf.pduData.L2Avg[q]
		l2Max := mf.pduData.L2Max[q]
		l3Min := mf.pduData.L3Min[q]
		l3Avg := mf.pduData.L3Avg[q]
		l3Max := mf.pduData.L3Max[q]

		mf.monthlyData[rowIndex][columnMapping["L1Min"]] = l1Min
		mf.monthlyData[rowIndex][columnMapping["L1Avg"]] = l1Avg
		mf.monthlyData[rowIndex][columnMapping["L1Max"]] = l1Max
		mf.monthlyData[rowIndex][columnMapping["L2Min"]] = l2Min
		mf.monthlyData[rowIndex][columnMapping["L2Avg"]] = l2Avg
		mf.monthlyData[rowIndex][columnMapping["L2Max"]] = l2Max
		mf.monthlyData[rowIndex][columnMapping["L3Min"]] = l3Min
		mf.monthlyData[rowIndex][columnMapping["L3Avg"]] = l3Avg
		mf.monthlyData[rowIndex][columnMapping["L3Max"]] = l3Max

		// Calculate summary per rack
		// Current Min = minimum of all min values (L1, L2, L3)
		summaryMin := mf.minValue(l1Min, l2Min, l3Min)

		// Current AVG = average of all avg values (L1, L2, L3)
		summaryAvg := (l1Avg + l2Avg + l3Avg) / 3.0

		// Current Max = maximum of all max values (L1, L2, L3)
		summaryMax := mf.maxValue(l1Max, l2Max, l3Max)

		// Fill summary columns
		mf.monthlyData[rowIndex][columnMapping["SummaryMin"]] = summaryMin
		mf.monthlyData[rowIndex][columnMapping["SummaryAvg"]] = summaryAvg
		mf.monthlyData[rowIndex][columnMapping["SummaryMax"]] = summaryMax
	}

	fmt.Printf("Successfully filled %s data into PDU %s section (including summary calculations)\n",
		mf.pduData.PDUName, section.Name)
	return nil
}

// minValue returns the minimum of three float64 values
func (mf *MonthlyFiller) minValue(a, b, c float64) float64 {
	min := a
	if b < min {
		min = b
	}
	if c < min {
		min = c
	}
	return min
}

// maxValue returns the maximum of three float64 values
func (mf *MonthlyFiller) maxValue(a, b, c float64) float64 {
	max := a
	if b > max {
		max = b
	}
	if c > max {
		max = c
	}
	return max
}

// ExportToCSV exports the filled template to CSV format
func (mf *MonthlyFiller) ExportToCSV(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Convert interface{} data to strings and write to CSV
	for _, row := range mf.monthlyData {
		record := make([]string, len(row))
		for i, cell := range row {
			if cell == nil {
				record[i] = ""
			} else {
				switch v := cell.(type) {
				case string:
					record[i] = v
				case float64:
					record[i] = fmt.Sprintf("%.3f", v)
				case int:
					record[i] = fmt.Sprintf("%d", v)
				default:
					record[i] = fmt.Sprintf("%v", v)
				}
			}
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV row: %v", err)
		}
	}

	fmt.Printf("Exported filled data to %s\n", filename)
	return nil
}

// ProcessFiles is the main processing function
func (mf *MonthlyFiller) ProcessFiles(pduDataFile, monthlyFile, outputFile string, preserveExisting bool) error {
	// Load PDU data
	if err := mf.LoadPDUData(pduDataFile); err != nil {
		return fmt.Errorf("error loading PDU data: %v", err)
	}

	// Load monthly template (with option to preserve existing data)
	if err := mf.LoadMonthlyTemplate(monthlyFile, preserveExisting, outputFile); err != nil {
		return fmt.Errorf("error loading monthly template: %v", err)
	}

	// Check if target PDU section already has data
	section, err := mf.FindPDUSection(mf.pduData.PDUName)
	if err != nil {
		return fmt.Errorf("error finding PDU section: %v", err)
	}

	// Check if this section already has data
	hasExistingData := false
	if section.StartRow < len(mf.monthlyData) {
		firstDataRow := mf.monthlyData[section.StartRow]
		if len(firstDataRow) > 3 && firstDataRow[3] != nil {
			hasExistingData = true
		}
	}

	if hasExistingData {
		fmt.Printf("⚠️  PDU %s section already contains data - it will be overwritten\n", mf.pduData.PDUName)
	}

	// Fill PDU data into template
	if err := mf.FillPDUData(); err != nil {
		return fmt.Errorf("error filling PDU data: %v", err)
	}

	// Export to CSV
	if err := mf.ExportToCSV(outputFile); err != nil {
		return fmt.Errorf("error exporting to CSV: %v", err)
	}

	return nil
}

func main() {
	// Check command line arguments
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <pdu_data_file> [options]\n", os.Args[0])
		fmt.Printf("\nOptions:\n")
		fmt.Printf("  -t, --template <file>    Monthly template file (default: monthlyjune2025.xlsx)\n")
		fmt.Printf("  -o, --output <file>      Output file (default: filled_monthly_report.csv)\n")
		fmt.Printf("  -c, --clean              Use clean template (don't preserve existing data)\n")
		fmt.Printf("  -p, --preserve           Preserve existing data (default behavior)\n")
		fmt.Printf("\nExamples:\n")
		fmt.Printf("  %s total_a1.csv                              # First PDU - creates new report\n", os.Args[0])
		fmt.Printf("  %s total_a2.csv                              # Second PDU - adds to existing report\n", os.Args[0])
		fmt.Printf("  %s total_a3.csv -o custom_report.csv         # Custom output file\n", os.Args[0])
		fmt.Printf("  %s total_a1.csv --clean                      # Force clean template\n", os.Args[0])
		fmt.Printf("\nWorkflow for multiple PDUs:\n")
		fmt.Printf("  %s total_a1.csv                              # Creates filled_monthly_report.csv with A1 data\n", os.Args[0])
		fmt.Printf("  %s total_a2.csv                              # Adds A2 data, keeps A1 data\n", os.Args[0])
		fmt.Printf("  %s total_b1.csv                              # Adds B1 data, keeps A1+A2 data\n", os.Args[0])
		fmt.Printf("\nSupported PDU files: total_a1.csv, total_a2.csv, total_b1.csv, etc.\n")
		os.Exit(1)
	}

	// Parse arguments
	pduDataFile := os.Args[1]
	monthlyFile := "monthlyjune2025.xlsx"     // Default template
	outputFile := "filled_monthly_report.csv" // Default output
	preserveExisting := true                  // Default: preserve existing data

	// Parse optional arguments
	for i := 2; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch arg {
		case "-t", "--template":
			if i+1 < len(os.Args) {
				monthlyFile = os.Args[i+1]
				i++ // Skip next argument
			}
		case "-o", "--output":
			if i+1 < len(os.Args) {
				outputFile = os.Args[i+1]
				i++ // Skip next argument
			}
		case "-c", "--clean":
			preserveExisting = false
		case "-p", "--preserve":
			preserveExisting = true
		default:
			// Assume it's a positional argument for backward compatibility
			if i == 2 && !strings.HasPrefix(arg, "-") {
				monthlyFile = arg
			} else if i == 3 && !strings.HasPrefix(arg, "-") {
				outputFile = arg
			}
		}
	}

	// Validate input files exist
	if _, err := os.Stat(pduDataFile); os.IsNotExist(err) {
		log.Fatalf("PDU data file %s not found", pduDataFile)
	}

	if _, err := os.Stat(monthlyFile); os.IsNotExist(err) {
		log.Fatalf("Monthly template file %s not found", monthlyFile)
	}

	// Create filler and process
	filler := NewMonthlyFiller()
	if err := filler.ProcessFiles(pduDataFile, monthlyFile, outputFile, preserveExisting); err != nil {
		log.Fatalf("Processing failed: %v", err)
	}

	fmt.Printf("\n=== Processing completed successfully! ===\n")
	fmt.Printf("PDU Data Source: %s\n", pduDataFile)
	fmt.Printf("Monthly Template: %s\n", monthlyFile)
	fmt.Printf("Output: %s\n", outputFile)
	fmt.Printf("Filled PDU section: %s\n", filler.pduData.PDUName)
	fmt.Printf("Summary calculations: ✅ Current Min/AVG/Max per rack calculated\n")
	if preserveExisting {
		fmt.Printf("Mode: Preserve existing data ✅\n")
		fmt.Println("\n📝 Next steps:")
		fmt.Printf("   Run: %s total_<next_pdu>.csv\n", os.Args[0])
		fmt.Println("   This will add more PDU data while keeping existing sections intact.")
		fmt.Println("\n📊 Summary per Rack columns now filled with:")
		fmt.Println("   - Current Min: Minimum across L1/L2/L3 min values")
		fmt.Println("   - Current AVG: Average across L1/L2/L3 avg values")
		fmt.Println("   - Current Max: Maximum across L1/L2/L3 max values")
	} else {
		fmt.Printf("Mode: Clean template (previous data erased) ⚠️\n")
	}
}
