package mapon

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/encoding/protojson"
)

func TestParsePushMessage_GoldenFiles(t *testing.T) {
	t.Parallel()
	testdataDir := "testdata/push_messages"
	entries, err := os.ReadDir(testdataDir)
	if err != nil {
		if os.IsNotExist(err) {
			t.Logf("testdata/push_messages directory does not exist, skipping")
			return
		}
		t.Fatalf("read testdata directory: %v", err)
	}
	var jsonFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) == ".json" && !strings.HasSuffix(name, ".golden.json") {
			jsonFiles = append(jsonFiles, name)
		}
	}
	slices.Sort(jsonFiles)
	if len(jsonFiles) == 0 {
		t.Fatal("no .json files found under testdata/push_messages")
	}
	for _, jsonFile := range jsonFiles {
		t.Run(jsonFile, func(t *testing.T) {
			t.Parallel()
			inputPath := filepath.Join(testdataDir, jsonFile)
			goldenPath := filepath.Join(testdataDir, strings.TrimSuffix(jsonFile, ".json")+".golden.json")
			inputData, err := os.ReadFile(inputPath)
			if err != nil {
				t.Fatalf("read input %s: %v", inputPath, err)
			}
			msg, err := ParsePushMessage(inputData)
			if err != nil {
				t.Fatalf("ParsePushMessage(%s): %v", inputPath, err)
			}
			unstableJSON, err := protojson.Marshal(msg)
			if err != nil {
				t.Fatalf("protojson.Marshal: %v", err)
			}
			var stableJSON bytes.Buffer
			if err := json.Indent(&stableJSON, unstableJSON, "", "  "); err != nil {
				t.Fatalf("json.Indent: %v", err)
			}
			stableJSON.WriteByte('\n')
			if *update {
				if err := os.WriteFile(goldenPath, stableJSON.Bytes(), 0o644); err != nil {
					t.Fatalf("write golden %s: %v", goldenPath, err)
				}
				return
			}
			goldenData, err := os.ReadFile(goldenPath)
			if err != nil {
				if os.IsNotExist(err) {
					t.Fatalf("golden file %s does not exist; run with -update", goldenPath)
				}
				t.Fatalf("read golden %s: %v", goldenPath, err)
			}
			if !bytes.Equal(goldenData, stableJSON.Bytes()) {
				t.Fatalf("golden mismatch for %s; run: go test -args -update", jsonFile)
			}
		})
	}
}

func TestParsePushMessage_Errors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
	}{
		{name: "invalid json", input: `not json`},
		{name: "empty gmt", input: `{"car_id":1,"pack_id":1,"gmt":""}`},
		{name: "bad gmt format", input: `{"car_id":1,"pack_id":1,"gmt":"2025/01/01"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := ParsePushMessage([]byte(tt.input))
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}
