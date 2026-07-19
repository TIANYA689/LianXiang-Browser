package backend

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
)

type WindowSyncActionResult struct {
	Requested int      `json:"requested"`
	Succeeded int      `json:"succeeded"`
	Failed    []string `json:"failed"`
}

type windowSyncWindowRef struct {
	ProfileID string
	DebugPort int
	Target    windowSyncPageTarget
	WindowID  int
}

func normalizeWindowSyncProfileIDs(profileIDs []string) ([]string, error) {
	seen := make(map[string]struct{}, len(profileIDs))
	normalized := make([]string, 0, len(profileIDs))
	for _, profileID := range profileIDs {
		profileID = strings.TrimSpace(profileID)
		if profileID == "" {
			continue
		}
		if _, exists := seen[profileID]; exists {
			continue
		}
		seen[profileID] = struct{}{}
		normalized = append(normalized, profileID)
	}
	if len(normalized) == 0 {
		return nil, fmt.Errorf("请至少选择一个可用窗口")
	}
	return normalized, nil
}

func (a *App) windowSyncDebugPorts(profileIDs []string) (map[string]int, error) {
	normalized, err := normalizeWindowSyncProfileIDs(profileIDs)
	if err != nil {
		return nil, err
	}
	if a == nil || a.browserMgr == nil {
		return nil, fmt.Errorf("浏览器管理器尚未初始化")
	}

	ports := make(map[string]int, len(normalized))
	a.browserMgr.Mutex.Lock()
	defer a.browserMgr.Mutex.Unlock()
	for _, profileID := range normalized {
		profile := a.browserMgr.Profiles[profileID]
		if profile == nil {
			return nil, fmt.Errorf("未找到实例：%s", profileID)
		}
		if !profile.Running || !profile.DebugReady || profile.DebugPort <= 0 {
			return nil, fmt.Errorf("实例「%s」尚未运行或调试接口未就绪", profile.ProfileName)
		}
		ports[profileID] = profile.DebugPort
	}
	return ports, nil
}

func newWindowSyncActionResult(profileIDs []string) WindowSyncActionResult {
	return WindowSyncActionResult{Requested: len(profileIDs), Failed: []string{}}
}

func (r *WindowSyncActionResult) addFailure(profileID string, err error) {
	if err != nil {
		r.Failed = append(r.Failed, fmt.Sprintf("%s: %v", profileID, err))
	}
}

func (r WindowSyncActionResult) finish() (WindowSyncActionResult, error) {
	if len(r.Failed) > 0 {
		sort.Strings(r.Failed)
	}
	return r, nil
}

func getWindowSyncWindowRef(profileID string, debugPort int) (windowSyncWindowRef, error) {
	target, err := lookupWindowSyncPageTarget(debugPort)
	if err != nil {
		return windowSyncWindowRef{}, err
	}
	result, err := cdpBrowserCallResult(debugPort, "Browser.getWindowForTarget", map[string]any{"targetId": target.Id})
	if err != nil {
		return windowSyncWindowRef{}, err
	}
	windowID, ok := result["windowId"].(float64)
	if !ok || windowID <= 0 {
		return windowSyncWindowRef{}, fmt.Errorf("浏览器未返回有效窗口 ID")
	}
	return windowSyncWindowRef{ProfileID: profileID, DebugPort: debugPort, Target: target, WindowID: int(windowID)}, nil
}

// WindowSyncWindowAction applies the same native browser-window command to every selected instance.
func (a *App) WindowSyncWindowAction(profileIDs []string, action string) (WindowSyncActionResult, error) {
	normalized, err := normalizeWindowSyncProfileIDs(profileIDs)
	if err != nil {
		return WindowSyncActionResult{}, err
	}
	ports, err := a.windowSyncDebugPorts(normalized)
	if err != nil {
		return WindowSyncActionResult{}, err
	}
	action = strings.ToLower(strings.TrimSpace(action))
	if action != "tile" && action != "cascade" && action != "maximize" && action != "minimize" && action != "normal" {
		return WindowSyncActionResult{}, fmt.Errorf("不支持的窗口操作：%s", action)
	}

	refs := make([]windowSyncWindowRef, 0, len(normalized))
	result := newWindowSyncActionResult(normalized)
	for _, profileID := range normalized {
		ref, refErr := getWindowSyncWindowRef(profileID, ports[profileID])
		if refErr != nil {
			result.addFailure(profileID, refErr)
			continue
		}
		refs = append(refs, ref)
	}

	for index, ref := range refs {
		bounds := windowSyncWindowBounds(action, index, len(refs))
		if _, callErr := cdpBrowserCallResult(ref.DebugPort, "Browser.setWindowBounds", map[string]any{"windowId": ref.WindowID, "bounds": bounds}); callErr != nil {
			result.addFailure(ref.ProfileID, callErr)
			continue
		}
		result.Succeeded++
	}
	return result.finish()
}

