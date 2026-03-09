package ociinstaller

import (
	"encoding/json"
	"testing"
)

func TestDbImageConfigUnmarshal(t *testing.T) {
	raw := []byte(`{
		"schemaVersion": "2020-11-18",
		"db": {
			"name": "db",
			"organization": "turbot",
			"version": "14.21.0",
			"dbVersion": "14.21.0"
		}
	}`)

	var cfg dbImageConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		t.Fatalf("unmarshal db image config: %v", err)
	}

	if cfg.Database.Version != "14.21.0" {
		t.Fatalf("unexpected db image version %q", cfg.Database.Version)
	}
	if cfg.Database.DBVersion != "14.21.0" {
		t.Fatalf("unexpected postgres version %q", cfg.Database.DBVersion)
	}
}
