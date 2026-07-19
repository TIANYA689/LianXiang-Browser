import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  AlignJustify,
  Columns,
  Copy,
  Crown,
  ExternalLink,
  Grid2X2,
  Maximize2,
  Minimize2,
  MonitorDown,
  Navigation,
  Play,
  Plus,
  RefreshCw,
  RotateCcw,
  Square,
  TextCursorInput,
  Trash2,
  X,
} from 'lucide-react'
import { Badge, Button, ConfirmModal, Input, Select, Switch, Textarea, toast } from '../../../shared/components'
import {
  copyMasterWindowSyncTabs,
  defaultWindowSyncSettings,
  emptyWindowSyncState,
  fetchWindowSyncSessions,
  fetchWindowSyncState,
  runWindowSyncTabAction,
  runWindowSyncTextAction,
  runWindowSyncWindowAction,
  showWindowSyncProfile,
  startWindowSync,
  stopWindowSync,
} from '../api/windowSync'
import type { WindowSyncActionResult, WindowSyncSession, WindowSyncSettings, WindowSyncState } from '../types'

type WorkspaceTab = 'window' | 'text' | 'tabs'

const workspaceTabs: Array<{ id: WorkspaceTab; label: string; icon: typeof Grid2X2 }> = [
  { id: 'window', label: '窗口管理', icon: Grid2X2 },
  { id: 'text', label: '文本管理', icon: TextCursorInput },
  { id: 'tabs', label: '标签页管理', icon: AlignJustify },
]

const syncOptions: Array<{ key: keyof WindowSyncSettings; label: string }> = [
  { key: 'syncClicks', label: '点击与确认键' },
  { key: 'syncInputs', label: '输入内容' },
  { key: 'syncScroll', label: '页面滚动' },
  { key: 'syncNavigation', label: '页面跳转' },
]

function formatActionResult(label: string, result: WindowSyncActionResult) {
  if (result.failed.length) {
    const detail = result.failed.slice(0, 2).join('；')
    toast.warning(label + '完成：' + result.succeeded + '/' + result.requested + ' 个窗口成功' + (detail ? '。失败：' + detail : ''), 7000)
    return
  }
  toast.success(label + '已应用到 ' + result.succeeded + ' 个窗口')
}

