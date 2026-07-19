export interface BrowserProfile {
  profileId: string
  profileName: string
  userDataDir: string
  coreId: string
  fingerprintArgs: string[]
  proxyId: string
  proxyConfig: string
  proxyBindSourceId?: string
  proxyBindSourceUrl?: string
  proxyBindName?: string
  proxyBindUpdatedAt?: string
  launchArgs: string[]
  tags: string[]
  keywords: string[]
  groupId?: string
  running: boolean
  debugPort: number
  debugReady: boolean
  pid: number
  runtimeWarning: string
  lastError: string
  createdAt: string
  updatedAt: string
  deletedAt?: string
  lastStartAt?: string
  lastStopAt?: string
  launchCode?: string
}

export interface BrowserProfileInput {
  profileName: string
  userDataDir: string
  coreId: string
  fingerprintArgs: string[]
  proxyId: string
  proxyConfig: string
  launchArgs: string[]
  tags: string[]
  keywords: string[]
  groupId?: string
}

export interface BrowserProfilePackageExportResult {
  cancelled: boolean
  zipPath: string
  profileCount: number
  fileCount: number
  message: string
}

export interface BrowserProfilePackageImportResult {
  cancelled: boolean
  importedCount: number
  profileMappings: Record<string, string>
  warnings?: string[]
  message: string
}

export interface ChromeUserDataImportResult {
  cancelled: boolean
  profileId: string
  profileName: string
  sourceDir: string
  copiedFiles: number
  skippedFiles: number
  message: string
}

export type BrowserProfileCopyMode = 'auto_fingerprint' | 'regular'

export type BrowserProfileAutomationTarget =
  | 'seed'
  | 'identity'
  | 'locale'
  | 'screen'
  | 'hardware'
  | 'render'
  | 'fonts'
  | 'network'
  | 'devices'

export interface BrowserProfileCopyOptions {
  mode: BrowserProfileCopyMode
  automationTargets: BrowserProfileAutomationTarget[]
}

export interface BrowserTab {
  tabId: string
  title: string
  url: string
  active: boolean
}

export interface BrowserSettings {
  userDataRoot: string
  defaultFingerprintArgs: string[]
  defaultLaunchArgs: string[]
  defaultStartUrls: string[]
  lightStartEnabled: boolean
  restoreLastSession: boolean
  startReadyTimeoutMs: number
  startStableWindowMs: number
  // xray 表示 Xray + sing-box 组合连接栈；mihomo 表示独立 Mihomo 连接栈。
  defaultConnectorType: 'xray' | 'mihomo' | string
}

export interface ProxyCheckTarget {
  id: string
  name: string
  type: string
  url: string
  parser?: string
  timeoutMs?: number
  expectedStatus?: number[]
}

export interface ProxyCheckSettings {
  bridgeStartTimeoutMs: number
  speedTargetId: string
  ipHealthTargetId: string
  targets: ProxyCheckTarget[]
}

export interface BrowserCore {
  coreId: string
  coreName: string
  corePath: string
  isDefault: boolean
}

export interface BrowserCoreInput {
  coreId: string
  coreName: string
  corePath: string
  isDefault: boolean
}

export interface BrowserCoreValidateResult {
  valid: boolean
  message: string
}

export interface BrowserProxy {
  proxyId: string
  proxyName: string
  proxyConfig: string
  preferredKernel?: 'auto' | 'xray' | 'sing-box' | 'mihomo' | string
  dnsServers?: string
  groupName?: string
  sourceId?: string
  sourceUrl?: string
  sourceNamePrefix?: string
  sourceAutoRefresh?: boolean
  sourceRefreshIntervalM?: number
  sourceLastRefreshAt?: string
  lastLatencyMs?: number
  lastTestOk?: boolean
  lastTestedAt?: string
  lastIPHealthJson?: string
}

