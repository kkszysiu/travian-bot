<script lang="ts">
  import { accounts, selectedAccountId, type AccountListItem } from '../../stores/accounts'
  import { statuses, statusLabels } from '../../stores/status'
  import { showOverlay, hideOverlay } from '../../stores/waiting'
  import { GetAccounts, Login, Logout, Pause, Restart, DeleteAccount, LogForAccount } from '../../../../wailsjs/go/app/App'
  import { EventsOn } from '../../../../wailsjs/runtime/runtime'
  import { t, locale, localeNames, type Locale } from '../../i18n'

  let activeTab = $state<string>('none')
  let accountList = $state<AccountListItem[]>([])
  let selectedId = $state<number | null>(null)
  let statusMap: Record<number, { status: number; color: string }> = $state({})
  let tr = $state<(key: string, params?: Record<string, string | number>) => string>((k) => k)

  accounts.subscribe(v => accountList = v)
  selectedAccountId.subscribe(v => selectedId = v)
  statuses.subscribe(v => statusMap = v)
  t.subscribe(v => tr = v)

  export function setActiveTab(tab: string) {
    activeTab = tab
  }

  let { onTabChange }: { onTabChange: (tab: string) => void } = $props()

  async function loadAccounts() {
    try {
      const list = await GetAccounts()
      accounts.set(list || [])
    } catch (e) {
      console.error('Failed to load accounts:', e)
    }
  }

  function selectAccount(id: number) {
    selectedAccountId.set(id)
    onTabChange('villages')
  }

  async function handleLogin() {
    if (selectedId == null) return
    showOverlay(tr('overlay.loggingIn'))
    try {
      await Login(selectedId)
    } catch (e: any) {
      const msg = e?.message || String(e)
      console.error('Login failed:', msg)
      await LogForAccount(selectedId, 'error', `Login failed: ${msg}`).catch(() => {})
    } finally {
      hideOverlay()
    }
  }

  async function handleLogout() {
    if (selectedId == null) return
    showOverlay(tr('overlay.loggingOut'))
    try {
      await Logout(selectedId)
    } catch (e: any) {
      const msg = e?.message || String(e)
      console.error('Logout failed:', msg)
      await LogForAccount(selectedId, 'error', `Logout failed: ${msg}`).catch(() => {})
    } finally {
      hideOverlay()
    }
  }

  async function handlePause() {
    if (selectedId == null) return
    const s = statusMap[selectedId]
    if (s && s.status === 4) {
      try { await Restart(selectedId) } catch (e) { console.error(e) }
    } else {
      try { await Pause(selectedId) } catch (e) { console.error(e) }
    }
  }

  let confirmingDelete = $state(false)
  let deleteTimer: ReturnType<typeof setTimeout> | null = null

  function handleDelete() {
    if (selectedId == null) return
    if (!confirmingDelete) {
      confirmingDelete = true
      deleteTimer = setTimeout(() => { confirmingDelete = false }, 3000)
      return
    }
    if (deleteTimer) clearTimeout(deleteTimer)
    confirmingDelete = false
    doDelete()
  }

  async function doDelete() {
    if (selectedId == null) return
    try {
      await DeleteAccount(selectedId)
      selectedAccountId.set(null)
      onTabChange('none')
      await loadAccounts()
    } catch (e) { console.error(e) }
  }

  function getStatusColor(id: number): string {
    return statusMap[id]?.color ?? 'black'
  }

  function getPauseLabel(): string {
    if (selectedId == null) return tr('sidebar.pause')
    const s = statusMap[selectedId]
    return (s && s.status === 4) ? tr('sidebar.restart') : tr('sidebar.pause')
  }

  let currentLocale = $state<Locale>('en')
  locale.subscribe(v => currentLocale = v)

  function switchLocale(e: Event) {
    const val = (e.target as HTMLSelectElement).value as Locale
    locale.set(val)
  }

  // Load accounts on mount and listen for changes
  $effect(() => {
    loadAccounts()
    EventsOn('accounts:modified', () => loadAccounts())
    EventsOn('status:modified', (data: { accountId: number; status: number; color: string }) => {
      statuses.update(s => ({ ...s, [data.accountId]: { status: data.status, color: data.color } }))
    })
  })
</script>

<div class="flex flex-col h-full bg-gray-50 border-r border-gray-200">
  <!-- Action Buttons -->
  <div class="p-2 space-y-1">
    <div class="flex gap-1">
      <button class="flex-1 px-2 py-1 text-xs bg-blue-500 text-white rounded hover:bg-blue-600"
        onclick={() => onTabChange('addAccount')}>
        {tr('sidebar.addAccount')}
      </button>
      <button class="flex-1 px-2 py-1 text-xs bg-blue-500 text-white rounded hover:bg-blue-600"
        onclick={() => onTabChange('addAccounts')}>
        {tr('sidebar.addAccounts')}
      </button>
    </div>
    <div class="flex gap-1">
      <button class="flex-1 px-2 py-1 text-xs bg-green-500 text-white rounded hover:bg-green-600"
        onclick={handleLogin} disabled={selectedId == null}>
        {tr('sidebar.login')}
      </button>
      <button class="flex-1 px-2 py-1 text-xs bg-red-500 text-white rounded hover:bg-red-600"
        onclick={handleLogout} disabled={selectedId == null}>
        {tr('sidebar.logout')}
      </button>
    </div>
    <div class="flex gap-1">
      <button class="flex-1 px-2 py-1 text-xs text-white rounded {confirmingDelete ? 'bg-orange-500 hover:bg-orange-600' : 'bg-gray-500 hover:bg-gray-600'}"
        onclick={handleDelete} disabled={selectedId == null}>
        {confirmingDelete ? tr('sidebar.confirmDelete') : tr('sidebar.delete')}
      </button>
      <button class="flex-1 px-2 py-1 text-xs bg-yellow-500 text-white rounded hover:bg-yellow-600"
        onclick={handlePause} disabled={selectedId == null}>
        {getPauseLabel()}
      </button>
    </div>
  </div>

  <!-- Account List -->
  <div class="flex-1 overflow-y-auto">
    {#each accountList as account (account.id)}
      <button
        class="w-full text-left px-3 py-2 text-sm border-b border-gray-100 hover:bg-blue-50 transition-colors"
        class:bg-blue-100={selectedId === account.id}
        onclick={() => selectAccount(account.id)}
      >
        <span style="color: {getStatusColor(account.id)}">{account.username}</span>
        <span class="text-xs text-gray-400 block truncate">{account.server}</span>
      </button>
    {/each}
    {#if accountList.length === 0}
      <p class="p-4 text-xs text-gray-400 text-center">{tr('sidebar.noAccounts')}</p>
    {/if}
  </div>

  <!-- Footer: Version + Language -->
  <div class="p-2 border-t border-gray-200 space-y-1">
    <div class="text-xs text-gray-400 text-center">{tr('app.title')}</div>
    <select
      class="w-full px-1 py-0.5 text-xs border rounded bg-white text-gray-600"
      value={currentLocale}
      onchange={switchLocale}
    >
      {#each Object.entries(localeNames) as [code, name]}
        <option value={code}>{name}</option>
      {/each}
    </select>
  </div>
</div>
