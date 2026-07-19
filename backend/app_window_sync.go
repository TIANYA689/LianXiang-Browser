package backend

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// WindowSyncSettings controls which page-level operations are mirrored.
// Browser chrome (address bar, extension popups and native dialogs) is outside
// CDP's page target and is intentionally not included here.
type WindowSyncSettings struct {
	SyncClicks     bool `json:"syncClicks"`
	SyncInputs     bool `json:"syncInputs"`
	SyncScroll     bool `json:"syncScroll"`
	SyncNavigation bool `json:"syncNavigation"`
}

type WindowSyncStartInput struct {
	MasterProfileId  string             `json:"masterProfileId"`
	TargetProfileIds []string           `json:"targetProfileIds"`
	Settings         WindowSyncSettings `json:"settings"`
}

type WindowSyncSession struct {
	ProfileId   string   `json:"profileId"`
	ProfileName string   `json:"profileName"`
	Tags        []string `json:"tags"`
	Running     bool     `json:"running"`
	DebugReady  bool     `json:"debugReady"`
	DebugPort   int      `json:"debugPort"`
	PageTitle   string   `json:"pageTitle"`
	PageURL     string   `json:"pageUrl"`
	Available   bool     `json:"available"`
	Warning     string   `json:"warning"`
}

type WindowSyncState struct {
	Active           bool               `json:"active"`
	MasterProfileId  string             `json:"masterProfileId"`
	TargetProfileIds []string           `json:"targetProfileIds"`
	Settings         WindowSyncSettings `json:"settings"`
	StartedAt        string             `json:"startedAt"`
	LastEventAt      string             `json:"lastEventAt"`
	LastEventType    string             `json:"lastEventType"`
	EventCount       int64              `json:"eventCount"`
	LastError        string             `json:"lastError"`
}

func defaultWindowSyncSettings() WindowSyncSettings {
	return WindowSyncSettings{
		SyncClicks:     true,
		SyncInputs:     true,
		SyncScroll:     true,
		SyncNavigation: true,
	}
}

func normalizeWindowSyncInput(input WindowSyncStartInput) (WindowSyncStartInput, error) {
	input.MasterProfileId = strings.TrimSpace(input.MasterProfileId)
	if input.MasterProfileId == "" {
		return input, fmt.Errorf("请先设置主窗口")
	}

	seen := map[string]struct{}{input.MasterProfileId: {}}
	targets := make([]string, 0, len(input.TargetProfileIds))
	for _, profileId := range input.TargetProfileIds {
		profileId = strings.TrimSpace(profileId)
		if profileId == "" {
			continue
		}
		if _, exists := seen[profileId]; exists {
			continue
		}
		seen[profileId] = struct{}{}
		targets = append(targets, profileId)
	}
	if len(targets) == 0 {
		return input, fmt.Errorf("请至少选择一个被控窗口")
	}
	input.TargetProfileIds = targets
	return input, nil
}

func (a *App) WindowSyncListSessions() []WindowSyncSession {
	if a == nil || a.browserMgr == nil {
		return []WindowSyncSession{}
	}

	a.browserMgr.Mutex.Lock()
	sessions := make([]WindowSyncSession, 0, len(a.browserMgr.Profiles))
	for _, profile := range a.browserMgr.Profiles {
		if profile == nil || strings.TrimSpace(profile.DeletedAt) != "" {
			continue
		}
		sessions = append(sessions, WindowSyncSession{
			ProfileId:   profile.ProfileId,
			ProfileName: profile.ProfileName,
			Tags:        append([]string{}, profile.Tags...),
			Running:     profile.Running,
			DebugReady:  profile.DebugReady,
			DebugPort:   profile.DebugPort,
		})
	}
	a.browserMgr.Mutex.Unlock()

	var lookupWait sync.WaitGroup
	lookupSlots := make(chan struct{}, 8)
	for index := range sessions {
		index := index
		session := &sessions[index]
		if !session.Running {
			session.Warning = "实例未启动"
			continue
		}
		if !session.DebugReady || session.DebugPort <= 0 {
			session.Warning = "调试接口尚未就绪"
			continue
		}
		lookupWait.Add(1)
		go func() {
			defer lookupWait.Done()
			lookupSlots <- struct{}{}
			defer func() { <-lookupSlots }()
			target, err := lookupWindowSyncPageTarget(sessions[index].DebugPort)
			if err != nil {
				sessions[index].Warning = err.Error()
				return
			}
			sessions[index].PageTitle = target.Title
			sessions[index].PageURL = target.URL
			sessions[index].Available = true
		}()
	}
	lookupWait.Wait()

	sort.SliceStable(sessions, func(i, j int) bool {
		if sessions[i].Available != sessions[j].Available {
			return sessions[i].Available
		}
		return strings.ToLower(sessions[i].ProfileName) < strings.ToLower(sessions[j].ProfileName)
	})
	return sessions
}