func windowSyncWindowBounds(action string, index int, total int) map[string]any {
	switch action {
	case "maximize", "minimize", "normal":
		return map[string]any{"windowState": action}
	case "cascade":
		step := index % 8
		return map[string]any{"left": 48 + step*36, "top": 48 + step*28, "width": 1120, "height": 760, "windowState": "normal"}
	default:
		count := maxWindowSyncInt(total, 1)
		columns := int(math.Ceil(math.Sqrt(float64(count))))
		rows := int(math.Ceil(float64(count) / float64(columns)))
		return map[string]any{"left": (index % columns) * (1600 / columns), "top": (index / columns) * (900 / rows), "width": 1600 / columns, "height": 900 / rows, "windowState": "normal"}
	}
}

func maxWindowSyncInt(value int, floor int) int {
	if value < floor {
		return floor
	}
	return value
}

func windowSyncRunPageAction(debugPort int, action func(*windowSyncCDPClient) error) error {
	client, _, err := dialWindowSyncCDP(debugPort)
	if err != nil {
		return err
	}
	defer client.Close()
	return action(client)
}

func windowSyncRunTargetAction(target windowSyncPageTarget, action func(*windowSyncCDPClient) error) error {
	if strings.TrimSpace(target.WebSocketDebuggerUrl) == "" {
		return fmt.Errorf("目标标签页调试通道不可用")
	}
	conn, err := cdpDialWebSocket(target.WebSocketDebuggerUrl)
	if err != nil {
		return fmt.Errorf("连接页面调试通道失败: %w", err)
	}
	client := &windowSyncCDPClient{
		conn:    conn,
		pending: make(map[int64]chan windowSyncCDPReply),
		events:  make(chan windowSyncCDPEvent, 256),
	}
	go client.readLoop()
	defer client.Close()
	return action(client)
}

// WindowSyncTextAction writes into the focused editable element in each selected page.
func (a *App) WindowSyncTextAction(profileIDs []string, text string, clear bool) (WindowSyncActionResult, error) {
	normalized, err := normalizeWindowSyncProfileIDs(profileIDs)
	if err != nil {
		return WindowSyncActionResult{}, err
	}
	ports, err := a.windowSyncDebugPorts(normalized)
	if err != nil {
		return WindowSyncActionResult{}, err
	}
	if !clear && text == "" {
		return WindowSyncActionResult{}, fmt.Errorf("请输入要批量写入的文本")
	}
	result := newWindowSyncActionResult(normalized)
	for _, profileID := range normalized {
		callErr := windowSyncRunPageAction(ports[profileID], func(client *windowSyncCDPClient) error {
			focusCheck := "(() => { const e = document.activeElement; return e instanceof HTMLInputElement || e instanceof HTMLTextAreaElement || Boolean(e?.isContentEditable); })()"
			raw, err := callWindowSyncCDP(client, "Runtime.evaluate", map[string]any{"expression": focusCheck, "returnByValue": true})
			if err != nil {
				return err
			}
			ok, known := runtimeBooleanResult(raw)
			if !known {
				return fmt.Errorf("无法确认当前页面的输入焦点")
			}
			if !ok {
				return fmt.Errorf("请先在该窗口聚焦输入框")
			}
			if clear {
				expression := "(() => { const e=document.activeElement; if (!(e instanceof HTMLInputElement || e instanceof HTMLTextAreaElement || e?.isContentEditable)) return false; if (e.isContentEditable) e.textContent=''; else { const p=e instanceof HTMLTextAreaElement ? HTMLTextAreaElement.prototype : HTMLInputElement.prototype; const d=Object.getOwnPropertyDescriptor(p,'value'); if (d?.set) d.set.call(e,''); else e.value=''; } e.dispatchEvent(new Event('input',{bubbles:true,composed:true})); e.dispatchEvent(new Event('change',{bubbles:true,composed:true})); return true; })()"
				raw, err := callWindowSyncCDP(client, "Runtime.evaluate", map[string]any{"expression": expression, "returnByValue": true})
				if err != nil {
					return err
				}
				ok, known := runtimeBooleanResult(raw)
				if !known {
					return fmt.Errorf("无法确认焦点内容是否已清空")
				}
				if !ok {
					return fmt.Errorf("请先在该窗口聚焦输入框")
				}
				return nil
			}
			_, err = callWindowSyncCDP(client, "Input.insertText", map[string]any{"text": text})
			return err
		})
		if callErr != nil {
			result.addFailure(profileID, callErr)
			continue
		}
		result.Succeeded++
	}
	return result.finish()
}

