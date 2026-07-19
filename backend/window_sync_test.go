package backend

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNormalizeWindowSyncInput(t *testing.T) {
	input, err := normalizeWindowSyncInput(WindowSyncStartInput{
		MasterProfileId:  " master ",
		TargetProfileIds: []string{" follower-1 ", "master", "", "follower-1", "follower-2"},
		Settings:         defaultWindowSyncSettings(),
	})
	if err != nil {
		t.Fatalf("normalizeWindowSyncInput returned error: %v", err)
	}
	if input.MasterProfileId != "master" {
		t.Fatalf("master = %q, want master", input.MasterProfileId)
	}
	if got := strings.Join(input.TargetProfileIds, ","); got != "follower-1,follower-2" {
		t.Fatalf("targets = %q, want follower-1,follower-2", got)
	}
}

func TestNormalizeWindowSyncInputRequiresMasterAndFollower(t *testing.T) {
	if _, err := normalizeWindowSyncInput(WindowSyncStartInput{}); err == nil {
		t.Fatal("expected missing master error")
	}
	if _, err := normalizeWindowSyncInput(WindowSyncStartInput{
		MasterProfileId:  "only",
		TargetProfileIds: []string{"only"},
	}); err == nil {
		t.Fatal("expected missing follower error")
	}
}

func TestWindowSyncNavigableURL(t *testing.T) {
	for _, value := range []string{"https://example.com", "http://127.0.0.1", "about:blank"} {
		if !isWindowSyncNavigableURL(value) {
			t.Fatalf("expected %q to be navigable", value)
		}
	}
	for _, value := range []string{"javascript:alert(1)", "file:///tmp/private", "chrome://settings", ""} {
		if isWindowSyncNavigableURL(value) {
			t.Fatalf("expected %q to be rejected", value)
		}
	}
}

func TestWindowSyncRuntimeResultParsing(t *testing.T) {
	point := json.RawMessage(`{"result":{"type":"object","value":{"x":12.5,"y":24.25}}}`)
	x, y, ok := runtimePointResult(point)
	if !ok || x != 12.5 || y != 24.25 {
		t.Fatalf("point = (%v, %v, %v)", x, y, ok)
	}
	boolean := json.RawMessage(`{"result":{"type":"boolean","value":true}}`)
	value, known := runtimeBooleanResult(boolean)
	if !known || !value {
		t.Fatalf("boolean = (%v, %v)", value, known)
	}
}

func TestWindowSyncCaptureScriptDeclaresBinding(t *testing.T) {
	if !strings.Contains(windowSyncCaptureScript, windowSyncBindingName) {
		t.Fatalf("capture script does not reference %s", windowSyncBindingName)
	}
	if !strings.Contains(windowSyncApplyInputExpression, "%s") {
		t.Fatal("input expression is missing payload placeholder")
	}
}