export function WindowSyncPage() {
  const [sessions, setSessions] = useState<WindowSyncSession[]>([])
  const [syncState, setSyncState] = useState<WindowSyncState>(emptyWindowSyncState)
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set())
  const [masterId, setMasterId] = useState('')
  const [settings, setSettings] = useState<WindowSyncSettings>(defaultWindowSyncSettings)
  const [activeTab, setActiveTab] = useState<WorkspaceTab>('window')
  const [groupFilter, setGroupFilter] = useState('all')
  const [text, setText] = useState('')
  const [targetUrl, setTargetUrl] = useState('https://example.com')
  const [loading, setLoading] = useState(true)
  const [action, setAction] = useState('')
  const [confirmTabSync, setConfirmTabSync] = useState(false)

  const load = useCallback(async (silent = false) => {
    if (!silent) setLoading(true)
    try {
      const [nextSessions, nextState] = await Promise.all([fetchWindowSyncSessions(), fetchWindowSyncState()])
      setSessions(nextSessions)
      setSyncState(nextState)
      const availableIds = new Set(nextSessions.filter((item) => item.available).map((item) => item.profileId))
      if (nextState.active) {
        setSelectedIds(new Set([nextState.masterProfileId, ...(nextState.targetProfileIds || [])]))
        setMasterId(nextState.masterProfileId)
        setSettings(nextState.settings || defaultWindowSyncSettings)
      } else {
        setSelectedIds((current) => {
          const retained = new Set([...current].filter((id) => availableIds.has(id)))
          return retained.size ? retained : new Set(nextSessions.filter((item) => item.available).map((item) => item.profileId))
        })
        setMasterId((current) => availableIds.has(current) ? current : nextSessions.find((item) => item.available)?.profileId || '')
      }
    } catch (error: any) {
      if (!silent) toast.error(error?.message || '窗口会话读取失败')
    } finally {
      if (!silent) setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
    const timer = window.setInterval(() => void load(true), 4000)
    return () => window.clearInterval(timer)
  }, [load])

  const groups = useMemo(() => [...new Set(sessions.flatMap((item) => item.tags || []))].sort(), [sessions])
  const visibleSessions = useMemo(
    () => groupFilter === 'all' ? sessions : sessions.filter((item) => item.tags?.includes(groupFilter)),
    [groupFilter, sessions],
  )
  const availableVisibleSessions = useMemo(() => visibleSessions.filter((item) => item.available), [visibleSessions])
  const allVisibleSelected = availableVisibleSessions.length > 0 && availableVisibleSessions.every((item) => selectedIds.has(item.profileId))
  const selectedProfileIds = useMemo(() => [...selectedIds], [selectedIds])
  const targetIds = useMemo(() => selectedProfileIds.filter((id) => id !== masterId), [masterId, selectedProfileIds])
  const canRun = selectedProfileIds.length > 0

  const toggleSession = (session: WindowSyncSession) => {
    if (syncState.active || !session.available) return
    setSelectedIds((current) => {
      const next = new Set(current)
      if (next.has(session.profileId)) {
        next.delete(session.profileId)
        if (session.profileId === masterId) setMasterId([...next][0] || '')
      } else {
        next.add(session.profileId)
        if (!masterId) setMasterId(session.profileId)
      }
      return next
    })
  }

  const selectAllVisible = () => {
    if (syncState.active) return
    const availableVisibleIds = availableVisibleSessions.map((item) => item.profileId)
    setSelectedIds((current) => {
      const next = new Set(current)
      for (const id of availableVisibleIds) {
        if (allVisibleSelected) next.delete(id)
        else next.add(id)
      }
      if (masterId && !next.has(masterId)) setMasterId([...next][0] || '')
      else if (!masterId) setMasterId([...next][0] || '')
      return next
    })
  }

  const chooseMaster = (profileId: string) => {
    if (syncState.active) return
    setSelectedIds((current) => new Set([...current, profileId]))
    setMasterId(profileId)
  }

  const withAction = async (key: string, task: () => Promise<WindowSyncActionResult>, label: string) => {
    if (!canRun) {
      toast.warning('请先选择至少一个可用窗口')
      return
    }
    setAction(key)
    try {
      formatActionResult(label, await task())
      await load(true)
    } catch (error: any) {
      toast.error(error?.message || label + '失败', 6000)
    } finally {
      setAction('')
    }
  }

  const handleStart = async () => {
    if (!masterId || targetIds.length === 0) {
      toast.warning('请选中至少两个窗口，并设置其中一个为主窗口')
      return
    }
    setAction('start')
    try {
      const state = await startWindowSync({ masterProfileId: masterId, targetProfileIds: targetIds, settings })
      setSyncState(state)
      toast.success('同步已启动：主窗口会驱动 ' + state.targetProfileIds.length + ' 个被控窗口')
    } catch (error: any) {
      toast.error(error?.message || '窗口同步启动失败')
    } finally {
      setAction('')
    }
  }

  const handleStop = async () => {
    setAction('stop')
    try {
      setSyncState(await stopWindowSync())
      toast.success('窗口同步已停止')
    } catch (error: any) {
      toast.error(error?.message || '停止同步失败')
    } finally {
      setAction('')
    }
  }

  const handleShow = async (profileId: string) => {
    setAction('show:' + profileId)
    try {
      await showWindowSyncProfile(profileId)
      toast.success('已恢复显示浏览器窗口')
    } catch (error: any) {
      toast.error(error?.message || '显示窗口失败')
    } finally {
      setAction('')
    }
  }

  return (
    <div className="mx-auto flex min-h-full max-w-[1720px] flex-col gap-4 animate-fade-in">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-xl font-semibold text-[var(--color-text-primary)]">窗口同步器</h1>
            <Badge variant={syncState.active ? 'success' : 'default'} dot>{syncState.active ? '同步运行中' : '同步未启动'}</Badge>
          </div>
          <p className="mt-1 text-sm text-[var(--color-text-muted)]">窗口、文本和标签页的本机 CDP 批量管理</p>
        </div>
        <Button variant="secondary" onClick={() => { setAction('refresh'); void load().finally(() => setAction('')) }} loading={action === 'refresh'}>
          <RefreshCw className="h-4 w-4" />刷新会话
        </Button>
      </div>

      <div className="grid min-h-[680px] gap-4 xl:grid-cols-[minmax(0,1.65fr)_minmax(400px,0.9fr)]">
        <section className="flex min-h-0 flex-col rounded-xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)]">
          <div className="flex flex-wrap items-center gap-2 border-b border-[var(--color-border-muted)] p-4">
            <Select value={groupFilter} onChange={(event) => setGroupFilter(event.target.value)} options={[{ value: 'all', label: '全部分组' }, ...groups.map((group) => ({ value: group, label: group }))]} className="w-44" />
            {syncState.active ? (
              <Button variant="danger" onClick={handleStop} loading={action === 'stop'}><Square className="h-4 w-4" />停止同步</Button>
            ) : (
              <Button onClick={handleStart} loading={action === 'start'} disabled={!masterId || targetIds.length === 0}><Play className="h-4 w-4" />启动同步</Button>
            )}
            <Button variant="secondary" onClick={selectAllVisible} disabled={syncState.active || availableVisibleSessions.length === 0}>{allVisibleSelected ? '取消选择当前分组' : '选择当前分组'}</Button>
            <span className="ml-auto text-xs text-[var(--color-text-muted)]">已选 <strong className="text-[var(--color-text-primary)]">{selectedProfileIds.length}</strong> 个 · 主窗口 {masterId ? '已设置' : '未设置'}</span>
          </div>
          <div className="min-h-0 flex-1 overflow-auto">
            <table className="min-w-full">
              <thead className="sticky top-0 z-10 bg-[var(--color-bg-muted)]">
                <tr>
                  <th className="w-12 px-4 py-3 text-left text-xs font-semibold text-[var(--color-text-muted)]"><input type="checkbox" className="h-4 w-4 accent-[var(--color-accent)]" checked={allVisibleSelected} onChange={selectAllVisible} disabled={syncState.active || availableVisibleSessions.length === 0} /></th>
                  <th className="px-4 py-3 text-left text-xs font-semibold text-[var(--color-text-muted)]">实例名称</th>
                  <th className="px-4 py-3 text-left text-xs font-semibold text-[var(--color-text-muted)]">当前活动页</th>
                  <th className="w-28 px-4 py-3 text-left text-xs font-semibold text-[var(--color-text-muted)]">状态</th>
                  <th className="w-40 px-4 py-3 text-right text-xs font-semibold text-[var(--color-text-muted)]">操作</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[var(--color-border-muted)]">
                {loading ? <tr><td colSpan={5} className="px-5 py-20 text-center text-sm text-[var(--color-text-muted)]">正在读取已启动的浏览器会话...</td></tr> : null}
                {!loading && visibleSessions.length === 0 ? <tr><td colSpan={5} className="px-5 py-20 text-center text-sm text-[var(--color-text-muted)]">暂无已打开的浏览器环境。请先在实例列表启动至少两个实例。</td></tr> : null}
                {!loading && visibleSessions.map((session) => {
                  const selected = selectedIds.has(session.profileId)
                  const isMaster = selected && masterId === session.profileId
                  return <tr key={session.profileId} className={isMaster ? 'bg-[var(--color-accent-muted)]/70' : 'hover:bg-[var(--color-bg-muted)]/40'}>
                    <td className="px-4 py-3"><input type="checkbox" className="h-4 w-4 accent-[var(--color-accent)]" checked={selected} disabled={!session.available || syncState.active} onChange={() => toggleSession(session)} /></td>
                    <td className="px-4 py-3"><div className="flex items-center gap-2"><span className="text-sm font-medium text-[var(--color-text-primary)]">{session.profileName}</span>{isMaster ? <Badge variant="info" size="sm"><Crown className="h-3 w-3" />主控</Badge> : null}</div><div className="mt-1 flex flex-wrap gap-1">{session.tags?.map((tag) => <Badge key={tag} size="sm">{tag}</Badge>)}</div></td>
                    <td className="max-w-[420px] px-4 py-3">{session.available ? <div className="min-w-0"><p className="truncate text-sm text-[var(--color-text-secondary)]">{session.pageTitle || '未命名页面'}</p><p className="truncate text-xs text-[var(--color-text-muted)]">{session.pageUrl}</p></div> : <span className="text-xs text-[var(--color-text-muted)]">{session.warning || '未就绪'}</span>}</td>
                    <td className="px-4 py-3"><Badge variant={session.available ? 'success' : session.running ? 'warning' : 'default'} size="sm" dot>{session.available ? '可用' : session.running ? '连接中' : '未启动'}</Badge></td>
                    <td className="px-4 py-3"><div className="flex justify-end gap-1"><Button size="sm" variant="ghost" disabled={!session.available} loading={action === 'show:' + session.profileId} onClick={() => void handleShow(session.profileId)}><ExternalLink className="h-3.5 w-3.5" />显示</Button><Button size="sm" variant={isMaster ? 'secondary' : 'ghost'} disabled={syncState.active || !session.available || isMaster} onClick={() => chooseMaster(session.profileId)}><Crown className="h-3.5 w-3.5" />主控</Button></div></td>
                  </tr>
                })}
              </tbody>
            </table>
          </div>
        </section>

        <aside className="flex min-h-0 flex-col overflow-hidden rounded-xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)]">
          <div className="flex items-center justify-between border-b border-[var(--color-border-muted)] px-5 py-4"><h2 className="text-base font-semibold text-[var(--color-text-primary)]">操作台</h2><span className="text-xs text-[var(--color-text-muted)]">已选 {selectedProfileIds.length}</span></div>
          <div className="grid grid-cols-3 border-b border-[var(--color-border-muted)] bg-[var(--color-bg-muted)] p-1.5">
            {workspaceTabs.map((tab) => { const Icon = tab.icon; return <button key={tab.id} type="button" onClick={() => setActiveTab(tab.id)} className={'flex h-10 items-center justify-center gap-1.5 rounded-lg text-xs font-medium transition-colors ' + (activeTab === tab.id ? 'bg-[var(--color-bg-surface)] text-[var(--color-text-primary)] shadow-[var(--shadow-xs)]' : 'text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)]')}><Icon className="h-4 w-4" />{tab.label}</button> })}
          </div>
          <div className="min-h-0 flex-1 overflow-y-auto p-5">
            {activeTab === 'window' ? <div className="space-y-5">
              <div className="grid grid-cols-2 gap-2">
                <Button variant="secondary" disabled={!canRun} loading={action === 'window:tile'} onClick={() => void withAction('window:tile', () => runWindowSyncWindowAction(selectedProfileIds, 'tile'), '平铺窗口')}><Grid2X2 className="h-4 w-4" />平铺窗口</Button>
                <Button variant="secondary" disabled={!canRun} loading={action === 'window:cascade'} onClick={() => void withAction('window:cascade', () => runWindowSyncWindowAction(selectedProfileIds, 'cascade'), '层叠窗口')}><Columns className="h-4 w-4" />层叠窗口</Button>
                <Button variant="secondary" disabled={!canRun} loading={action === 'window:maximize'} onClick={() => void withAction('window:maximize', () => runWindowSyncWindowAction(selectedProfileIds, 'maximize'), '最大化窗口')}><Maximize2 className="h-4 w-4" />最大化</Button>
                <Button variant="secondary" disabled={!canRun} loading={action === 'window:minimize'} onClick={() => void withAction('window:minimize', () => runWindowSyncWindowAction(selectedProfileIds, 'minimize'), '最小化窗口')}><Minimize2 className="h-4 w-4" />最小化</Button>
                <Button variant="secondary" className="col-span-2" disabled={!canRun} loading={action === 'window:normal'} onClick={() => void withAction('window:normal', () => runWindowSyncWindowAction(selectedProfileIds, 'normal'), '恢复窗口')}><MonitorDown className="h-4 w-4" />恢复窗口</Button>
              </div>
              <div className="rounded-lg bg-[var(--color-bg-muted)] p-4"><p className="text-sm font-semibold text-[var(--color-text-primary)]">启动同步</p><p className="mt-2 text-xs leading-5 text-[var(--color-text-secondary)]">主窗口的网页点击、输入、滚动和跳转会按当前开关转发给被控窗口。</p><div className="mt-3 space-y-2">{syncOptions.map((option) => <div key={option.key} className="flex items-center justify-between"><span className="text-xs text-[var(--color-text-secondary)]">{option.label}</span><Switch checked={settings[option.key]} disabled={syncState.active} onChange={(checked) => setSettings((current) => ({ ...current, [option.key]: checked }))} /></div>)}</div></div>
            </div> : null}
            {activeTab === 'text' ? <div className="space-y-5">
              <Button variant="secondary" disabled={!canRun} loading={action === 'text:clear'} onClick={() => void withAction('text:clear', () => runWindowSyncTextAction(selectedProfileIds, '', true), '清空焦点内容')}><Trash2 className="h-4 w-4" />清空焦点内容</Button>
              <div><label className="mb-2 block text-xs font-medium text-[var(--color-text-secondary)]">相同文本</label><Textarea rows={8} value={text} onChange={(event) => setText(event.target.value)} placeholder="先在各浏览器中点击输入框，再批量写入文本" /></div>
              <Button className="w-full" disabled={!canRun || !text.trim()} loading={action === 'text:write'} onClick={() => void withAction('text:write', () => runWindowSyncTextAction(selectedProfileIds, text, false), '批量写入文本')}><TextCursorInput className="h-4 w-4" />写入焦点输入框</Button>
              <p className="text-xs leading-5 text-[var(--color-text-muted)]">文本命令写入每个窗口当前聚焦的输入框或可编辑区域；未聚焦输入框的窗口会返回失败状态。</p>
            </div> : null}
            {activeTab === 'tabs' ? <div className="space-y-4">
              <div><label className="mb-2 block text-xs font-medium text-[var(--color-text-secondary)]">目标 URL</label><Input value={targetUrl} onChange={(event) => setTargetUrl(event.target.value)} placeholder="https://example.com" /></div>
              <div className="grid grid-cols-2 gap-2">
                <Button variant="secondary" disabled={!canRun} loading={action === 'tab:new'} onClick={() => void withAction('tab:new', () => runWindowSyncTabAction(selectedProfileIds, 'new', targetUrl), '新建标签页')}><Plus className="h-4 w-4" />新建标签页</Button>
                <Button variant="secondary" disabled={!canRun} loading={action === 'tab:navigate'} onClick={() => void withAction('tab:navigate', () => runWindowSyncTabAction(selectedProfileIds, 'navigate', targetUrl), '导航当前页')}><Navigation className="h-4 w-4" />导航当前页</Button>
                <Button variant="secondary" disabled={!canRun} loading={action === 'tab:reload'} onClick={() => void withAction('tab:reload', () => runWindowSyncTabAction(selectedProfileIds, 'reload'), '刷新页面')}><RotateCcw className="h-4 w-4" />刷新页面</Button>
                <Button variant="secondary" disabled={!canRun} loading={action === 'tab:close'} onClick={() => void withAction('tab:close', () => runWindowSyncTabAction(selectedProfileIds, 'close'), '关闭当前页')}><X className="h-4 w-4" />关闭当前页</Button>
              </div>
              <Button className="w-full" variant="secondary" disabled={!masterId || targetIds.length === 0} loading={action === 'tab:copy'} onClick={() => setConfirmTabSync(true)}><Copy className="h-4 w-4" />从主控同步全部标签页</Button>
              <p className="text-xs leading-5 text-[var(--color-text-muted)]">“从主控同步”会将主窗口普通网页标签复制到被控窗口，并关闭被控窗口现有普通网页标签。</p>
            </div> : null}
          </div>
          <div className="border-t border-[var(--color-border-muted)] px-5 py-3 text-xs text-[var(--color-text-muted)]">{syncState.active ? '同步中 · 已转发 ' + syncState.eventCount + ' 个事件' : '窗口操作可独立使用；同步需设置主窗口与被控窗口。'}</div>
        </aside>
      </div>
      <ConfirmModal open={confirmTabSync} onClose={() => setConfirmTabSync(false)} title="同步主控标签页" content="这会关闭每个被控窗口现有的普通网页标签，并按主窗口当前标签页重新创建。浏览器内置页、扩展页不在本次操作范围内。" confirmText="覆盖并同步" danger onConfirm={() => void withAction('tab:copy', () => copyMasterWindowSyncTabs(masterId, targetIds), '同步主控标签页')} />
    </div>
  )
}
