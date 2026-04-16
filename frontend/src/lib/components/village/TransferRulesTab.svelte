<script lang="ts">
  import { GetVillageSettings, SaveVillageSettings } from '../../../../wailsjs/go/app/App'
  import RangeInput from '../shared/RangeInput.svelte'
  import { t } from '../../i18n'

  let { villageId, accountId }: { villageId: number; accountId: number } = $props()

  let settings = $state<Record<string, number>>({})
  let error = $state('')
  let success = $state('')
  let tr = $state<(key: string, params?: Record<string, string | number>) => string>((k) => k)
  t.subscribe(v => tr = v)

  async function loadSettings() {
    try { settings = await GetVillageSettings(villageId) || {} } catch (e) { console.error(e) }
  }

  $effect(() => { loadSettings() })

  function get(key: string): number { return settings[key] ?? 0 }
  function set(key: string, val: number) { settings = { ...settings, [key]: val } }

  async function handleSave() {
    error = ''; success = ''
    try {
      await SaveVillageSettings(villageId, settings)
      success = 'Settings saved'
    } catch (e: any) { error = e?.message || String(e) }
  }
</script>

<div class="max-w-2xl space-y-4 overflow-y-auto">
  {#if error}<div class="p-2 bg-red-100 text-red-700 text-sm rounded">{error}</div>{/if}
  {#if success}<div class="p-2 bg-green-100 text-green-700 text-sm rounded">{success}</div>{/if}

  <div class="border rounded p-3 space-y-2">
    <label class="flex items-center gap-2 text-sm">
      <input type="checkbox" checked={get('AutoSendResourceEnable') === 1}
        onchange={(e) => set('AutoSendResourceEnable', (e.target as HTMLInputElement).checked ? 1 : 0)} />
      {tr('transferRules.enable')}
    </label>
    <RangeInput label={tr('transferRules.repeatTime')} unit={tr('transferRules.sec')}
      min={get('AutoSendResourceRepeatMin')} max={get('AutoSendResourceRepeatMax')}
      onMinChange={(v) => set('AutoSendResourceRepeatMin', v)} onMaxChange={(v) => set('AutoSendResourceRepeatMax', v)} />
    <label class="flex items-center gap-2 text-sm">
      <span class="text-xs text-gray-500">{tr('transferRules.threshold')}</span>
      <input type="number" class="w-16 px-2 py-1 border rounded text-sm" min="0" max="100"
        value={get('AutoSendResourceThreshold')} onchange={(e) => set('AutoSendResourceThreshold', parseInt((e.target as HTMLInputElement).value))} />
      <span class="text-xs text-gray-400">%</span>
    </label>
    <p class="text-xs text-gray-400">{tr('transferRules.description')}</p>
    <button class="px-4 py-1.5 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm" onclick={handleSave}>
      {tr('transferRules.save')}
    </button>
  </div>
</div>
