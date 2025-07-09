// go.mod file - create this as a separate file
/*
module pdu-processor

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
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// Statistics holds min, max, and average values
type Statistics struct {
	Min float64
	Max float64
	Avg float64
}

// DataProcessor handles the Excel file processing
type DataProcessor struct {
	pduName string
	data    map[string][]float64 // key: "Q1_l1", "Q1_l2", etc.
}

// NewDataProcessor creates a new processor instance
func NewDataProcessor() *DataProcessor {
	return &DataProcessor{
		data: make(map[string][]float64),
	}
}

// LoadInputFile loads data from input files like A1.xlsx, B1.xlsx
func (dp *DataProcessor) LoadInputFile(filename string) error {
	f, err := excelize.OpenFile(filename)
	if err != nil {
		return fmt.Errorf("failed to open input file %s: %v", filename, err)
	}
	defer f.Close()

	// Get the first sheet
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return fmt.Errorf("no sheets found in %s", filename)
	}

	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return fmt.Errorf("failed to get rows from %s: %v", filename, err)
	}

	if len(rows) < 2 {
		return fmt.Errorf("insufficient data in %s", filename)
	}

	// Parse headers to identify columns and PDU name
	headers := rows[0]
	columnMap := make(map[string]int) // key: "Q1_l1", value: column index

	for i, header := range headers {
		if i == 0 {
			continue // Skip timestamp column
		}

		header = strings.TrimSpace(header)
		if header == "" {
			continue
		}

		// Parse header like "A1 Q1 Current : l1"
		parts := strings.Split(header, " ")
		if len(parts) < 5 {
			continue
		}

		pduName := parts[0]             // "A1"
		rackName := parts[1]            // "Q1"
		lineType := parts[len(parts)-1] // "l1"

		// Set PDU name (should be consistent across all columns)
		if dp.pduName == "" {
			dp.pduName = pduName
		}

		key := fmt.Sprintf("%s_%s", rackName, lineType)
		columnMap[key] = i
	}

	// Process data rows
	for rowIndex := 1; rowIndex < len(rows); rowIndex++ {
		row := rows[rowIndex]

		for key, colIndex := range columnMap {
			if colIndex >= len(row) {
				continue
			}

			cellValue := strings.TrimSpace(row[colIndex])
			if cellValue == "" {
				continue
			}

			value, err := strconv.ParseFloat(cellValue, 64)
			if err != nil {
				continue // Skip invalid values
			}

			dp.data[key] = append(dp.data[key], value)
		}
	}

	fmt.Printf("Loaded data from %s for PDU %s: %d columns processed\n", filename, dp.pduName, len(columnMap))
	return nil
}

// CalculateStatistics calculates min, max, and average for a slice of values
func (dp *DataProcessor) CalculateStatistics(values []float64) Statistics {
	if len(values) == 0 {
		return Statistics{0, 0, 0}
	}

	min := values[0]
	max := values[0]
	sum := 0.0

	for _, val := range values {
		if val < min {
			min = val
		}
		if val > max {
			max = val
		}
		sum += val
	}

	avg := sum / float64(len(values))
	return Statistics{Min: min, Max: max, Avg: avg}
}

// GenerateOutput creates the output CSV file in the exact format expected
func (dp *DataProcessor) GenerateOutput(outputFile string) error {
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer outFile.Close()

	writer := csv.NewWriter(outFile)
	defer writer.Flush()

	// Create header row
	header := []string{"Measurement Type"}
	for i := 1; i <= 18; i++ {
		header = append(header, fmt.Sprintf("Q%d", i))
	}

	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %v", err)
	}

	// Define the order of measurement types
	measurementTypes := []string{
		"l1 min", "l1 avg", "l1 max",
		"l2 min", "l2 avg", "l2 max",
		"l3 min", "l3 avg", "l3 max",
	}

	// Process each measurement type
	for _, measurementType := range measurementTypes {
		row := []string{measurementType}

		// Extract line type and stat type
		parts := strings.Split(measurementType, " ")
		lineType := parts[0] // "l1", "l2", "l3"
		statType := parts[1] // "min", "avg", "max"

		// Process each rack Q1 to Q18
		for i := 1; i <= 18; i++ {
			rackName := fmt.Sprintf("Q%d", i)
			key := fmt.Sprintf("%s_%s", rackName, lineType)

			values, exists := dp.data[key]
			var value float64

			if exists && len(values) > 0 {
				stats := dp.CalculateStatistics(values)
				switch statType {
				case "min":
					value = stats.Min
				case "avg":
					value = stats.Avg
				case "max":
					value = stats.Max
				}
			}

			row = append(row, fmt.Sprintf("%.3f", value))
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write data row: %v", err)
		}
	}

	fmt.Printf("Output written to %s\n", outputFile)
	return nil
}

// ProcessFile is the main processing function for a single PDU file
func (dp *DataProcessor) ProcessFile(inputFile, outputFile string) error {
	// Load input file
	if err := dp.LoadInputFile(inputFile); err != nil {
		return fmt.Errorf("error loading input file: %v", err)
	}

	// Generate output
	if err := dp.GenerateOutput(outputFile); err != nil {
		return fmt.Errorf("error generating output: %v", err)
	}

	return nil
}

func main() {
	// Check command line arguments
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <input_file> [output_file]\n", os.Args[0])
		fmt.Printf("Example: %s A1.xlsx total_a1.csv\n", os.Args[0])
		fmt.Printf("Example: %s B1.xlsx total_b1.csv\n", os.Args[0])
		os.Exit(1)
	}

	inputFile := os.Args[1]

	// Generate default output filename based on input
	outputFile := "result.csv"
	if len(os.Args) > 2 {
		outputFile = os.Args[2]
	} else {
		// Auto-generate filename: A1.xlsx -> total_a1.csv
		inputName := strings.TrimSuffix(inputFile, ".xlsx")
		outputFile = fmt.Sprintf("total_%s.csv", strings.ToLower(inputName))
	}

	// Validate input file exists
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		log.Fatalf("File %s not found", inputFile)
	}

	// Create processor and run
	processor := NewDataProcessor()
	if err := processor.ProcessFile(inputFile, outputFile); err != nil {
		log.Fatalf("Processing failed: %v", err)
	}

	fmt.Printf("Processing completed successfully!\n")
	fmt.Printf("Input: %s\n", inputFile)
	fmt.Printf("Output: %s\n", outputFile)
	fmt.Printf("PDU processed: %s\n", processor.pduName)
}
