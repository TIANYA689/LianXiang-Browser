package backend

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestWindowSyncControllerWithChrome(t *testing.T) {
	chromePath := os.Getenv("LIANXIANG_CHROME_PATH")
	if chromePath == "" {
		t.Skip("set LIANXIANG_CHROME_PATH to run the Chrome integration test")
	}
	if _, err := os.Stat(chromePath); err != nil {
		t.Fatalf("Chrome executable is unavailable: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<!doctype html><html><body style="height:2400px">
			<input id="message"><button id="submit" onclick="document.body.dataset.clicked='yes'">submit</button>
		</body></html>`))
	}))
	defer server.Close()

	masterPort := reserveWindowSyncTestPort(t)
	followerPort := reserveWindowSyncTestPort(t)
	startWindowSyncTestChrome(t, chromePath, masterPort, filepath.Join(t.TempDir(), "master"), server.URL)
	startWindowSyncTestChrome(t, chromePath, followerPort, filepath.Join(t.TempDir(), "follower"), server.URL)

	controller := newWindowSyncController(WindowSyncStartInput{
		MasterProfileId:  "master",
		TargetProfileIds: []string{"follower"},
		Settings:         defaultWindowSyncSettings(),
	}, map[string]int{"master": masterPort, "follower": followerPort})
	if err := controller.Start(); err != nil {
		t.Fatalf("controller.Start failed: %v", err)
	}
	t.Cleanup(func() { controller.Stop("") })

	_, err := callWindowSyncCDP(controller.master, "Runtime.evaluate", map[string]any{
		"expression": `(() => {
			const input = document.querySelector('#message');
			input.value = 'hello-sync';
			input.dispatchEvent(new Event('input', {bubbles:true}));
			document.querySelector('#submit').click();
			window.scrollTo(0, 1200);
			return true;
		})()`,
	})
	if err != nil {
		t.Fatalf("master event injection failed: %v", err)
	}

	deadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) {
		result, queryErr := callWindowSyncCDP(controller.followers[0].Client, "Runtime.evaluate", map[string]any{
			"expression":    `JSON.stringify({value:document.querySelector('#message').value,clicked:document.body.dataset.clicked||'',scrollY:window.scrollY})`,
			"returnByValue": true,
		})
		if queryErr == nil {
			var runtimeResult struct {
				Result struct {
					Value string `json:"value"`
				} `json:"result"`
			}
			if json.Unmarshal(result, &runtimeResult) == nil {
				var pageState struct {
					Value   string  `json:"value"`
					Clicked string  `json:"clicked"`
					ScrollY float64 `json:"scrollY"`
				}
				if json.Unmarshal([]byte(runtimeResult.Result.Value), &pageState) == nil &&
					pageState.Value == "hello-sync" && pageState.Clicked == "yes" && pageState.ScrollY > 0 {
					if controller.Snapshot().EventCount < 3 {
						t.Fatalf("eventCount = %d, want at least 3", controller.Snapshot().EventCount)
					}
					return
				}
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("follower did not converge; final state: %+v", controller.Snapshot())
}

func reserveWindowSyncTestPort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	_ = listener.Close()
	return port
}

func startWindowSyncTestChrome(t *testing.T, chromePath string, debugPort int, userDataDir string, targetURL string) {
	t.Helper()
	cmd := exec.Command(chromePath,
		"--headless=new",
		"--disable-gpu",
		"--disable-background-networking",
		"--no-first-run",
		"--no-default-browser-check",
		"--remote-debugging-address=127.0.0.1",
		fmt.Sprintf("--remote-debugging-port=%d", debugPort),
		fmt.Sprintf("--user-data-dir=%s", userDataDir),
		targetURL,
	)
	hideWindow(cmd)
	if err := cmd.Start(); err != nil {
		t.Fatalf("start Chrome on %d: %v", debugPort, err)
	}
	t.Cleanup(func() {
		_ = cdpBrowserCall(debugPort, "Browser.close", map[string]any{})
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
	})

	deadline := time.Now().Add(12 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := lookupWindowSyncPageTarget(debugPort); err == nil {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("Chrome debug port %d did not become ready", debugPort)
}
