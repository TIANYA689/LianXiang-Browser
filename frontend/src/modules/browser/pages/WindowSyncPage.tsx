import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  AlertTriangle,
  Crown,
  ExternalLink,
  Keyboard,
  MousePointer2,
  Navigation,
  Play,
  RefreshCw,
  Square,
  TextCursorInput,
  Waypoints,
} from 'lucide-react'
import { Badge, Button, Card, toast } from '../../../shared/components'
import {
  defaultWindowSyncSettings,
  emptyWindowSyncState,
  fetchWindowSyncSessions,
  fetchWindowSyncState,
  showWindowSyncProfile,
  startWindowSync,
  stopWindowSync,
} from '../api/windowSync'
import type { WindowSyncSession, WindowSyncSettings, WindowSyncState } from '../types'

const settingOptions: Array<{
  key: keyof WindowSyncSettings
  title: string
  description: string
  icon: typeof MousePointer2
}> = [
  { key: 'syncClicks', title: '点击与确认键', description: '同步网页点击及 Enter、Tab、Esc', icon: MousePointer2 },
  { key: 'syncInputs', title: '输入内容', description: '同步输入框、选择框和可编辑文本', icon: TextCursorInput },
  { key: 'syncScroll', title: '页面滚动', description: '按页面滚动比例同步当前位置', icon: Waypoints },
  { key: 'syncNavigation', title: '页面跳转', description: '主窗口跳转时同步当前活动页网址', icon: Navigation },
]

function formatStateTime(value: string) {
  if (!value) return '尚无事件'
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleTimeString('zh-CN', { hour12: false })
}

function eventTypeLabel(value: string) {
  return ({ click: '点击', input: '输入', scroll: '滚动', key: '按键', navigation: '跳转' } as Record<string, string>)[value] || '等待操作'
}

