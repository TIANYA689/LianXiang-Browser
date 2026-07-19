package backend

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

const windowSyncBindingName = "__lianxiangWindowSyncEmit"

type windowSyncPageTarget struct {
	Id                   string `json:"id"`
	Title                string `json:"title"`
	Type                 string `json:"type"`
	URL                  string `json:"url"`
	WebSocketDebuggerUrl string `json:"webSocketDebuggerUrl"`
}

type windowSyncEvent struct {
	Type       string  `json:"type"`
	Selector   string  `json:"selector,omitempty"`
	Value      string  `json:"value,omitempty"`
	Checked    bool    `json:"checked,omitempty"`
	InputKind  string  `json:"inputKind,omitempty"`
	Key        string  `json:"key,omitempty"`
	Code       string  `json:"code,omitempty"`
	Button     string  `json:"button,omitempty"`
	ClickCount int     `json:"clickCount,omitempty"`
	XRatio     float64 `json:"xRatio,omitempty"`
	YRatio     float64 `json:"yRatio,omitempty"`
	ScrollX    float64 `json:"scrollX,omitempty"`
	ScrollY    float64 `json:"scrollY,omitempty"`
	URL        string  `json:"url,omitempty"`
}

type windowSyncCDPEvent struct {
	Method string
	Params json.RawMessage
}

type windowSyncCDPReply struct {
	Result json.RawMessage
	Err    error
}

