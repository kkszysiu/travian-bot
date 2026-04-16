<script lang="ts">
  import { GetFarmLists, GetAccountSettings, SaveAccountSettings, ToggleFarmList } from '../../../../wailsjs/go/app/App'
  import { EventsOn } from '../../../../wailsjs/runtime/runtime'
  import RangeInput from '../shared/RangeInput.svelte'
  import { t } from '../../i18n'

  let { accountId }: { accountId: number } = $props()

  let farms = $state<{ id: number; name: string; isActive: boolean }[]>([])
  let settings = $state<Record<string, number>>({})
  let error = $state('')
  let success = $state('')
  let tr = $state<(key: string, params?: Record<string, string | number>) => string>((k) => k)
  t.subscribe(v => tr = v)

  async function loadFarms() {
    try { farms = await GetFarmLists(accountId) || [] } catch (e) { console.error(e) }
  }

  async function loadSettings() {
    try { settings = await GetAccountSettings(accountId) || {} } catch (e) { console.error(e) }
  }

  $effect(() => {
    loadFarms()
    loadSettings()
    const off = EventsOn('farms:modified', () => loadFarms())
    return () => off()
  })

  function get(key: string): number { return settings[key] ?? 0 }
  function set(key: string, val: number) { settings = { ...settings, [key]: val } }

  async function handleToggle(farm: { id: number; isActive: boolean }) {
    try {
      await ToggleFarmList(farm.id, !farm.isActive)
    } catch (e) { console.error(e) }
  }

  async function handleSave() {
    error = ''; success = ''
    try {
      await SaveAccountSettings(accountId, settings)
      success = 'Settings saved'
    } catch (e: any) { error = e?.message || String(e) }
  }
</script>

<div class="flex gap-4 h-full">
  <!-- Farm List -->
  <div class="w-64 border rounded overflow-y-auto">
    <h3 class="text-sm font-medium p-2 border-b bg-gray-50 flex justify-between">
      <span>{tr('farming.farmLists')}</span>
      <span class="text-gray-400 text-xs">{farms.length} {tr('farming.lists')}</span>
    </h3>
    {#each farms as farm (farm.id)}
      <div class="flex items-center px-3 py-2 text-sm border-b border-gray-50 hover:bg-blue-50">
        <input type="checkbox" checked={farm.isActive}
          onchange={() => handleToggle(farm)}
          class="mr-2" />
        <span class:text-green-600={farm.isActive} class:text-gray-400={!farm.isActive}>
          {farm.name}
        </span>
      </div>
    {/each}
    {#if farms.length === 0}
      <p class="p-3 text-xs text-gray-400">{tr('farming.noFarmLists')}</p>
    {/if}
  </div>

  <!-- Farm Settings -->
  <div class="flex-1 space-y-4">
    {#if error}<div class="p-2 bg-red-100 text-red-700 text-sm rounded">{error}</div>{/if}
    {#if success}<div class="p-2 bg-green-100 text-green-700 text-sm rounded">{success}</div>{/if}

    <RangeInput label={tr('farming.farmInterval')} unit={tr('farming.sec')}
      min={get('FarmIntervalMin')} max={get('FarmIntervalMax')}
      onMinChange={(v) => set('FarmIntervalMin', v)} onMaxChange={(v) => set('FarmIntervalMax', v)} />

    <label class="flex items-center gap-2 text-sm">
      <input type="checkbox" checked={get('UseStartAllButton') === 1}
        onchange={(e) => set('UseStartAllButton', (e.target as HTMLInputElement).checked ? 1 : 0)} />
      {tr('farming.useStartAll')}
    </label>

    <div class="flex gap-2">
      <button class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm" onclick={handleSave}>{tr('farming.save')}</button>
    </div>
  </div>
</div>