export function WindowSyncPage() {
  const [sessions, setSessions] = useState<WindowSyncSession[]>([])
  const [syncState, setSyncState] = useState<WindowSyncState>(emptyWindowSyncState)
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set())
  const [masterId, setMasterId] = useState('')
  const [settings, setSettings] = useState<WindowSyncSettings>(defaultWindowSyncSettings)
  const [loading, setLoading] = useState(true)
  const [action, setAction] = useState<'start' | 'stop' | 'refresh' | string>('')

  const load = useCallback(async (silent = false) => {
    if (!silent) setLoading(true)
    try {
      const [nextSessions, nextState] = await Promise.all([
        fetchWindowSyncSessions(),
        fetchWindowSyncState(),
      ])
      setSessions(nextSessions)
      setSyncState(nextState)
      const availableIds = new Set(nextSessions.filter((item) => item.available).map((item) => item.profileId))
      if (nextState.active) {
        const lockedIds = new Set([nextState.masterProfileId, ...(nextState.targetProfileIds || [])])
        setSelectedIds(lockedIds)
        setMasterId(nextState.masterProfileId)
        setSettings(nextState.settings || defaultWindowSyncSettings)
      } else {
        setSelectedIds((current) => {
          const retained = new Set([...current].filter((id) => availableIds.has(id)))
          if (retained.size > 0) return retained
          return new Set(nextSessions.filter((item) => item.available).map((item) => item.profileId))
        })
        setMasterId((current) => availableIds.has(current) ? current : nextSessions.find((item) => item.available)?.profileId || '')
      }
    } catch (error: any) {
      if (!silent) toast.error(error?.message || '窗口同步状态加载失败')
    } finally {
      if (!silent) setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
    const timer = window.setInterval(() => void load(true), 4000)
    return () => window.clearInterval(timer)
  }, [load])

  const availableSessions = useMemo(() => sessions.filter((item) => item.available), [sessions])
  const selectedTargets = useMemo(
    () => [...selectedIds].filter((id) => id !== masterId),
    [masterId, selectedIds],
  )

  const toggleSession = (session: WindowSyncSession) => {
    if (syncState.active || !session.available) return
    setSelectedIds((current) => {
      const next = new Set(current)
      if (next.has(session.profileId)) {
        next.delete(session.profileId)
        if (session.profileId === masterId) {
          const nextMaster = [...next][0] || ''
          setMasterId(nextMaster)
        }
      } else {
        next.add(session.profileId)
        if (!masterId) setMasterId(session.profileId)
      }
      return next
    })
  }

  const chooseMaster = (profileId: string) => {
    if (syncState.active) return
    setSelectedIds((current) => new Set([...current, profileId]))
    setMasterId(profileId)
  }

  const handleStart = async () => {
    if (!masterId) {
      toast.warning('请先设置主窗口')
      return
    }
    if (selectedTargets.length === 0) {
      toast.warning('请至少选择一个被控窗口')
      return
    }
    setAction('start')
    try {
      const state = await startWindowSync({
        masterProfileId: masterId,
        targetProfileIds: selectedTargets,
        settings,
      })
      setSyncState(state)
      toast.success(`同步已启动：1 个主窗口，${state.targetProfileIds.length} 个被控窗口`)
      await load(true)
    } catch (error: any) {
      toast.error(error?.message || '窗口同步启动失败', 6000)
    } finally {
      setAction('')
    }
  }

  const handleStop = async () => {
    setAction('stop')
    try {
      const state = await stopWindowSync()
      setSyncState(state)
      toast.success('窗口同步已停止')
      await load(true)
    } catch (error: any) {
      toast.error(error?.message || '停止窗口同步失败')
    } finally {
      setAction('')
    }
  }

  const handleRefresh = async () => {
    setAction('refresh')
    await load()
    setAction('')
  }

  const handleShowWindow = async (profileId: string) => {
    setAction(`show:${profileId}`)
    try {
      await showWindowSyncProfile(profileId)
      toast.success('浏览器窗口已恢复显示')
    } catch (error: any) {
      toast.error(error?.message || '显示窗口失败')
    } finally {
      setAction('')
    }
  }

  const selectAllAvailable = () => {
    if (syncState.active) return
    const all = new Set(availableSessions.map((item) => item.profileId))
    setSelectedIds(all)
    if (!all.has(masterId)) setMasterId(availableSessions[0]?.profileId || '')
  }

  return (
    <div className="mx-auto flex min-h-full max-w-[1500px] flex-col gap-4 animate-fade-in">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <div className="mb-1 flex items-center gap-2">
            <h1 className="text-xl font-semibold text-[var(--color-text-primary)]">窗口同步器</h1>
            <Badge variant={syncState.active ? 'success' : 'default'} dot>
              {syncState.active ? '同步中' : '未启动'}
            </Badge>
          </div>
          <p className="text-sm text-[var(--color-text-muted)]">
            在主窗口操作当前网页，点击、输入、滚动和页面跳转会同步到被控窗口。
          </p>
        </div>
        <Button variant="secondary" onClick={handleRefresh} loading={action === 'refresh'}>
          <RefreshCw className="h-4 w-4" />
          刷新实例
        </Button>
      </div>

      <div className="flex items-start gap-3 rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 text-amber-900 dark:border-amber-900/60 dark:bg-amber-950/30 dark:text-amber-200">
        <AlertTriangle className="mt-0.5 h-4 w-4 flex-none" />
        <p className="text-xs leading-5">
          同步范围是浏览器当前活动网页。地址栏、书签栏、扩展弹窗、下载框和系统对话框属于浏览器原生界面，不会同步；启动前请让各窗口保持相近尺寸和相同页面结构。
        </p>
      </div>

      <Card padding="none">
        <div className="flex flex-wrap items-center justify-between gap-3 border-b border-[var(--color-border-muted)] px-5 py-4">
          <div>
            <h2 className="text-sm font-semibold text-[var(--color-text-primary)]">同步窗口</h2>
            <p className="mt-0.5 text-xs text-[var(--color-text-muted)]">
              已选 {selectedIds.size} 个 · 主窗口 1 个 · 被控窗口 {selectedTargets.length} 个
            </p>
          </div>
          <button
            type="button"
            className="text-xs font-medium text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)] disabled:opacity-50"
            onClick={selectAllAvailable}
            disabled={syncState.active || availableSessions.length === 0}
          >
            选择全部可用实例
          </button>
        </div>

        <div className="overflow-x-auto">
          <table className="min-w-full">
            <thead>
              <tr className="bg-[var(--color-bg-muted)]">
                <th className="w-12 px-4 py-3 text-left text-xs font-semibold text-[var(--color-text-muted)]">选择</th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-[var(--color-text-muted)]">实例</th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-[var(--color-text-muted)]">当前活动页</th>
                <th className="w-28 px-4 py-3 text-left text-xs font-semibold text-[var(--color-text-muted)]">角色</th>
                <th className="w-48 px-4 py-3 text-right text-xs font-semibold text-[var(--color-text-muted)]">操作</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[var(--color-border-muted)]">
              {loading ? (
                <tr><td colSpan={5} className="px-5 py-14 text-center text-sm text-[var(--color-text-muted)]">正在读取运行实例...</td></tr>
              ) : sessions.length === 0 ? (
                <tr><td colSpan={5} className="px-5 py-14 text-center text-sm text-[var(--color-text-muted)]">暂无浏览器实例，请先在实例列表中创建并启动实例。</td></tr>
              ) : sessions.map((session) => {
                const selected = selectedIds.has(session.profileId)
                const isMaster = selected && masterId === session.profileId
                return (
                  <tr key={session.profileId} className={isMaster ? 'bg-[var(--color-accent-muted)]/80' : 'hover:bg-[var(--color-bg-muted)]/50'}>
                    <td className="px-4 py-3">
                      <input
                        type="checkbox"
                        className="h-4 w-4 accent-[var(--color-accent)]"
                        checked={selected}
                        disabled={syncState.active || !session.available}
                        onChange={() => toggleSession(session)}
                        aria-label={`选择 ${session.profileName}`}
                      />
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-2">
                        <span className="text-sm font-medium text-[var(--color-text-primary)]">{session.profileName}</span>
                        <Badge variant={session.available ? 'success' : session.running ? 'warning' : 'default'} size="sm" dot>
                          {session.available ? '可同步' : session.running ? '连接中' : '未启动'}
                        </Badge>
                      </div>
                      <div className="mt-1 flex flex-wrap gap-1">
                        {session.tags?.slice(0, 3).map((tag) => <Badge key={tag} size="sm">{tag}</Badge>)}
                        {!session.available && <span className="text-xs text-[var(--color-text-muted)]">{session.warning}</span>}
                      </div>
                    </td>
                    <td className="max-w-[520px] px-4 py-3">
                      {session.available ? (
                        <div className="min-w-0">
                          <p className="truncate text-sm text-[var(--color-text-secondary)]">{session.pageTitle || '未命名页面'}</p>
                          <p className="truncate text-xs text-[var(--color-text-muted)]" title={session.pageUrl}>{session.pageUrl}</p>
                        </div>
                      ) : <span className="text-xs text-[var(--color-text-muted)]">—</span>}
                    </td>
                    <td className="px-4 py-3">
                      {isMaster ? (
                        <Badge variant="info"><Crown className="h-3 w-3" />主窗口</Badge>
                      ) : selected ? (
                        <Badge>被控窗口</Badge>
                      ) : (
                        <span className="text-xs text-[var(--color-text-muted)]">未选择</span>
                      )}
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex justify-end gap-2">
                        <Button
                          size="sm"
                          variant="ghost"
                          disabled={!session.available}
                          loading={action === `show:${session.profileId}`}
                          onClick={() => handleShowWindow(session.profileId)}
                        >
                          <ExternalLink className="h-3.5 w-3.5" />显示窗口
                        </Button>
                        <Button
                          size="sm"
                          variant={isMaster ? 'secondary' : 'ghost'}
                          disabled={syncState.active || !session.available || isMaster}
                          onClick={() => chooseMaster(session.profileId)}
                        >
                          <Crown className="h-3.5 w-3.5" />{isMaster ? '当前主窗口' : '设为主窗口'}
                        </Button>
                      </div>
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      </Card>

      <Card title="同步内容" subtitle="同步运行后将锁定窗口选择与设置；停止后可以重新调整。">
        <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
          {settingOptions.map((option) => {
            const Icon = option.icon
            const enabled = settings[option.key]
            return (
              <button
                type="button"
                key={option.key}
                disabled={syncState.active}
                onClick={() => setSettings((current) => ({ ...current, [option.key]: !current[option.key] }))}
                className={`flex items-start gap-3 rounded-lg border p-3 text-left transition-colors ${enabled ? 'border-[var(--color-accent)] bg-[var(--color-accent-muted)]' : 'border-[var(--color-border-default)] bg-[var(--color-bg-surface)]'} disabled:cursor-not-allowed disabled:opacity-70`}
              >
                <span className={`mt-0.5 flex h-8 w-8 items-center justify-center rounded-lg ${enabled ? 'bg-[var(--color-accent)] text-[var(--color-text-inverse)]' : 'bg-[var(--color-bg-muted)] text-[var(--color-text-muted)]'}`}>
                  <Icon className="h-4 w-4" />
                </span>
                <span className="min-w-0 flex-1">
                  <span className="flex items-center justify-between gap-2 text-sm font-medium text-[var(--color-text-primary)]">
                    {option.title}
                    <span className={`h-2 w-2 rounded-full ${enabled ? 'bg-[var(--color-success)]' : 'bg-[var(--color-border-strong)]'}`} />
                  </span>
                  <span className="mt-1 block text-xs leading-5 text-[var(--color-text-muted)]">{option.description}</span>
                </span>
              </button>
            )
          })}
        </div>
      </Card>

      <div className="sticky bottom-0 z-10 flex flex-wrap items-center justify-between gap-4 rounded-xl border border-[var(--color-border-default)] bg-[var(--color-bg-elevated)]/95 px-5 py-4 shadow-[var(--shadow-lg)] backdrop-blur">
        <div className="flex min-w-0 items-center gap-3">
          <span className={`flex h-9 w-9 flex-none items-center justify-center rounded-full ${syncState.active ? 'bg-[var(--color-success)]/15 text-[var(--color-success)]' : 'bg-[var(--color-bg-muted)] text-[var(--color-text-muted)]'}`}>
            {syncState.active ? <Keyboard className="h-4 w-4" /> : <MousePointer2 className="h-4 w-4" />}
          </span>
          <div className="min-w-0">
            <p className="text-sm font-medium text-[var(--color-text-primary)]">
              {syncState.active ? `同步运行中 · 已转发 ${syncState.eventCount} 个事件` : '选择主窗口和被控窗口后启动同步'}
            </p>
            <p className={`truncate text-xs ${syncState.lastError ? 'text-[var(--color-error)]' : 'text-[var(--color-text-muted)]'}`}>
              {syncState.lastError || `${eventTypeLabel(syncState.lastEventType)} · ${formatStateTime(syncState.lastEventAt)}`}
            </p>
          </div>
        </div>
        {syncState.active ? (
          <Button variant="danger" onClick={handleStop} loading={action === 'stop'}>
            <Square className="h-4 w-4" />停止同步
          </Button>
        ) : (
          <Button onClick={handleStart} loading={action === 'start'} disabled={!masterId || selectedTargets.length === 0}>
            <Play className="h-4 w-4" />启动同步
          </Button>
        )}
      </div>
    </div>
  )
}