// WindowSyncTabAction performs a page command for every selected profile.
func (a *App) WindowSyncTabAction(profileIDs []string, action string, targetURL string) (WindowSyncActionResult, error) {
	normalized, err := normalizeWindowSyncProfileIDs(profileIDs)
	if err != nil {
		return WindowSyncActionResult{}, err
	}
	ports, err := a.windowSyncDebugPorts(normalized)
	if err != nil {
		return WindowSyncActionResult{}, err
	}
	action = strings.ToLower(strings.TrimSpace(action))
	targetURL = strings.TrimSpace(targetURL)
	if (action == "new" || action == "navigate") && !isWindowSyncNavigableURL(targetURL) {
		return WindowSyncActionResult{}, fmt.Errorf("请输入有效的 http、https 或 about:blank 地址")
	}
	if action != "new" && action != "navigate" && action != "reload" && action != "close" {
		return WindowSyncActionResult{}, fmt.Errorf("不支持的标签页操作：%s", action)
	}
	result := newWindowSyncActionResult(normalized)
	for _, profileID := range normalized {
		var callErr error
		if action == "new" {
			_, callErr = cdpBrowserCallResult(ports[profileID], "Target.createTarget", map[string]any{"url": targetURL})
		} else {
			callErr = windowSyncRunPageAction(ports[profileID], func(client *windowSyncCDPClient) error {
				method, params := "Page.navigate", map[string]any{"url": targetURL}
				if action == "reload" {
					method, params = "Page.reload", map[string]any{}
				}
				if action == "close" {
					method, params = "Page.close", map[string]any{}
				}
				_, err := callWindowSyncCDP(client, method, params)
				return err
			})
		}
		if callErr != nil {
			result.addFailure(profileID, callErr)
			continue
		}
		result.Succeeded++
	}
	return result.finish()
}

// WindowSyncCopyMasterTabs replaces each target page-tab set with the master page-tab set.
func (a *App) WindowSyncCopyMasterTabs(masterProfileID string, targetProfileIDs []string) (WindowSyncActionResult, error) {
	masterProfileID = strings.TrimSpace(masterProfileID)
	if masterProfileID == "" {
		return WindowSyncActionResult{}, fmt.Errorf("请先设置主窗口")
	}
	targets, err := normalizeWindowSyncProfileIDs(targetProfileIDs)
	if err != nil {
		return WindowSyncActionResult{}, err
	}
	ports, err := a.windowSyncDebugPorts(append([]string{masterProfileID}, targets...))
	if err != nil {
		return WindowSyncActionResult{}, err
	}
	masterTabs, err := listWindowSyncPageTargets(ports[masterProfileID])
	if err != nil {
		return WindowSyncActionResult{}, fmt.Errorf("读取主窗口标签页失败: %w", err)
	}
	urls := make([]string, 0, len(masterTabs))
	for _, tab := range masterTabs {
		if isWindowSyncNavigableURL(tab.URL) {
			urls = append(urls, tab.URL)
		}
	}
	if len(urls) == 0 {
		urls = []string{"about:blank"}
	}
	result := newWindowSyncActionResult(targets)
	for _, profileID := range targets {
		failed := false
		pages, listErr := listWindowSyncPageTargets(ports[profileID])
		if listErr != nil {
			result.addFailure(profileID, listErr)
			continue
		}
		if len(pages) == 0 {
			for _, url := range urls {
				if _, createErr := cdpBrowserCallResult(ports[profileID], "Target.createTarget", map[string]any{"url": url}); createErr != nil {
					result.addFailure(profileID, createErr)
					failed = true
					break
				}
			}
		} else {
			// Reuse one page before closing the remainder so the target browser never loses its last tab.
			if navigateErr := windowSyncRunTargetAction(pages[0], func(client *windowSyncCDPClient) error {
				_, err := callWindowSyncCDP(client, "Page.navigate", map[string]any{"url": urls[0]})
				return err
			}); navigateErr != nil {
				result.addFailure(profileID, navigateErr)
				continue
			}
			for _, url := range urls[1:] {
				if _, createErr := cdpBrowserCallResult(ports[profileID], "Target.createTarget", map[string]any{"url": url}); createErr != nil {
					result.addFailure(profileID, createErr)
					failed = true
					break
				}
			}
			if failed {
				continue
			}
			for _, page := range pages[1:] {
				if _, closeErr := cdpBrowserCallResult(ports[profileID], "Target.closeTarget", map[string]any{"targetId": page.Id}); closeErr != nil {
					result.addFailure(profileID, closeErr)
					failed = true
					break
				}
			}
		}
		if !failed {
			result.Succeeded++
		}
	}
	return result.finish()
}

func listWindowSyncPageTargets(debugPort int) ([]windowSyncPageTarget, error) {
	body, err := cdpGetEndpointBody(debugPort, "/json")
	if err != nil {
		return nil, err
	}
	var targets []windowSyncPageTarget
	if err := json.Unmarshal(body, &targets); err != nil {
		return nil, err
	}
	pages := make([]windowSyncPageTarget, 0, len(targets))
	for _, target := range targets {
		if target.Type == "page" && strings.TrimSpace(target.Id) != "" {
			pages = append(pages, target)
		}
	}
	return pages, nil
}
