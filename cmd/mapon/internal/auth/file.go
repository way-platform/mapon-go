package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/adrg/xdg"
	"github.com/way-platform/mapon-go"
)

// File storing authentication credentials for the CLI.
type File struct {
	// APIKey is the API key for Mapon API.
	APIKey string `json:"apiKey,omitempty"`
}

func resolveFilepath() (string, error) {
	return xdg.ConfigFile("mapon-go/auth.json")
}

// NewClient creates a new Mapon API client using the API key from the CLI credentials.
func NewClient(ctx context.Context, opts ...mapon.ClientOption) (*mapon.Client, error) {
	cf, err := ReadFile()
	if err != nil {
		return nil, err
	}
	if cf.APIKey == "" {
		return nil, fmt.Errorf("no API key found, please login using `mapon auth login --api-key <api-key>`")
	}
	return mapon.NewClient(
		ctx,
		append(
			opts,
			mapon.WithAPIKey(cf.APIKey),
		)...,
	)
}

// ReadFile reads the currently stored [File].
func ReadFile() (*File, error) {
	fp, err := resolveFilepath()
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(fp); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no credentials found, please login using `mapon auth login`")
		}
		return nil, err
	}
	data, err := os.ReadFile(fp)
	if err != nil {
		return nil, err
	}
	var f File
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, err
	}
	return &f, nil
}

// writeFile writes the stored [File].
func writeFile(f *File) error {
	fp, err := resolveFilepath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(fp, data, 0o600)
}

// removeFile removes the stored [File].
func removeFile() error {
	fp, err := resolveFilepath()
	if err != nil {
		return err
	}
	return os.RemoveAll(fp)
}