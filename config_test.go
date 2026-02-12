package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Model != "" {
		t.Errorf("Model = %q, want %q", cfg.Model, "")
	}
	if cfg.LibPath != "" {
		t.Errorf("LibPath = %q, want %q", cfg.LibPath, "")
	}
	if cfg.DBPath != "rag.db" {
		t.Errorf("DBPath = %q, want %q", cfg.DBPath, "rag.db")
	}
	if cfg.ContextSize != 512 {
		t.Errorf("ContextSize = %d, want %d", cfg.ContextSize, 512)
	}
	if cfg.BatchSize != 512 {
		t.Errorf("BatchSize = %d, want %d", cfg.BatchSize, 512)
	}
	if cfg.Verbose != false {
		t.Errorf("Verbose = %v, want %v", cfg.Verbose, false)
	}
	if cfg.Server.Port != "8080" {
		t.Errorf("Server.Port = %q, want %q", cfg.Server.Port, "8080")
	}
}

func TestLoadConfig_NonExistentFile(t *testing.T) {
	cfg, err := LoadConfig("/tmp/does_not_exist_ydrag_test.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DBPath != "rag.db" {
		t.Errorf("DBPath = %q, want default %q", cfg.DBPath, "rag.db")
	}
	if cfg.ContextSize != 512 {
		t.Errorf("ContextSize = %d, want default %d", cfg.ContextSize, 512)
	}
}

func TestLoadConfig_ValidYAML(t *testing.T) {
	yamlContent := []byte(`model: test-model
lib_path: /usr/lib/test
db_path: custom.db
context_size: 1024
batch_size: 256
verbose: true
server:
  port: "9090"
`)
	tmpFile := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(tmpFile, yamlContent, 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	cfg, err := LoadConfig(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Model != "test-model" {
		t.Errorf("Model = %q, want %q", cfg.Model, "test-model")
	}
	if cfg.LibPath != "/usr/lib/test" {
		t.Errorf("LibPath = %q, want %q", cfg.LibPath, "/usr/lib/test")
	}
	if cfg.DBPath != "custom.db" {
		t.Errorf("DBPath = %q, want %q", cfg.DBPath, "custom.db")
	}
	if cfg.ContextSize != 1024 {
		t.Errorf("ContextSize = %d, want %d", cfg.ContextSize, 1024)
	}
	if cfg.BatchSize != 256 {
		t.Errorf("BatchSize = %d, want %d", cfg.BatchSize, 256)
	}
	if cfg.Verbose != true {
		t.Errorf("Verbose = %v, want %v", cfg.Verbose, true)
	}
	if cfg.Server.Port != "9090" {
		t.Errorf("Server.Port = %q, want %q", cfg.Server.Port, "9090")
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "bad.yaml")
	if err := os.WriteFile(tmpFile, []byte("{{invalid yaml:::"), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	_, err := LoadConfig(tmpFile)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestLoadConfig_EmptyPath(t *testing.T) {
	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DBPath != "rag.db" {
		t.Errorf("DBPath = %q, want default %q", cfg.DBPath, "rag.db")
	}
	if cfg.Server.Port != "8080" {
		t.Errorf("Server.Port = %q, want default %q", cfg.Server.Port, "8080")
	}
}

func TestApplyEnvOverrides(t *testing.T) {
	t.Setenv("YDRAG_MODEL", "env-model")
	t.Setenv("YZMA_LIB", "/env/lib")
	t.Setenv("YDRAG_DB_PATH", "env.db")
	t.Setenv("YDRAG_CONTEXT_SIZE", "2048")
	t.Setenv("YDRAG_BATCH_SIZE", "128")
	t.Setenv("YDRAG_VERBOSE", "true")
	t.Setenv("YDRAG_SERVER_PORT", "3000")

	cfg := DefaultConfig()
	cfg.applyEnvOverrides()

	if cfg.Model != "env-model" {
		t.Errorf("Model = %q, want %q", cfg.Model, "env-model")
	}
	if cfg.LibPath != "/env/lib" {
		t.Errorf("LibPath = %q, want %q", cfg.LibPath, "/env/lib")
	}
	if cfg.DBPath != "env.db" {
		t.Errorf("DBPath = %q, want %q", cfg.DBPath, "env.db")
	}
	if cfg.ContextSize != 2048 {
		t.Errorf("ContextSize = %d, want %d", cfg.ContextSize, 2048)
	}
	if cfg.BatchSize != 128 {
		t.Errorf("BatchSize = %d, want %d", cfg.BatchSize, 128)
	}
	if cfg.Verbose != true {
		t.Errorf("Verbose = %v, want %v", cfg.Verbose, true)
	}
	if cfg.Server.Port != "3000" {
		t.Errorf("Server.Port = %q, want %q", cfg.Server.Port, "3000")
	}
}

func TestApplyEnvOverrides_InvalidNumbers(t *testing.T) {
	t.Setenv("YDRAG_CONTEXT_SIZE", "not-a-number")
	t.Setenv("YDRAG_BATCH_SIZE", "also-bad")

	cfg := DefaultConfig()
	cfg.applyEnvOverrides()

	if cfg.ContextSize != 512 {
		t.Errorf("ContextSize = %d, want original %d (invalid env should be ignored)", cfg.ContextSize, 512)
	}
	if cfg.BatchSize != 512 {
		t.Errorf("BatchSize = %d, want original %d (invalid env should be ignored)", cfg.BatchSize, 512)
	}
}

func TestApplyEnvOverrides_VerboseVariants(t *testing.T) {
	tests := []struct {
		envVal string
		want   bool
	}{
		{"true", true},
		{"1", true},
		{"false", false},
		{"0", false},
		{"yes", false},
	}

	for _, tt := range tests {
		t.Run("YDRAG_VERBOSE="+tt.envVal, func(t *testing.T) {
			t.Setenv("YDRAG_VERBOSE", tt.envVal)
			cfg := DefaultConfig()
			cfg.applyEnvOverrides()
			if cfg.Verbose != tt.want {
				t.Errorf("Verbose = %v, want %v for env value %q", cfg.Verbose, tt.want, tt.envVal)
			}
		})
	}
}
