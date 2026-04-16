<script lang="ts">
  import { AddAccount, GetAccounts } from '../../../../wailsjs/go/app/App'
  import { accounts } from '../../stores/accounts'
  import { t } from '../../i18n'

  let { onDone }: { onDone: () => void } = $props()

  let input = $state('')
  let error = $state('')
  let tr = $state<(key: string, params?: Record<string, string | number>) => string>((k) => k)
  t.subscribe(v => tr = v)
  let parsed = $state<{
    server: string
    username: string
    password: string
    proxyHost: string
    proxyPort: number
    proxyUsername: string
    proxyPassword: string
  }[]>([])

  function parseInput() {
    parsed = input.split('\n').filter(l => l.trim()).map(line => {
      const parts = line.trim().split(/\s+/)
      const server = parts[0] || ''
      const username = parts[1] || ''
      const password = parts[2] || ''

      // 4th field: proxy as host:port
      let proxyHost = ''
      let proxyPort = 0
      if (parts[3]) {
        const colonIdx = parts[3].lastIndexOf(':')
        if (colonIdx > 0) {
          proxyHost = parts[3].substring(0, colonIdx)
          proxyPort = parseInt(parts[3].substring(colonIdx + 1), 10) || 0
        } else {
          proxyHost = parts[3]
        }
      }

      // 5th field: proxy username, 6th field: proxy password
      const proxyUsername = parts[4] || ''
      const proxyPassword = parts[5] || ''

      return { server, username, password, proxyHost, proxyPort, proxyUsername, proxyPassword }
    })
  }

  async function handleAdd() {
    if (parsed.length === 0) {
      error = tr('addAccounts.noAccounts')
      return
    }
    try {
      for (const p of parsed) {
        await AddAccount({
          id: 0,
          username: p.username,
          server: p.server,
          accesses: [{
            id: 0,
            username: p.username,
            password: p.password,
            proxyHost: p.proxyHost,
            proxyPort: p.proxyPort,
            proxyUsername: p.proxyUsername,
            proxyPassword: p.proxyPassword,
            useragent: '', lastUsed: ''
          }]
        })
      }
      const list = await GetAccounts()
      accounts.set(list || [])
      onDone()
    } catch (e: any) {
      error = e?.message || String(e)
    }
  }
</script>

<div class="max-w-2xl mx-auto space-y-4">
  <h2 class="text-lg font-semibold">{tr('addAccounts.title')}</h2>

  {#if error}
    <div class="p-2 bg-red-100 text-red-700 text-sm rounded">{error}</div>
  {/if}

  <p class="text-xs text-gray-500">
    {tr('addAccounts.format')}
  </p>

  <textarea bind:value={input} oninput={parseInput} rows="8"
    class="w-full px-3 py-2 border rounded text-sm font-mono" placeholder={tr('addAccounts.placeholder')}></textarea>

  {#if parsed.length > 0}
    <table class="w-full text-xs">
      <thead><tr class="border-b"><th class="text-left p-1">{tr('addAccounts.server')}</th><th class="text-left p-1">{tr('addAccounts.username')}</th><th class="text-left p-1">{tr('addAccounts.proxy')}</th><th class="text-left p-1">{tr('addAccounts.proxyAuth')}</th></tr></thead>
      <tbody>
        {#each parsed as p}
          <tr class="border-b">
            <td class="p-1">{p.server}</td>
            <td class="p-1">{p.username}</td>
            <td class="p-1 text-gray-400">{p.proxyHost ? `${p.proxyHost}:${p.proxyPort}` : tr('addAccounts.none')}</td>
            <td class="p-1 text-gray-400">{p.proxyUsername ? `${p.proxyUsername}` : ''}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  {/if}

  <div class="flex gap-2">
    <button class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm" onclick={handleAdd}>
      {tr('addAccounts.addButton', { count: parsed.length })}
    </button>
    <button class="px-4 py-2 bg-gray-300 text-gray-700 rounded hover:bg-gray-400 text-sm" onclick={onDone}>{tr('addAccounts.cancel')}</button>
  </div>
</div>
