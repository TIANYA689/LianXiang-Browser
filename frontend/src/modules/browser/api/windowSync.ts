import type {
  WindowSyncSession,
  WindowSyncActionResult,
  WindowSyncSettings,
  WindowSyncStartInput,
  WindowSyncState,
} from '../types'
import { getBindings, getMockProfiles } from './runtime'

export const defaultWindowSyncSettings: WindowSyncSettings = {
  syncClicks: true,
  syncInputs: true,
  syncScroll: true,
  syncNavigation: true,
}

export const emptyWindowSyncState: WindowSyncState = {
  active: false,
  masterProfileId: '',
  targetProfileIds: [],
  settings: defaultWindowSyncSettings,
  startedAt: '',
  lastEventAt: '',
  lastEventType: '',
  eventCount: 0,
  lastError: '',
}

export async function fetchWindowSyncSessions(): Promise<WindowSyncSession[]> {
  const bindings: any = await getBindings()
  if (bindings?.WindowSyncListSessions) {
    return (await bindings.WindowSyncListSessions()) || []
  }
  return getMockProfiles().map((profile) => ({
    profileId: profile.profileId,
    profileName: profile.profileName,
    tags: profile.tags || [],
    running: profile.running,
    debugReady: profile.debugReady,
    debugPort: profile.debugPort,
    pageTitle: profile.running ? '示例页面' : '',
    pageUrl: profile.running ? 'https://example.com' : '',
    available: profile.running && profile.debugReady,
    warning: profile.running && profile.debugReady ? '' : '实例未启动',
  }))
}

export async function fetchWindowSyncState(): Promise<WindowSyncState> {
  const bindings: any = await getBindings()
  if (bindings?.WindowSyncGetState) {
    return (await bindings.WindowSyncGetState()) || emptyWindowSyncState
  }
  return emptyWindowSyncState
}

export async function startWindowSync(input: WindowSyncStartInput): Promise<WindowSyncState> {
  const bindings: any = await getBindings()
  if (bindings?.WindowSyncStart) {
    return await bindings.WindowSyncStart(input)
  }
  return {
    ...emptyWindowSyncState,
    active: true,
    masterProfileId: input.masterProfileId,
    targetProfileIds: input.targetProfileIds,
    settings: input.settings,
    startedAt: new Date().toISOString(),
  }
}

export async function stopWindowSync(): Promise<WindowSyncState> {
  const bindings: any = await getBindings()
  if (bindings?.WindowSyncStop) {
    return await bindings.WindowSyncStop()
  }
  return emptyWindowSyncState
}

export async function showWindowSyncProfile(profileId: string): Promise<void> {
  const bindings: any = await getBindings()
  if (bindings?.WindowSyncShowWindow) {
    await bindings.WindowSyncShowWindow(profileId)
  }
}

export async function runWindowSyncWindowAction(profileIds: string[], action: string): Promise<WindowSyncActionResult> {
  const bindings: any = await getBindings()
  if (bindings?.WindowSyncWindowAction) {
    return await bindings.WindowSyncWindowAction(profileIds, action)
  }
  return { requested: profileIds.length, succeeded: profileIds.length, failed: [] }
}

export async function runWindowSyncTextAction(profileIds: string[], text: string, clear: boolean): Promise<WindowSyncActionResult> {
  const bindings: any = await getBindings()
  if (bindings?.WindowSyncTextAction) {
    return await bindings.WindowSyncTextAction(profileIds, text, clear)
  }
  return { requested: profileIds.length, succeeded: profileIds.length, failed: [] }
}

export async function runWindowSyncTabAction(profileIds: string[], action: string, targetUrl = ''): Promise<WindowSyncActionResult> {
  const bindings: any = await getBindings()
  if (bindings?.WindowSyncTabAction) {
    return await bindings.WindowSyncTabAction(profileIds, action, targetUrl)
  }
  return { requested: profileIds.length, succeeded: profileIds.length, failed: [] }
}

export async function copyMasterWindowSyncTabs(masterProfileId: string, targetProfileIds: string[]): Promise<WindowSyncActionResult> {
  const bindings: any = await getBindings()
  if (bindings?.WindowSyncCopyMasterTabs) {
    return await bindings.WindowSyncCopyMasterTabs(masterProfileId, targetProfileIds)
  }
  return { requested: targetProfileIds.length, succeeded: targetProfileIds.length, failed: [] }
}
