<script lang="ts">
  import { GetAccountSettings, SaveAccountSettings } from '../../../../wailsjs/go/app/App'
  import RangeInput from '../shared/RangeInput.svelte'
  import { t } from '../../i18n'

  let { accountId }: { accountId: number } = $props()

  let settings = $state<Record<string, number>>({})
  let error = $state('')
  let success = $state('')
  let tr = $state<(key: string, params?: Record<string, string | number>) => string>((k) => k)
  t.subscribe(v => tr = v)

  async function load() {
    try {
      settings = await GetAccountSettings(accountId) || {}
    } catch (e) { console.error(e) }
  }

  $effect(() => { load() })

  async function handleSave() {
    error = ''; success = ''
    try {
      await SaveAccountSettings(accountId, settings)
      success = 'Settings saved'
    } catch (e: any) { error = e?.message || String(e) }
  }

  function get(key: string): number { return settings[key] ?? 0 }
  function set(key: string, val: number) { settings = { ...settings, [key]: val } }
</script>

<div class="max-w-2xl mx-auto space-y-6">
  <h2 class="text-lg font-semibold">{tr('accountSettings.title')}</h2>

  {#if error}<div class="p-2 bg-red-100 text-red-700 text-sm rounded">{error}</div>{/if}
  {#if success}<div class="p-2 bg-green-100 text-green-700 text-sm rounded">{success}</div>{/if}

  <!-- Tribe -->
  <div class="border rounded p-3">
    <h3 class="text-sm font-medium text-gray-700 mb-2">{tr('accountSettings.accountInfo')}</h3>
    <label class="block">
      <span class="text-xs text-gray-500">{tr('accountSettings.tribe')}</span>
      <select class="w-full mt-1 px-2 py-1 border rounded text-sm"
        value={get('Tribe')}
        onchange={(e) => set('Tribe', parseInt((e.target as HTMLSelectElement).value))}>
        <option value={0}>{tr('accountSettings.tribes.any')}</option>
        <option value={1}>{tr('accountSettings.tribes.romans')}</option>
        <option value={2}>{tr('accountSettings.tribes.teutons')}</option>
        <option value={3}>{tr('accountSettings.tribes.gauls')}</option>
        <option value={6}>{tr('accountSettings.tribes.egyptians')}</option>
        <option value={7}>{tr('accountSettings.tribes.huns')}</option>
      </select>
    </label>
  </div>

  <!-- Work Window -->
  <div class="border rounded p-3 space-y-3">
    <h3 class="text-sm font-medium text-gray-700">{tr('accountSettings.workWindow')}</h3>
    <div class="flex items-center gap-2">
      <span class="text-xs text-gray-500 w-16">{tr('accountSettings.workStart')}</span>
      <input type="number" min={0} max={23} class="w-16 px-2 py-1 border rounded text-sm text-center"
        value={get('WorkStartHour')}
        onchange={(e) => set('WorkStartHour', parseInt((e.target as HTMLInputElement).value) || 0)} />
      <span class="text-sm font-medium">:</span>
      <input type="number" min={0} max={59} class="w-16 px-2 py-1 border rounded text-sm text-center"
        value={get('WorkStartMinute')}
        onchange={(e) => set('WorkStartMinute', parseInt((e.target as HTMLInputElement).value) || 0)} />
    </div>
    <div class="flex items-center gap-2">
      <span class="text-xs text-gray-500 w-16">{tr('accountSettings.workEnd')}</span>
      <input type="number" min={0} max={23} class="w-16 px-2 py-1 border rounded text-sm text-center"
        value={get('WorkEndHour')}
        onchange={(e) => set('WorkEndHour', parseInt((e.target as HTMLInputElement).value) || 0)} />
      <span class="text-sm font-medium">:</span>
      <input type="number" min={0} max={59} class="w-16 px-2 py-1 border rounded text-sm text-center"
        value={get('WorkEndMinute')}
        onchange={(e) => set('WorkEndMinute', parseInt((e.target as HTMLInputElement).value) || 0)} />
    </div>
    <div class="flex items-center gap-2">
      <span class="text-xs text-gray-500 w-16">{tr('accountSettings.sleepJitter')}</span>
      <input type="number" min={0} class="w-20 px-2 py-1 border rounded text-sm text-center"
        value={get('SleepRandomMinute')}
        onchange={(e) => set('SleepRandomMinute', parseInt((e.target as HTMLInputElement).value) || 0)} />
      <span class="text-xs text-gray-500">{tr('accountSettings.min')}</span>
    </div>
    <RangeInput label={tr('accountSettings.sleepTime')} unit={tr('accountSettings.min')}
      min={get('SleepTimeMin')} max={get('SleepTimeMax')}
      onMinChange={(v) => set('SleepTimeMin', v)} onMaxChange={(v) => set('SleepTimeMax', v)} />
  </div>

  <!-- Delays -->
  <div class="border rounded p-3 space-y-2">
    <h3 class="text-sm font-medium text-gray-700">{tr('accountSettings.delay')}</h3>
    <RangeInput label={tr('accountSettings.clickDelay')} unit={tr('accountSettings.ms')}
      min={get('ClickDelayMin')} max={get('ClickDelayMax')}
      onMinChange={(v) => set('ClickDelayMin', v)} onMaxChange={(v) => set('ClickDelayMax', v)} />
    <RangeInput label={tr('accountSettings.taskDelay')} unit={tr('accountSettings.ms')}
      min={get('TaskDelayMin')} max={get('TaskDelayMax')}
      onMinChange={(v) => set('TaskDelayMin', v)} onMaxChange={(v) => set('TaskDelayMax', v)} />
  </div>

  <!-- Features -->
  <div class="border rounded p-3 space-y-2">
    <h3 class="text-sm font-medium text-gray-700">{tr('accountSettings.features')}</h3>
    <label class="flex items-center gap-2 text-sm">
      <input type="checkbox" checked={get('EnableAutoLoadVillageBuilding') === 1}
        onchange={(e) => set('EnableAutoLoadVillageBuilding', (e.target as HTMLInputElement).checked ? 1 : 0)} />
      {tr('accountSettings.autoLoadBuildings')}
    </label>
    <label class="flex items-center gap-2 text-sm">
      <input type="checkbox" checked={get('EnableAutoStartAdventure') === 1}
        onchange={(e) => set('EnableAutoStartAdventure', (e.target as HTMLInputElement).checked ? 1 : 0)} />
      {tr('accountSettings.autoAdventure')}
    </label>
  </div>

  <!-- Chrome -->
  <div class="border rounded p-3">
    <h3 class="text-sm font-medium text-gray-700 mb-2">{tr('accountSettings.chrome')}</h3>
    <label class="flex items-center gap-2 text-sm">
      <input type="checkbox" checked={get('HeadlessChrome') === 1}
        onchange={(e) => set('HeadlessChrome', (e.target as HTMLInputElement).checked ? 1 : 0)} />
      {tr('accountSettings.headless')}
    </label>
  </div>

  <button class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm" onclick={handleSave}>
    {tr('accountSettings.save')}
  </button>
</div>