type windowSyncCDPEnvelope struct {
	Id     int64           `json:"id,omitempty"`
	Method string          `json:"method,omitempty"`
	Params json.RawMessage `json:"params,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type windowSyncCDPClient struct {
	conn      *websocket.Conn
	writeMu   sync.Mutex
	pendingMu sync.Mutex
	pending   map[int64]chan windowSyncCDPReply
	nextId    atomic.Int64
	events    chan windowSyncCDPEvent
	closeOnce sync.Once
}

type windowSyncFollower struct {
	ProfileId string
	Client    *windowSyncCDPClient
}

type windowSyncController struct {
	mu        sync.RWMutex
	state     WindowSyncState
	ports     map[string]int
	master    *windowSyncCDPClient
	followers []windowSyncFollower
	stopOnce  sync.Once
}

func lookupWindowSyncPageTarget(debugPort int) (windowSyncPageTarget, error) {
	body, err := cdpGetEndpointBody(debugPort, "/json")
	if err != nil {
		return windowSyncPageTarget{}, fmt.Errorf("无法连接调试接口: %w", err)
	}
	var targets []windowSyncPageTarget
	if err := json.Unmarshal(body, &targets); err != nil {
		return windowSyncPageTarget{}, fmt.Errorf("解析页面目标失败: %w", err)
	}
	for _, target := range targets {
		if target.Type == "page" && strings.TrimSpace(target.WebSocketDebuggerUrl) != "" {
			return target, nil
		}
	}
	return windowSyncPageTarget{}, fmt.Errorf("当前实例没有可同步的页面")
}

func dialWindowSyncCDP(debugPort int) (*windowSyncCDPClient, windowSyncPageTarget, error) {
	target, err := lookupWindowSyncPageTarget(debugPort)
	if err != nil {
		return nil, windowSyncPageTarget{}, err
	}
	conn, err := cdpDialWebSocket(target.WebSocketDebuggerUrl)
	if err != nil {
		return nil, windowSyncPageTarget{}, fmt.Errorf("连接页面调试通道失败: %w", err)
	}
	client := &windowSyncCDPClient{
		conn:    conn,
		pending: make(map[int64]chan windowSyncCDPReply),
		events:  make(chan windowSyncCDPEvent, 256),
	}
	go client.readLoop()
	return client, target, nil
}

func (c *windowSyncCDPClient) readLoop() {
	var readErr error
	defer func() {
		if readErr == nil {
			readErr = errors.New("CDP 连接已关闭")
		}
		c.pendingMu.Lock()
		pending := c.pending
		c.pending = make(map[int64]chan windowSyncCDPReply)
		c.pendingMu.Unlock()
		for _, ch := range pending {
			ch <- windowSyncCDPReply{Err: readErr}
			close(ch)
		}
		close(c.events)
	}()

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			readErr = err
			return
		}
		var message windowSyncCDPEnvelope
		if err := json.Unmarshal(data, &message); err != nil {
			continue
		}
		if message.Id > 0 {
			c.pendingMu.Lock()
			ch := c.pending[message.Id]
			delete(c.pending, message.Id)
			c.pendingMu.Unlock()
			if ch == nil {
				continue
			}
			if message.Error != nil {
				ch <- windowSyncCDPReply{Err: fmt.Errorf("CDP %d: %s", message.Error.Code, message.Error.Message)}
			} else {
				ch <- windowSyncCDPReply{Result: message.Result}
			}
			close(ch)
			continue
		}
		if message.Method != "" {
			select {
			case c.events <- windowSyncCDPEvent{Method: message.Method, Params: message.Params}:
			default:
				// Input and scroll bursts are allowed to coalesce under pressure.
			}
		}
	}
}

func (c *windowSyncCDPClient) Call(ctx context.Context, method string, params map[string]any) (json.RawMessage, error) {
	if c == nil || c.conn == nil {
		return nil, errors.New("CDP 客户端未连接")
	}
	id := c.nextId.Add(1)
	response := make(chan windowSyncCDPReply, 1)
	c.pendingMu.Lock()
	c.pending[id] = response
	c.pendingMu.Unlock()

	c.writeMu.Lock()
	err := c.conn.WriteJSON(map[string]any{"id": id, "method": method, "params": params})
	c.writeMu.Unlock()
	if err != nil {
		c.pendingMu.Lock()
		delete(c.pending, id)
		c.pendingMu.Unlock()
		return nil, fmt.Errorf("发送 %s 失败: %w", method, err)
	}

	select {
	case reply, ok := <-response:
		if !ok {
			return nil, errors.New("CDP 连接已关闭")
		}
		return reply.Result, reply.Err
	case <-ctx.Done():
		c.pendingMu.Lock()
		delete(c.pending, id)
		c.pendingMu.Unlock()
		return nil, fmt.Errorf("%s 超时: %w", method, ctx.Err())
	}
}

func (c *windowSyncCDPClient) Close() {
	if c == nil {
		return
	}
	c.closeOnce.Do(func() {
		_ = c.conn.Close()
	})
}

func (c *windowSyncCDPClient) Events() <-chan windowSyncCDPEvent {
	return c.events
}

func newWindowSyncController(input WindowSyncStartInput, ports map[string]int) *windowSyncController {
	portSnapshot := make(map[string]int, len(ports))
	for profileId, port := range ports {
		portSnapshot[profileId] = port
	}
	return &windowSyncController{
		state: WindowSyncState{
			MasterProfileId:  input.MasterProfileId,
			TargetProfileIds: append([]string{}, input.TargetProfileIds...),
			Settings:         input.Settings,
		},
		ports: portSnapshot,
	}
}

func (c *windowSyncController) Start() error {
	masterPort := c.portFor(c.state.MasterProfileId)
	master, masterTarget, err := dialWindowSyncCDP(masterPort)
	if err != nil {
		c.failStart(fmt.Errorf("主窗口连接失败: %w", err))
		return fmt.Errorf("启动窗口同步失败: %w", err)
	}
	c.master = master

	for _, profileId := range c.state.TargetProfileIds {
		client, _, err := dialWindowSyncCDP(c.portFor(profileId))
		if err != nil {
			c.failStart(fmt.Errorf("被控窗口 %s 连接失败: %w", profileId, err))
			return fmt.Errorf("启动窗口同步失败: 被控窗口连接失败（%s）：%w", profileId, err)
		}
		c.followers = append(c.followers, windowSyncFollower{ProfileId: profileId, Client: client})
	}

	if err := c.prepareMaster(); err != nil {
		c.failStart(err)
		return fmt.Errorf("启动窗口同步失败: %w", err)
	}
	if c.state.Settings.SyncNavigation && isWindowSyncNavigableURL(masterTarget.URL) {
		_ = c.forEachFollower(func(_ string, client *windowSyncCDPClient) error {
			_, err := callWindowSyncCDP(client, "Page.navigate", map[string]any{"url": masterTarget.URL})
			return err
		})
	}

	c.mu.Lock()
	c.state.Active = true
	c.state.StartedAt = windowSyncNow()
	c.state.LastError = ""
	c.mu.Unlock()
	go c.eventLoop()
	return nil
}

func (c *windowSyncController) portFor(profileId string) int {
	return c.ports[profileId]
}

func (c *windowSyncController) prepareMaster() error {
	for _, command := range []struct {
		method string
		params map[string]any
	}{
		{method: "Page.enable", params: map[string]any{}},
		{method: "Runtime.enable", params: map[string]any{}},
		{method: "Runtime.addBinding", params: map[string]any{"name": windowSyncBindingName}},
		{method: "Page.addScriptToEvaluateOnNewDocument", params: map[string]any{"source": windowSyncCaptureScript}},
		{method: "Runtime.evaluate", params: map[string]any{"expression": windowSyncCaptureScript}},
	} {
		if _, err := callWindowSyncCDP(c.master, command.method, command.params); err != nil {
			return fmt.Errorf("初始化主窗口监听失败（%s）: %w", command.method, err)
		}
	}
	return nil
}

func (c *windowSyncController) eventLoop() {
	for event := range c.master.Events() {
		switch event.Method {
		case "Runtime.bindingCalled":
			var params struct {
				Name    string `json:"name"`
				Payload string `json:"payload"`
			}
			if json.Unmarshal(event.Params, &params) != nil || params.Name != windowSyncBindingName {
				continue
			}
			var input windowSyncEvent
			if json.Unmarshal([]byte(params.Payload), &input) != nil {
				continue
			}
			c.dispatch(input)
		case "Page.frameNavigated":
			if !c.state.Settings.SyncNavigation {
				continue
			}
			var params struct {
				Frame struct {
					ParentId string `json:"parentId"`
					URL      string `json:"url"`
				} `json:"frame"`
			}
			if json.Unmarshal(event.Params, &params) != nil || params.Frame.ParentId != "" || !isWindowSyncNavigableURL(params.Frame.URL) {
				continue
			}
			c.dispatch(windowSyncEvent{Type: "navigation", URL: params.Frame.URL})
		}
	}
	c.Stop("主窗口调试连接已断开，同步已自动停止")
}

func (c *windowSyncController) dispatch(event windowSyncEvent) {
	if !c.shouldDispatch(event.Type) {
		return
	}
	err := c.forEachFollower(func(_ string, client *windowSyncCDPClient) error {
		switch event.Type {
		case "click":
			return dispatchWindowSyncClick(client, event)
		case "input":
			return dispatchWindowSyncInput(client, event)
		case "scroll":
			return dispatchWindowSyncScroll(client, event)
		case "key":
			return dispatchWindowSyncKey(client, event)
		case "navigation":
			_, err := callWindowSyncCDP(client, "Page.navigate", map[string]any{"url": event.URL})
			return err
		default:
			return nil
		}
	})

	c.mu.Lock()
	c.state.EventCount++
	c.state.LastEventAt = windowSyncNow()
	c.state.LastEventType = event.Type
	if err != nil {
		c.state.LastError = err.Error()
	} else {
		c.state.LastError = ""
	}
	c.mu.Unlock()
}

func (c *windowSyncController) shouldDispatch(eventType string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if !c.state.Active {
		return false
	}
	switch eventType {
	case "click", "key":
		return c.state.Settings.SyncClicks
	case "input":
		return c.state.Settings.SyncInputs
	case "scroll":
		return c.state.Settings.SyncScroll
	case "navigation":
		return c.state.Settings.SyncNavigation
	default:
		return false
	}
}

func (c *windowSyncController) forEachFollower(action func(string, *windowSyncCDPClient) error) error {
	type followerError struct {
		profileId string
		err       error
	}
	errorsCh := make(chan followerError, len(c.followers))
	var wg sync.WaitGroup
	for _, follower := range c.followers {
		follower := follower
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := action(follower.ProfileId, follower.Client); err != nil {
				errorsCh <- followerError{profileId: follower.ProfileId, err: err}
			}
		}()
	}
	wg.Wait()
	close(errorsCh)
	items := make([]string, 0)
	for item := range errorsCh {
		items = append(items, fmt.Sprintf("%s: %v", item.profileId, item.err))
	}
	sort.Strings(items)
	if len(items) > 0 {
		return fmt.Errorf("部分被控窗口同步失败：%s", strings.Join(items, "；"))
	}
	return nil
}

func (c *windowSyncController) Snapshot() WindowSyncState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	state := c.state
	state.TargetProfileIds = append([]string{}, c.state.TargetProfileIds...)
	return state
}

func (c *windowSyncController) Stop(reason string) {
	c.stopOnce.Do(func() {
		if c.master != nil {
			c.master.Close()
		}
		for _, follower := range c.followers {
			follower.Client.Close()
		}
		c.mu.Lock()
		c.state.Active = false
		if strings.TrimSpace(reason) != "" {
			c.state.LastError = strings.TrimSpace(reason)
		}
		c.mu.Unlock()
	})
}

func (c *windowSyncController) failStart(err error) {
	c.mu.Lock()
	c.state.Active = false
	c.state.LastError = err.Error()
	c.mu.Unlock()
	c.Stop(err.Error())
}

func callWindowSyncCDP(client *windowSyncCDPClient, method string, params map[string]any) (json.RawMessage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	return client.Call(ctx, method, params)
}

func dispatchWindowSyncClick(client *windowSyncCDPClient, event windowSyncEvent) error {
	x, y, err := resolveWindowSyncPoint(client, event)
	if err != nil {
		return err
	}
	button := event.Button
	if button != "left" && button != "middle" && button != "right" {
		button = "left"
	}
	clickCount := event.ClickCount
	if clickCount < 1 {
		clickCount = 1
	}
	for _, params := range []map[string]any{
		{"type": "mouseMoved", "x": x, "y": y, "button": "none"},
		{"type": "mousePressed", "x": x, "y": y, "button": button, "clickCount": clickCount},
		{"type": "mouseReleased", "x": x, "y": y, "button": button, "clickCount": clickCount},
	} {
		if _, err := callWindowSyncCDP(client, "Input.dispatchMouseEvent", params); err != nil {
			return err
		}
	}
	return nil
}

func dispatchWindowSyncInput(client *windowSyncCDPClient, event windowSyncEvent) error {
	payload, _ := json.Marshal(event)
	expression := fmt.Sprintf(windowSyncApplyInputExpression, string(payload))
	result, err := callWindowSyncCDP(client, "Runtime.evaluate", map[string]any{
		"expression":    expression,
		"returnByValue": true,
	})
	if err != nil {
		return err
	}
	if ok, known := runtimeBooleanResult(result); known && !ok {
		return errors.New("被控页面未找到对应输入控件")
	}
	return nil
}

func dispatchWindowSyncScroll(client *windowSyncCDPClient, event windowSyncEvent) error {
	payload, _ := json.Marshal(map[string]any{
		"selector": event.Selector,
		"scrollX":  clampRatio(event.ScrollX),
		"scrollY":  clampRatio(event.ScrollY),
	})
	expression := fmt.Sprintf(`(() => {
		const payload = %s;
		const element = payload.selector ? document.querySelector(payload.selector) : null;
		if (payload.selector && !element) return false;
		if (element) {
			element.scrollTo({left: Math.max(0, (element.scrollWidth-element.clientWidth)*payload.scrollX), top: Math.max(0, (element.scrollHeight-element.clientHeight)*payload.scrollY), behavior:'instant'});
		} else {
			window.scrollTo({left: Math.max(0, (document.documentElement.scrollWidth-innerWidth)*payload.scrollX), top: Math.max(0, (document.documentElement.scrollHeight-innerHeight)*payload.scrollY), behavior:'instant'});
		}
		return true;
	})()`, string(payload))
	_, err := callWindowSyncCDP(client, "Runtime.evaluate", map[string]any{"expression": expression})
	return err
}

func dispatchWindowSyncKey(client *windowSyncCDPClient, event windowSyncEvent) error {
	if event.Selector != "" {
		selectorJSON, _ := json.Marshal(event.Selector)
		_, _ = callWindowSyncCDP(client, "Runtime.evaluate", map[string]any{
			"expression": fmt.Sprintf("document.querySelector(%s)?.focus(); true", string(selectorJSON)),
		})
	}
	key := strings.TrimSpace(event.Key)
	if key == "" {
		return nil
	}
	for _, eventType := range []string{"keyDown", "keyUp"} {
		if _, err := callWindowSyncCDP(client, "Input.dispatchKeyEvent", map[string]any{
			"type": eventType,
			"key":  key,
			"code": event.Code,
		}); err != nil {
			return err
		}
	}
	return nil
}

func resolveWindowSyncPoint(client *windowSyncCDPClient, event windowSyncEvent) (float64, float64, error) {
	if strings.TrimSpace(event.Selector) != "" {
		selectorJSON, _ := json.Marshal(event.Selector)
		expression := fmt.Sprintf("(() => { const el=document.querySelector(%s); if(!el) return null; const r=el.getBoundingClientRect(); return {x:r.left+r.width/2,y:r.top+r.height/2}; })()", string(selectorJSON))
		result, err := callWindowSyncCDP(client, "Runtime.evaluate", map[string]any{"expression": expression, "returnByValue": true})
		if err == nil {
			if x, y, ok := runtimePointResult(result); ok {
				return x, y, nil
			}
		}
	}

	result, err := callWindowSyncCDP(client, "Page.getLayoutMetrics", map[string]any{})
	if err != nil {
		return 0, 0, err
	}
	var metrics struct {
		CSSVisualViewport struct {
			ClientWidth  float64 `json:"clientWidth"`
			ClientHeight float64 `json:"clientHeight"`
		} `json:"cssVisualViewport"`
	}
	if json.Unmarshal(result, &metrics) != nil || metrics.CSSVisualViewport.ClientWidth <= 0 || metrics.CSSVisualViewport.ClientHeight <= 0 {
		return 0, 0, errors.New("无法读取被控窗口页面尺寸")
	}
	return clampRatio(event.XRatio) * metrics.CSSVisualViewport.ClientWidth,
		clampRatio(event.YRatio) * metrics.CSSVisualViewport.ClientHeight, nil
}

func runtimePointResult(raw json.RawMessage) (float64, float64, bool) {
	var payload struct {
		Result struct {
			Value struct {
				X float64 `json:"x"`
				Y float64 `json:"y"`
			} `json:"value"`
			Subtype string `json:"subtype"`
		} `json:"result"`
	}
	if json.Unmarshal(raw, &payload) != nil || payload.Result.Subtype == "null" {
		return 0, 0, false
	}
	return payload.Result.Value.X, payload.Result.Value.Y, true
}

func runtimeBooleanResult(raw json.RawMessage) (bool, bool) {
	var payload struct {
		Result struct {
			Type  string `json:"type"`
			Value bool   `json:"value"`
		} `json:"result"`
	}
	if json.Unmarshal(raw, &payload) != nil || payload.Result.Type != "boolean" {
		return false, false
	}
	return payload.Result.Value, true
}

func clampRatio(value float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func isWindowSyncNavigableURL(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") || value == "about:blank"
}
