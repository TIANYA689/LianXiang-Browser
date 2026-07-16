package backend

import (
	"lianxiang-browser/backend/internal/automation"
	"lianxiang-browser/backend/internal/browser"
	"lianxiang-browser/backend/internal/config"
	"lianxiang-browser/backend/internal/database"
	"lianxiang-browser/backend/internal/launchcode"
	"lianxiang-browser/backend/internal/logger"
	"lianxiang-browser/backend/internal/proxy"
	"context"
	"strings"
	"sync"
)

type quitMode uint8

const (
	quitModeFull quitMode = iota
	quitModeAppOnly
)

// App 应用结构体
type App struct {
	ctx            context.Context
	config         *config.Config
	db             *database.DB
	interceptor    *logger.MethodInterceptor
	browserMgr     *browser.Manager
	xrayMgr        *proxy.XrayManager
	clashMgr       *proxy.ClashManager
	singboxMgr     *proxy.SingBoxManager
	launchCodeSvc  *launchcode.LaunchCodeService
	launchServer   *launchcode.LaunchServer
	automationMgr  *automation.Manager
	speedScheduler *browser.ProxySpeedScheduler
	appRoot        string
	version        string

	forceQuit              bool
	quitMode               quitMode
	maintenanceMu          sync.Mutex
	bridgeMu               sync.Mutex
	profileBridgeRefs      map[string]profileProxyBridgeRef
	deferredStartTargetsMu sync.Mutex
	deferredStartTargets   map[string][]string
	automationTargetMu     sync.Mutex
	automationTargetCursor map[string]string
	stopServicesOnce       sync.Once
	finalizeOnce           sync.Once
}

// NewApp 创建新的应用实例
func NewApp(appRoot string, appVersion ...string) *App {
	version := ""
	if len(appVersion) > 0 {
		version = strings.TrimSpace(appVersion[0])
	}
	return &App{
		appRoot:                strings.TrimSpace(appRoot),
		version:                version,
		profileBridgeRefs:      make(map[string]profileProxyBridgeRef),
		deferredStartTargets:   make(map[string][]string),
		automationTargetCursor: make(map[string]string),
	}
}

func (a *App) appName() string {
	if a.config != nil {
		if name := strings.TrimSpace(a.config.App.Name); name != "" {
			return name
		}
	}
	return "链享浏览器"
}

func (a *App) appVersion() string {
	version := strings.TrimSpace(a.version)
	if version == "" {
		return "unknown"
	}
	return version
}
