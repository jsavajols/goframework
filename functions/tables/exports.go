package tables

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"time"

	logs "github.com/jsavajols/goframework/functions/logs"
)

func ExportToJSON(data any, filepath string) (string, error) {
	logs.Logs("Exporting data to JSON format")
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	if filepath == "" {
		return string(jsonData), nil
	} else {
		// Add date and time to filename
		now := time.Now()
		timestamp := now.Format("20060102_150405")

		// Create new filename with timestamp
		filepath = filepath + timestamp + ".json"

		err = os.WriteFile(filepath, jsonData, 0644)
		if err != nil {
			return "", err
		}
		logs.Logs("JSON export completed:", filepath)
		return filepath, nil
	}
}

func ExportToCSV(data []map[string]interface{}, filepath string) (string, error) {
	logs.Logs("Exporting data to CSV format")

	if len(data) == 0 {
		return "", fmt.Errorf("no data to export")
	}

	// Add date and time to filename
	now := time.Now()
	timestamp := now.Format("20060102_150405")

	// Create new filename with timestamp
	filepath = filepath + timestamp + ".csv"

	// Create the file
	file, err := os.Create(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Get headers from first row
	var headers []string
	for key := range data[0] {
		headers = append(headers, key)
	}

	// Write headers
	if err := writer.Write(headers); err != nil {
		return "", fmt.Errorf("failed to write headers: %v", err)
	}

	// Write data rows
	for _, row := range data {
		var values []string
		for _, header := range headers {
			value := row[header]
			if value == nil {
				values = append(values, "")
			} else {
				values = append(values, fmt.Sprintf("%v", value))
			}
		}
		if err := writer.Write(values); err != nil {
			return "", fmt.Errorf("failed to write row: %v", err)
		}
	}

	logs.Logs("CSV export completed:", filepath)
	return filepath, nil
}