export interface ProxyIPHealthResult {
  proxyId: string
  ok: boolean
  source: string
  error: string
  ip: string
  fraudScore: number
  isResidential: boolean
  isBroadcast: boolean
  country: string
  region: string
  city: string
  asOrganization: string
  rawData: Record<string, any>
  updatedAt: string
}


export interface ProxyCoreDownloadProgress {
  core: string
  goos: string
  goarch: string
  phase: string
  progress: number
  message: string
}

export interface ProxyCoreStatusResult {
  core: string
  goos: string
  goarch: string
  installed: boolean
  configured: boolean
  active: boolean
  binaryPath: string
  source: string
  message: string
}

export interface ProxyCoreDownloadInfoResult {
  core: string
  goos: string
  goarch: string
  version: string
  repo: string
  releaseUrl: string
  downloadUrl: string
  assetName: string
  installDir: string
  binaryName: string
  message: string
}

export interface ProxyBridgeWarmupResult {
  proxyId: string
  ok: boolean
  engine: string
  socksUrl: string
  latencyMs: number
  error: string
}

export interface ProxySpeedTestResult {
  proxyId: string
  ok: boolean
  latencyMs: number
  engine?: string
  error: string
}


export interface ProxyLocationOption {
  label: string
  timezone: string
  lang: string
}

export interface ProxyLocationResolveResult {
  proxyId: string
  ok: boolean
  auto: boolean
  source: string
  error: string
  ip: string
  country: string
  region: string
  city: string
  timezone: string
  lang: string
  health?: ProxyIPHealthResult
  alternates?: ProxyLocationOption[]
  resolvedAt: string
}

export interface BrowserCoreExtended {
  coreId: string
  chromeVersion: string
  instanceCount: number
}

export interface BrowserExtension {
  extensionId: string
  name: string
  version: string
  description: string
  iconDataUrl: string
  manifestJson: string
  sourceUrl: string
  installDir: string
  enabled: boolean
  installedAt: string
  updatedAt: string
}

export interface BrowserExtensionLookupResult {
  extensionId: string
  name: string
  version: string
  description: string
  storeUrl: string
  installable: boolean
  message: string
}

export interface BrowserProfileExtensionSettings {
  profileId: string
  configured: boolean
  extensionIds: string[]
  updatedAt: string
}

export interface CookieInfo {
  name: string
  value: string
  domain: string
  path: string
  expires: number
  httpOnly: boolean
  secure: boolean
  sameSite: string
}

export interface SnapshotInfo {
  snapshotId: string
  profileId: string
  name: string
  sizeMB: number
  createdAt: string
}

export interface BrowserBookmark {
  name: string
  url: string
  folder?: string
  openOnStart?: boolean
  disabled?: boolean
  disabledProfileIds?: string[]
}

export interface WindowSyncSettings {
  syncClicks: boolean
  syncInputs: boolean
  syncScroll: boolean
  syncNavigation: boolean
}

export interface WindowSyncStartInput {
  masterProfileId: string
  targetProfileIds: string[]
  settings: WindowSyncSettings
}

export interface WindowSyncSession {
  profileId: string
  profileName: string
  tags: string[]
  running: boolean
  debugReady: boolean
  debugPort: number
  pageTitle: string
  pageUrl: string
  available: boolean
  warning: string
}

export interface WindowSyncState {
  active: boolean
  masterProfileId: string
  targetProfileIds: string[]
  settings: WindowSyncSettings
  startedAt: string
  lastEventAt: string
  lastEventType: string
  eventCount: number
  lastError: string
}

export interface BookmarkSyncResult {
  total: number
  synced: number
  skipped: number
  failed: number
  skippedList: string[]
  failedList: string[]
}


// 分组相关类型
export interface BrowserGroup {
  groupId: string
  groupName: string
  parentId: string
  sortOrder: number
  createdAt: string
  updatedAt: string
}

export interface BrowserGroupInput {
  groupName: string
  parentId: string
  sortOrder: number
}

export interface BrowserGroupWithCount extends BrowserGroup {
  instanceCount: number
}
