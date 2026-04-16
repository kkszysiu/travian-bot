<script lang="ts">
  import { GetTasks, GetLogs, GetAccountDetail, GetStatus } from '../../../../wailsjs/go/app/App'
  import { EventsOn } from '../../../../wailsjs/runtime/runtime'
  import { statusLabels } from '../../stores/status'
  import { t } from '../../i18n'

  let { accountId }: { accountId: number } = $props()

  let tasks = $state<{ task: string; executeAt: string; stage: string }[]>([])
  let logText = $state('')
  let logEl: HTMLTextAreaElement | undefined = $state()
  let tr = $state<(key: string, params?: Record<string, string | number>) => string>((k) => k)
  t.subscribe(v => tr = v)

  // Account info
  let accountServer = $state('')
  let accountUsername = $state('')
  let accessCount = $state(0)
  let currentStatus = $state(0)
  let statusColor = $state('black')

  const statusColorMap: Record<number, string> = {
    0: 'bg-gray-400',
    1: 'bg-orange-400',
    2: 'bg-green-500',
    3: 'bg-orange-400',
    4: 'bg-red-500',
    5: 'bg-orange-400',
  }

  async function loadInfo() {
    try {
      const detail = await GetAccountDetail(accountId)
      if (detail) {
        accountServer = detail.server
        accountUsername = detail.username
        accessCount = detail.accesses?.length ?? 0
      }
      currentStatus = await GetStatus(accountId) ?? 0
    } catch (e) { console.error(e) }
  }

  async function load() {
    try {
      tasks = await GetTasks(accountId) || []
      const logs = await GetLogs(accountId) || []
      logText = logs.map(l => `[${l.time}] ${l.level}: ${l.message}`).join('\n')
      scrollToBottom()
    } catch (e) { console.error(e) }
  }

  function scrollToBottom() {
    requestAnimationFrame(() => {
      if (logEl) {
        logEl.scrollTop = logEl.scrollHeight
      }
    })
  }

  function clearLogs() {
    logText = ''
  }

  $effect(() => {
    loadInfo()
    load()
    const offTasks = EventsOn('tasks:modified', (data: { accountId?: number }) => {
      if (!data || !data.accountId || data.accountId === accountId) {
        load()
      }
    })
    const offLogs = EventsOn('log:emitted', (data: { accountId: number; message: string; level: string }) => {
      if (data.accountId === accountId) {
        const prefix = data.level === 'error' ? '!!' : data.level === 'warn' ? '**' : ''
        logText += `\n[${new Date().toLocaleTimeString()}] ${data.level}: ${prefix ? prefix + ' ' : ''}${data.message}`
        scrollToBottom()
      }
    })
    const offStatus = EventsOn('status:modified', (data: { accountId: number; status: number; color: string }) => {
      if (data.accountId === accountId) {
        currentStatus = data.status
        statusColor = data.color
      }
    })
    return () => { offTasks(); offLogs(); offStatus() }
  })
</script>

<div class="flex flex-col h-full gap-4">
  <!-- Account Info & Status -->
  <div class="border rounded p-3 bg-gray-50">
    <div class="flex items-center justify-between mb-2">
      <h3 class="text-sm font-medium">{tr('debug.accountInfo')}</h3>
      <div class="flex items-center gap-2">
        <span class="inline-block w-3 h-3 rounded-full {statusColorMap[currentStatus] ?? 'bg-gray-400'}"></span>
        <span class="text-sm font-medium">{statusLabels[currentStatus] ?? 'Unknown'}</span>
      </div>
    </div>
    <div class="grid grid-cols-3 gap-2 text-xs">
      <div>
        <span class="text-gray-500">{tr('debug.username')}</span>
        <span class="ml-1 font-mono">{accountUsername || tr('debug.dash')}</span>
      </div>
      <div>
        <span class="text-gray-500">{tr('debug.server')}</span>
        <span class="ml-1 font-mono">{accountServer || tr('debug.dash')}</span>
      </div>
      <div>
        <span class="text-gray-500">{tr('debug.credentials')}</span>
        <span class="ml-1">{accessCount}</span>
      </div>
    </div>
  </div>

  <!-- Tasks Table -->
  <div class="border rounded overflow-auto max-h-64">
    <div class="flex items-center justify-between px-3 py-2 bg-gray-50 border-b sticky top-0">
      <h3 class="text-sm font-medium">{tr('debug.tasks')} ({tasks.length})</h3>
      <button
        class="px-2 py-1 text-xs bg-gray-200 text-gray-700 rounded hover:bg-gray-300"
        onclick={() => { load(); loadInfo() }}
      >
        {tr('debug.refresh')}
      </button>
    </div>
    <table class="w-full text-sm">
      <thead class="bg-gray-50 sticky top-[41px]">
        <tr>
          <th class="text-left px-3 py-2">{tr('debug.task')}</th>
          <th class="text-left px-3 py-2">{tr('debug.executeAt')}</th>
          <th class="text-left px-3 py-2">{tr('debug.stage')}</th>
        </tr>
      </thead>
      <tbody>
        {#each tasks as task}
          <tr class="border-t">
            <td class="px-3 py-1">{task.task}</td>
            <td class="px-3 py-1 text-gray-500">{task.executeAt}</td>
            <td class="px-3 py-1 text-gray-500">{task.stage}</td>
          </tr>
        {/each}
        {#if tasks.length === 0}
          <tr><td colspan="3" class="px-3 py-4 text-gray-400 text-center">{tr('debug.noTasks')}</td></tr>
        {/if}
      </tbody>
    </table>
  </div>

  <!-- Log Viewer -->
  <div class="flex-1 border rounded overflow-hidden flex flex-col">
    <div class="flex items-center justify-between p-2 border-b bg-gray-50">
      <h3 class="text-sm font-medium">{tr('debug.logs')}</h3>
      <button
        class="px-2 py-1 text-xs bg-gray-200 text-gray-700 rounded hover:bg-gray-300"
        onclick={clearLogs}
      >
        {tr('debug.clearLogs')}
      </button>
    </div>
    <textarea bind:this={logEl} readonly class="flex-1 p-2 text-xs font-mono bg-gray-900 text-green-400 resize-none" value={logText}></textarea>
  </div>
</div>
