<script lang="ts">
  import { GetAccountDetail, UpdateAccount } from '../../../../wailsjs/go/app/App'
  import type { AccessDetail } from '../../stores/accounts'
  import { t } from '../../i18n'

  let { accountId }: { accountId: number } = $props()

  let server = $state('')
  let username = $state('')
  let accUsername = $state('')
  let accPassword = $state('')
  let proxyHost = $state('')
  let proxyPort = $state(0)
  let proxyUsername = $state('')
  let proxyPassword = $state('')
  let useragent = $state('')
  let accesses = $state<AccessDetail[]>([])
  let selectedAccessIdx = $state<number | null>(null)
  let error = $state('')
  let success = $state('')
  let tr = $state<(key: string, params?: Record<string, string | number>) => string>((k) => k)
  t.subscribe(v => tr = v)

  async function load() {
    try {
      const detail = await GetAccountDetail(accountId)
      if (detail) {
        server = detail.server
        username = detail.username
        accesses = detail.accesses || []
      }
    } catch (e) { console.error(e) }
  }

  $effect(() => { load() })

  function addAccess() {
    if (!accUsername || !accPassword) return
    accesses = [...accesses, {
      id: 0, username: accUsername, password: accPassword,
      proxyHost, proxyPort, proxyUsername, proxyPassword,
      useragent, lastUsed: ''
    }]
    clearForm()
  }

  function editAccess() {
    if (selectedAccessIdx == null) return
    accesses[selectedAccessIdx] = {
      ...accesses[selectedAccessIdx],
      username: accUsername, password: accPassword,
      proxyHost, proxyPort, proxyUsername, proxyPassword, useragent
    }
    accesses = [...accesses]
    clearForm()
  }

  function deleteAccess() {
    if (selectedAccessIdx == null) return
    accesses = accesses.filter((_, i) => i !== selectedAccessIdx)
    selectedAccessIdx = null
    clearForm()
  }

  function selectAccess(idx: number) {
    selectedAccessIdx = idx
    const a = accesses[idx]
    accUsername = a.username; accPassword = a.password
    proxyHost = a.proxyHost; proxyPort = a.proxyPort
    proxyUsername = a.proxyUsername; proxyPassword = a.proxyPassword
    useragent = a.useragent
  }

  function clearForm() {
    accUsername = ''; accPassword = ''
    proxyHost = ''; proxyPort = 0
    proxyUsername = ''; proxyPassword = ''
    useragent = ''
    selectedAccessIdx = null
  }

  async function handleSave() {
    error = ''; success = ''
    try {
      await UpdateAccount({ id: accountId, username, server, accesses: $state.snapshot(accesses) })
      success = tr('editAccount.updated')
    } catch (e: any) {
      error = e?.message || String(e)
    }
  }
</script>

<div class="max-w-2xl mx-auto space-y-4">
  <h2 class="text-lg font-semibold">{tr('editAccount.title')}</h2>

  {#if error}<div class="p-2 bg-red-100 text-red-700 text-sm rounded">{error}</div>{/if}
  {#if success}<div class="p-2 bg-green-100 text-green-700 text-sm rounded">{success}</div>{/if}

  <div class="grid grid-cols-2 gap-3">
    <label class="block">
      <span class="text-sm text-gray-600">{tr('editAccount.serverUrl')}</span>
      <input type="text" bind:value={server} class="w-full mt-1 px-3 py-2 border rounded text-sm" />
    </label>
    <label class="block">
      <span class="text-sm text-gray-600">{tr('editAccount.nickname')}</span>
      <input type="text" bind:value={username} class="w-full mt-1 px-3 py-2 border rounded text-sm" />
    </label>
  </div>

  <div class="border rounded p-3 space-y-3">
    <h3 class="text-sm font-medium text-gray-700">{tr('editAccount.credentials')}</h3>
    <div class="grid grid-cols-2 gap-3">
      <label class="block"><span class="text-xs text-gray-500">{tr('editAccount.username')}</span>
        <input type="text" bind:value={accUsername} class="w-full mt-1 px-2 py-1 border rounded text-sm" /></label>
      <label class="block"><span class="text-xs text-gray-500">{tr('editAccount.password')}</span>
        <input type="password" bind:value={accPassword} class="w-full mt-1 px-2 py-1 border rounded text-sm" /></label>
      <label class="block"><span class="text-xs text-gray-500">{tr('editAccount.proxyHost')}</span>
        <input type="text" bind:value={proxyHost} class="w-full mt-1 px-2 py-1 border rounded text-sm" /></label>
      <label class="block"><span class="text-xs text-gray-500">{tr('editAccount.proxyPort')}</span>
        <input type="number" bind:value={proxyPort} class="w-full mt-1 px-2 py-1 border rounded text-sm" /></label>
      <label class="block"><span class="text-xs text-gray-500">{tr('editAccount.proxyUsername')}</span>
        <input type="text" bind:value={proxyUsername} class="w-full mt-1 px-2 py-1 border rounded text-sm" /></label>
      <label class="block"><span class="text-xs text-gray-500">{tr('editAccount.proxyPassword')}</span>
        <input type="password" bind:value={proxyPassword} class="w-full mt-1 px-2 py-1 border rounded text-sm" /></label>
      <label class="block col-span-2"><span class="text-xs text-gray-500">{tr('editAccount.userAgent')}</span>
        <input type="text" bind:value={useragent} class="w-full mt-1 px-2 py-1 border rounded text-sm" /></label>
    </div>
    <div class="flex gap-2">
      <button class="px-3 py-1 text-xs bg-green-500 text-white rounded hover:bg-green-600" onclick={addAccess}>{tr('editAccount.add')}</button>
      <button class="px-3 py-1 text-xs bg-blue-500 text-white rounded hover:bg-blue-600" onclick={editAccess} disabled={selectedAccessIdx == null}>{tr('editAccount.edit')}</button>
      <button class="px-3 py-1 text-xs bg-red-500 text-white rounded hover:bg-red-600" onclick={deleteAccess} disabled={selectedAccessIdx == null}>{tr('editAccount.delete')}</button>
    </div>
    {#if accesses.length > 0}
      <table class="w-full text-xs">
        <thead><tr class="border-b"><th class="text-left p-1">{tr('editAccount.username')}</th><th class="text-left p-1">{tr('editAccount.proxy')}</th><th class="text-left p-1">{tr('editAccount.userAgent')}</th><th class="text-left p-1">{tr('editAccount.lastUsed')}</th></tr></thead>
        <tbody>
          {#each accesses as access, idx}
            <tr class="border-b cursor-pointer hover:bg-gray-50" class:bg-blue-50={selectedAccessIdx === idx} onclick={() => selectAccess(idx)}>
              <td class="p-1">{access.username}</td>
              <td class="p-1 text-gray-400">{access.proxyHost || tr('editAccount.none')}</td>
              <td class="p-1 text-gray-400 truncate max-w-32">{access.useragent || tr('editAccount.dash')}</td>
              <td class="p-1 text-gray-400">{access.lastUsed || tr('editAccount.dash')}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    {/if}
  </div>

  <button class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm" onclick={handleSave}>
    {tr('editAccount.save')}
  </button>
</div>
