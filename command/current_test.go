package command

import (
	"path/filepath"
	"testing"

	"github.com/Jabba-Team/jabba/cfg"
)

func TestCurrent(t *testing.T) {
	var prevLookPath = lookPath
	defer func() { lookPath = prevLookPath }()
	lookPath = func(file string) (string, error) {
		return filepath.Join(cfg.Dir(), "jdk", "1.8.0", "Contents", "Home", "bin", "java"), nil
	}
	actual := Current()
	expected := "1.8.0"
	if actual != expected {
		t.Fatalf("actual: %v != expected: %v", actual, expected)
	}
}
