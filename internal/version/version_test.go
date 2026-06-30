package version_test

import (
	"testing"

	"github.com/nkootstra/cli/internal/version"
)

func TestStringUsesDevelopmentVersionByDefault(t *testing.T) {
	if got := version.String(); got != "v0.0.0-dev" {
		t.Fatalf("version.String() = %q, want %q", got, "v0.0.0-dev")
	}
}