func (a *App) WindowSyncGetState() WindowSyncState {
	if a == nil {
		return WindowSyncState{Settings: defaultWindowSyncSettings()}
	}
	a.windowSyncMu.Lock()
	controller := a.windowSync
	a.windowSyncMu.Unlock()
	if controller == nil {
		return WindowSyncState{Settings: defaultWindowSyncSettings()}
	}
	return controller.Snapshot()
}

func (a *App) WindowSyncStart(input WindowSyncStartInput) (WindowSyncState, error) {
	if a == nil || a.browserMgr == nil {
		return WindowSyncState{}, fmt.Errorf("浏览器管理器尚未初始化")
	}
	normalized, err := normalizeWindowSyncInput(input)
	if err != nil {
		return WindowSyncState{}, err
	}

	ports := make(map[string]int, len(normalized.TargetProfileIds)+1)
	profileIds := append([]string{normalized.MasterProfileId}, normalized.TargetProfileIds...)
	a.browserMgr.Mutex.Lock()
	for _, profileId := range profileIds {
		profile := a.browserMgr.Profiles[profileId]
		if profile == nil {
			a.browserMgr.Mutex.Unlock()
			return WindowSyncState{}, fmt.Errorf("未找到实例：%s", profileId)
		}
		if !profile.Running || !profile.DebugReady || profile.DebugPort <= 0 {
			a.browserMgr.Mutex.Unlock()
			return WindowSyncState{}, fmt.Errorf("实例「%s」尚未运行或调试接口未就绪", profile.ProfileName)
		}
		ports[profileId] = profile.DebugPort
	}
	a.browserMgr.Mutex.Unlock()

	a.windowSyncMu.Lock()
	previous := a.windowSync
	a.windowSync = nil
	a.windowSyncMu.Unlock()
	if previous != nil {
		previous.Stop("")
	}

	controller := newWindowSyncController(normalized, ports)
	if err := controller.Start(); err != nil {
		return controller.Snapshot(), err
	}

	a.windowSyncMu.Lock()
	a.windowSync = controller
	a.windowSyncMu.Unlock()
	a.emitWindowSyncState(controller.Snapshot())
	return controller.Snapshot(), nil
}

func (a *App) WindowSyncStop() WindowSyncState {
	if a == nil {
		return WindowSyncState{Settings: defaultWindowSyncSettings()}
	}
	a.windowSyncMu.Lock()
	controller := a.windowSync
	a.windowSyncMu.Unlock()
	if controller == nil {
		return WindowSyncState{Settings: defaultWindowSyncSettings()}
	}
	controller.Stop("")
	state := controller.Snapshot()
	a.emitWindowSyncState(state)
	return state
}

func (a *App) WindowSyncShowWindow(profileId string) error {
	debugPort, err := a.getDebugPort(strings.TrimSpace(profileId))
	if err != nil {
		return err
	}
	target, err := lookupWindowSyncPageTarget(debugPort)
	if err != nil {
		return err
	}
	result, err := cdpBrowserCallResult(debugPort, "Browser.getWindowForTarget", map[string]any{
		"targetId": target.Id,
	})
	if err != nil {
		return fmt.Errorf("获取浏览器窗口失败: %w", err)
	}
	windowId, ok := result["windowId"].(float64)
	if !ok || windowId <= 0 {
		data, _ := json.Marshal(result["windowId"])
		return fmt.Errorf("浏览器未返回有效窗口 ID：%s", string(data))
	}
	_, err = cdpBrowserCallResult(debugPort, "Browser.setWindowBounds", map[string]any{
		"windowId": int(windowId),
		"bounds": map[string]any{
			"windowState": "normal",
		},
	})
	if err != nil {
		return fmt.Errorf("显示浏览器窗口失败: %w", err)
	}
	return nil
}

func (a *App) stopWindowSyncRuntime() {
	if a == nil {
		return
	}
	a.windowSyncMu.Lock()
	controller := a.windowSync
	a.windowSyncMu.Unlock()
	if controller != nil {
		controller.Stop("")
	}
}

func (a *App) emitWindowSyncState(state WindowSyncState) {
	if a == nil || a.ctx == nil {
		return
	}
	runtime.EventsEmit(a.ctx, "window-sync:state", state)
}

func windowSyncNow() string {
	return time.Now().Format(time.RFC3339)
}
