//go:build windows

package backend

import (
	"encoding/json"
	"testing"
)

func TestRunPowerShellJSONBindsNamedArgument(t *testing.T) {
	const want = `C:\Chrome Data\User's Profile`
	output, err := runPowerShellJSON(`param([string]$Value)
[pscustomobject]@{ value = $Value } | ConvertTo-Json -Compress
`, "-Value", want)
	if err != nil {
		t.Fatalf("runPowerShellJSON returned error: %v", err)
	}

	var result struct {
		Value string `json:"value"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("decode PowerShell output %q: %v", output, err)
	}
	if result.Value != want {
		t.Fatalf("PowerShell argument = %q, want %q", result.Value, want)
	}
}
