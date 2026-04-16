<script lang="ts">
  import { AddAccount, GetAccounts } from '../../../../wailsjs/go/app/App'
  import { accounts } from '../../stores/accounts'
  import type { AccessDetail } from '../../stores/accounts'
  import { t } from '../../i18n'

  let { onDone }: { onDone: () => void } = $props()

  let server = $state('')
  let username = $state('')
  let tr = $state<(key: string, params?: Record<string, string | number>) => string>((k) => k)
  t.subscribe(v => tr = v)

  // Access form fields
  let accUsername = $state('')
  let accPassword = $state('')
  let proxyHost = $state('')
  let proxyPort = $state(0)
  let proxyUsername = $state('')
  let proxyPassword = $state('')

  let accesses = $state<AccessDetail[]>([])
  let selectedAccessIdx = $state<number | null>(null)
  let error = $state('')

  function addAccess() {
    if (!accUsername || !accPassword) return
    accesses = [...accesses, {
      id: 0, username: accUsername, password: accPassword,
      proxyHost, proxyPort, proxyUsername, proxyPassword,
      useragent: '', lastUsed: ''
    }]
    clearAccessForm()
  }

  function editAccess() {
    if (selectedAccessIdx == null) return
    accesses[selectedAccessIdx] = {
      ...accesses[selectedAccessIdx],
      username: accUsername, password: accPassword,
      proxyHost, proxyPort, proxyUsername, proxyPassword
    }
    accesses = [...accesses]
    clearAccessForm()
  }

  function deleteAccess() {
    if (selectedAccessIdx == null) return
    accesses = accesses.filter((_, i) => i !== selectedAccessIdx)
    selectedAccessIdx = null
    clearAccessForm()
  }

  function selectAccess(idx: number) {
    selectedAccessIdx = idx
    const a = accesses[idx]
    accUsername = a.username
    accPassword = a.password
    proxyHost = a.proxyHost
    proxyPort = a.proxyPort
    proxyUsername = a.proxyUsername
    proxyPassword = a.proxyPassword
  }

  function clearAccessForm() {
    accUsername = ''; accPassword = ''
    proxyHost = ''; proxyPort = 0
    proxyUsername = ''; proxyPassword = ''
    selectedAccessIdx = null
  }

  async function handleAdd() {
    error = ''
    if (!server || !username) {
      error = tr('addAccount.serverRequired')
      return
    }
    if (accesses.length === 0) {
      error = tr('addAccount.credentialsRequired')
      return
    }
    try {
      await AddAccount({ id: 0, username, server, accesses: $state.snapshot(accesses) })
      const list = await GetAccounts()
      accounts.set(list || [])
      onDone()
    } catch (e: any) {
      error = e?.message || String(e)
    }
  }
</script>

<div class="max-w-2xl mx-auto space-y-4">
  <h2 class="text-lg font-semibold">{tr('addAccount.title')}</h2>

  {#if error}
    <div class="p-2 bg-red-100 text-red-700 text-sm rounded">{error}</div>
  {/if}

  <div class="grid grid-cols-2 gap-3">
    <label class="block">
      <span class="text-sm text-gray-600">{tr('addAccount.serverUrl')}</span>
      <input type="text" bind:value={server} placeholder={tr('addAccount.serverPlaceholder')}
        oninput={() => error = ''}
        class="w-full mt-1 px-3 py-2 border rounded text-sm" />
    </label>
    <label class="block">
      <span class="text-sm text-gray-600">{tr('addAccount.nickname')}</span>
      <input type="text" bind:value={username} placeholder={tr('addAccount.nicknamePlaceholder')}
        oninput={() => error = ''}
        class="w-full mt-1 px-3 py-2 border rounded text-sm" />
    </label>
  </div>

  <div class="border rounded p-3 space-y-3">
    <h3 class="text-sm font-medium text-gray-700">{tr('addAccount.credentials')}</h3>
    <div class="grid grid-cols-2 gap-3">
      <label class="block">
        <span class="text-xs text-gray-500">{tr('addAccount.username')}</span>
        <input type="text" bind:value={accUsername} class="w-full mt-1 px-2 py-1 border rounded text-sm" />
      </label>
      <label class="block">
        <span class="text-xs text-gray-500">{tr('addAccount.password')}</span>
        <input type="password" bind:value={accPassword} class="w-full mt-1 px-2 py-1 border rounded text-sm" />
      </label>
      <label class="block">
        <span class="text-xs text-gray-500">{tr('addAccount.proxyHost')}</span>
        <input type="text" bind:value={proxyHost} class="w-full mt-1 px-2 py-1 border rounded text-sm" />
      </label>
      <label class="block">
        <span class="text-xs text-gray-500">{tr('addAccount.proxyPort')}</span>
        <input type="number" bind:value={proxyPort} class="w-full mt-1 px-2 py-1 border rounded text-sm" />
      </label>
      <label class="block">
        <span class="text-xs text-gray-500">{tr('addAccount.proxyUsername')}</span>
        <input type="text" bind:value={proxyUsername} class="w-full mt-1 px-2 py-1 border rounded text-sm" />
      </label>
      <label class="block">
        <span class="text-xs text-gray-500">{tr('addAccount.proxyPassword')}</span>
        <input type="password" bind:value={proxyPassword} class="w-full mt-1 px-2 py-1 border rounded text-sm" />
      </label>
    </div>
    <div class="flex gap-2">
      <button class="px-3 py-1 text-xs bg-green-500 text-white rounded hover:bg-green-600" onclick={addAccess}>{tr('addAccount.add')}</button>
      <button class="px-3 py-1 text-xs bg-blue-500 text-white rounded hover:bg-blue-600" onclick={editAccess}
        disabled={selectedAccessIdx == null}>{tr('addAccount.edit')}</button>
      <button class="px-3 py-1 text-xs bg-red-500 text-white rounded hover:bg-red-600" onclick={deleteAccess}
        disabled={selectedAccessIdx == null}>{tr('addAccount.delete')}</button>
    </div>

    {#if accesses.length > 0}
      <table class="w-full text-xs">
        <thead>
          <tr class="border-b">
            <th class="text-left p-1">{tr('addAccount.username')}</th>
            <th class="text-left p-1">{tr('addAccount.proxy')}</th>
          </tr>
        </thead>
        <tbody>
          {#each accesses as access, idx}
            <tr class="border-b cursor-pointer hover:bg-gray-50"
              class:bg-blue-50={selectedAccessIdx === idx}
              onclick={() => selectAccess(idx)}>
              <td class="p-1">{access.username}</td>
              <td class="p-1 text-gray-400">{access.proxyHost || tr('addAccount.none')}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    {/if}
  </div>

  <div class="flex gap-2">
    <button class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm" onclick={handleAdd}>
      {tr('addAccount.addButton')}
    </button>
    <button class="px-4 py-2 bg-gray-300 text-gray-700 rounded hover:bg-gray-400 text-sm" onclick={onDone}>
      {tr('addAccount.cancel')}
    </button>
  </div>
</div>
