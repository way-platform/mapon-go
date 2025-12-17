package mapon

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	maponv1 "github.com/way-platform/mapon-go/proto/gen/go/wayplatform/connect/mapon/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

var update = flag.Bool("update", false, "update golden files")

func TestParseUnitsResponse_GoldenFiles(t *testing.T) {
	testdataDir := "testdata/units"

	// Check if testdata directory exists
	entries, err := os.ReadDir(testdataDir)
	if err != nil {
		if os.IsNotExist(err) {
			// Directory doesn't exist - test passes gracefully
			t.Logf("testdata/units directory does not exist, skipping golden file tests")
			return
		}
		t.Fatalf("failed to read testdata directory: %v", err)
	}

	// Find all JSON files (excluding .golden.json files)
	var jsonFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) == ".json" && !isGoldenFile(name) {
			jsonFiles = append(jsonFiles, name)
		}
	}

	// If no JSON files found, test passes gracefully
	if len(jsonFiles) == 0 {
		t.Logf("no JSON files found in testdata/units, skipping golden file tests")
		return
	}

	// Process each JSON file
	for _, jsonFile := range jsonFiles {
		t.Run(jsonFile, func(t *testing.T) {
			inputPath := filepath.Join(testdataDir, jsonFile)
			// Strip .json extension before adding .golden.json
			baseName := strings.TrimSuffix(jsonFile, ".json")
			goldenPath := filepath.Join(testdataDir, baseName+".golden.json")

			// Read input JSON file
			data, err := os.ReadFile(inputPath)
			if err != nil {
				t.Fatalf("failed to read input file %s: %v", inputPath, err)
			}

			// Parse JSON response to protobuf
			units, err := ParseUnitsResponse(data)
			if err != nil {
				t.Fatalf("failed to parse units response: %v", err)
			}

			// Generate actual output
			actual := generateGoldenOutput(units, t)

			// Update golden file if flag is set
			if *update {
				if err := os.WriteFile(goldenPath, actual, 0o644); err != nil {
					t.Fatalf("failed to write golden file %s: %v", goldenPath, err)
				}
				t.Logf("updated golden file: %s", goldenPath)
				return
			}

			// Read golden file and compare
			expected, err := os.ReadFile(goldenPath)
			if err != nil {
				if os.IsNotExist(err) {
					t.Fatalf("golden file %s does not exist. Run tests with -update flag to create it.", goldenPath)
				}
				t.Fatalf("failed to read golden file %s: %v", goldenPath, err)
			}

			// Compare actual and expected
			if !bytes.Equal(actual, expected) {
				t.Errorf("output does not match golden file %s\nRun tests with -update flag to update golden files.", goldenPath)
				t.Logf("Expected length: %d, Actual length: %d", len(expected), len(actual))
			}
		})
	}
}

func generateGoldenOutput(units []*maponv1.Unit, t *testing.T) []byte {
	// Serialize each unit to JSON using protojson.Marshal
	// Then combine into a JSON array
	unitJSONs := make([]json.RawMessage, len(units))
	for i, unit := range units {
		unitJSON, err := protojson.Marshal(unit)
		if err != nil {
			t.Fatalf("failed to marshal unit %d: %v", i, err)
		}
		unitJSONs[i] = unitJSON
	}

	// Create JSON array
	arrayJSON, err := json.Marshal(unitJSONs)
	if err != nil {
		t.Fatalf("failed to marshal units array: %v", err)
	}

	// Apply json.Indent for consistent formatting (2-space indent)
	var indentedOutput bytes.Buffer
	if err := json.Indent(&indentedOutput, arrayJSON, "", "  "); err != nil {
		t.Fatalf("failed to indent JSON: %v", err)
	}
	return indentedOutput.Bytes()
}

func isGoldenFile(filename string) bool {
	return filepath.Ext(filename) == ".json" &&
		len(filename) > len(".golden.json") &&
		filename[len(filename)-len(".golden.json"):] == ".golden.json"
}

