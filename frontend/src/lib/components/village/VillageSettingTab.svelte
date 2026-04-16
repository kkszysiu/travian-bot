<script lang="ts">
  import { GetVillageSettings, SaveVillageSettings, GetAccountSettings, GetVillages } from '../../../../wailsjs/go/app/App'
  import RangeInput from '../shared/RangeInput.svelte'
  import TroopSelector from '../shared/TroopSelector.svelte'
  import { t } from '../../i18n'

  let { villageId, accountId }: { villageId: number; accountId: number } = $props()

  let settings = $state<Record<string, number>>({})
  let tribe = $state(0)
  let error = $state('')
  let success = $state('')
  let villageList = $state<Array<{id: number; name: string; x: number; y: number}>>([])
  let tr = $state<(key: string, params?: Record<string, string | number>) => string>((k) => k)
  t.subscribe(v => tr = v)

  async function load() {
    try { settings = await GetVillageSettings(villageId) || {} } catch (e) { console.error(e) }
  }

  async function loadTribe() {
    try {
      const acctSettings = await GetAccountSettings(accountId) || {}
      tribe = acctSettings['Tribe'] ?? 0
    } catch (e) { console.error(e) }
  }

  async function loadVillages() {
    try { villageList = await GetVillages(accountId) || [] } catch (e) { console.error(e) }
  }

  $effect(() => { load() })
  $effect(() => { loadTribe() })
  $effect(() => { loadVillages() })

  function get(key: string): number { return settings[key] ?? 0 }
  function set(key: string, val: number) { settings = { ...settings, [key]: val } }

  async function handleSave() {
    error = ''; success = ''
    try {
      await SaveVillageSettings(villageId, settings)
      success = 'Settings saved'
    } catch (e: any) { error = e?.message || String(e) }
  }

  const buildingLabelKeys: Record<string, string> = {
    Barrack: 'villageSettings.barrack',
    Stable: 'villageSettings.stable',
    GreatBarrack: 'villageSettings.greatBarrack',
    GreatStable: 'villageSettings.greatStable',
    Workshop: 'villageSettings.workshop',
  }

  // Maps UI building names to their enum.Building IDs for troop filtering
  const buildingEnumIds: Record<string, number> = {
    Barrack: 19,
    Stable: 20,
    GreatBarrack: 29,
    GreatStable: 30,
    Workshop: 21,
  }
</script>

<div class="max-w-2xl space-y-4 overflow-y-auto">
  {#if error}<div class="p-2 bg-red-100 text-red-700 text-sm rounded">{error}</div>{/if}
  {#if success}<div class="p-2 bg-green-100 text-green-700 text-sm rounded">{success}</div>{/if}

  <!-- Upgrade Settings -->
  <div class="border rounded p-3 space-y-2">
    <h3 class="text-sm font-medium text-gray-700">{tr('villageSettings.upgrade')}</h3>
    <label class="flex items-center gap-2 text-sm">
      <input type="checkbox" checked={get('UseHeroResourceForBuilding') === 1}
        onchange={(e) => set('UseHeroResourceForBuilding', (e.target as HTMLInputElement).checked ? 1 : 0)} />
      {tr('villageSettings.useHeroResources')}
    </label>
    <label class="flex items-center gap-2 text-sm">
      <input type="checkbox" checked={get('ApplyRomanQueueLogicWhenBuilding') === 1}
        onchange={(e) => set('ApplyRomanQueueLogicWhenBuilding', (e.target as HTMLInputElement).checked ? 1 : 0)} />
      {tr('villageSettings.romanQueue')}
    </label>
    <label class="flex items-center gap-2 text-sm">
      <input type="checkbox" checked={get('CompleteImmediately') === 1}
        onchange={(e) => set('CompleteImmediately', (e.target as HTMLInputElement).checked ? 1 : 0)} />
      {tr('villageSettings.completeImmediately', { time: get('CompleteImmediatelyTime') })}
    </label>
    <label class="flex items-center gap-2 text-sm">
      <input type="checkbox" checked={get('UseSpecialUpgrade') === 1}
        onchange={(e) => set('UseSpecialUpgrade', (e.target as HTMLInputElement).checked ? 1 : 0)} />
      {tr('villageSettings.specialUpgrade')}
    </label>
  </div>

  <!-- Auto Refresh -->
  <div class="border rounded p-3 space-y-2">
    <h3 class="text-sm font-medium text-gray-700">{tr('villageSettings.autoRefresh')}</h3>
    <label class="flex items-center gap-2 text-sm">
      <input type="checkbox" checked={get('AutoRefreshEnable') === 1}
        onchange={(e) => set('AutoRefreshEnable', (e.target as HTMLInputElement).checked ? 1 : 0)} />
      {tr('villageSettings.enableAutoRefresh')}
    </label>
    <RangeInput label={tr('villageSettings.refreshTime')} unit={tr('villageSettings.min')}
      min={get('AutoRefreshMin')} max={get('AutoRefreshMax')}
      onMinChange={(v) => set('AutoRefreshMin', v)} onMaxChange={(v) => set('AutoRefreshMax', v)} />
  </div>

  <!-- NPC -->
  <div class="border rounded p-3 space-y-2">
    <h3 class="text-sm font-medium text-gray-700">{tr('villageSettings.npc')}</h3>
    <label class="flex items-center gap-2 text-sm">
      <input type="checkbox" checked={get('AutoNPCEnable') === 1}
        onchange={(e) => set('AutoNPCEnable', (e.target as HTMLInputElement).checked ? 1 : 0)} />
      {tr('villageSettings.enableNpc')}
    </label>
    <label class="flex items-center gap-2 text-sm">
      <input type="checkbox" checked={get('AutoNPCOverflow') === 1}
        onchange={(e) => set('AutoNPCOverflow', (e.target as HTMLInputElement).checked ? 1 : 0)} />
      {tr('villageSettings.npcOverflow')}
    </label>
    <label class="block text-sm">
      <span class="text-xs text-gray-500">{tr('villageSettings.granaryPercent')}</span>
      <input type="number" class="ml-2 w-16 px-2 py-1 border rounded text-sm" min="0" max="100"
        value={get('AutoNPCGranaryPercent')} onchange={(e) => set('AutoNPCGranaryPercent', parseInt((e.target as HTMLInputElement).value))} />
    </label>
    <div class="flex gap-2 text-xs">
      <label>{tr('villageSettings.wood')} <input type="number" class="w-12 px-1 py-0.5 border rounded" value={get('AutoNPCWood')} onchange={(e) => set('AutoNPCWood', parseInt((e.target as HTMLInputElement).value))} /></label>
      <label>{tr('villageSettings.clay')} <input type="number" class="w-12 px-1 py-0.5 border rounded" value={get('AutoNPCClay')} onchange={(e) => set('AutoNPCClay', parseInt((e.target as HTMLInputElement).value))} /></label>
      <label>{tr('villageSettings.iron')} <input type="number" class="w-12 px-1 py-0.5 border rounded" value={get('AutoNPCIron')} onchange={(e) => set('AutoNPCIron', parseInt((e.target as HTMLInputElement).value))} /></label>
      <label>{tr('villageSettings.crop')} <input type="number" class="w-12 px-1 py-0.5 border rounded" value={get('AutoNPCCrop')} onchange={(e) => set('AutoNPCCrop', parseInt((e.target as HTMLInputElement).value))} /></label>
    </div>
  </div>

  <!-- Troop Training -->
  <div class="border rounded p-3 space-y-2">
    <h3 class="text-sm font-medium text-gray-700">{tr('villageSettings.troopTraining')}</h3>
    <label class="flex items-center gap-2 text-sm">
      <input type="checkbox" checked={get('TrainTroopEnable') === 1}
        onchange={(e) => set('TrainTroopEnable', (e.target as HTMLInputElement).checked ? 1 : 0)} />
      {tr('villageSettings.enableTroopTraining')}
    </label>
    <RangeInput label={tr('villageSettings.repeatTime')} unit={tr('villageSettings.min')}
      min={get('TrainTroopRepeatTimeMin')} max={get('TrainTroopRepeatTimeMax')}
      onMinChange={(v) => set('TrainTroopRepeatTimeMin', v)} onMaxChange={(v) => set('TrainTroopRepeatTimeMax', v)} />
    <label class="flex items-center gap-2 text-sm">
      <input type="checkbox" checked={get('TrainWhenLowResource') === 1}
        onchange={(e) => set('TrainWhenLowResource', (e.target as HTMLInputElement).checked ? 1 : 0)} />
      {tr('villageSettings.trainLowResource')}
    </label>
    {#each ['Barrack', 'Stable', 'GreatBarrack', 'GreatStable', 'Workshop'] as building}
      <div class="pl-2 border-l-2 border-gray-200 space-y-1">
        <span class="text-xs font-medium text-gray-600">{tr(buildingLabelKeys[building])}</span>
        <div class="flex gap-2 items-center text-xs">
          <span class="text-xs text-gray-500 w-16 flex-shrink-0">{tr('villageSettings.troop')}</span>
          <div class="w-40">
            <TroopSelector {tribe} building={buildingEnumIds[building]} value={get(`${building}Troop`)}
              onchange={(v) => set(`${building}Troop`, v)} />
          </div>
          <RangeInput label={tr('villageSettings.amount')} unit=""
            min={get(`${building}AmountMin`)} max={get(`${building}AmountMax`)}
            onMinChange={(v) => set(`${building}AmountMin`, v)} onMaxChange={(v) => set(`${building}AmountMax`, v)} />
        </div>
      </div>
    {/each}
  </div>

  <!-- Quest -->
  <div class="border rounded p-3">
    <label class="flex items-center gap-2 text-sm">
      <input type="checkbox" checked={get('AutoClaimQuestEnable') === 1}
        onchange={(e) => set('AutoClaimQuestEnable', (e.target as HTMLInputElement).checked ? 1 : 0)} />
      {tr('villageSettings.autoClaimQuests')}
    </label>
  </div>

  <!-- Attack Evasion -->
  <div class="border rounded p-3 space-y-2">
    <h3 class="text-sm font-medium text-gray-700">{tr('villageSettings.attackEvasion')}</h3>
    <label class="flex items-center gap-2 text-sm">
      <input type="checkbox" checked={get('AttackEvasionEnable') === 1}
        onchange={(e) => set('AttackEvasionEnable', (e.target as HTMLInputElement).checked ? 1 : 0)} />
      {tr('villageSettings.enableAttackEvasion')}
    </label>
    <label class="block text-sm">
      <span class="text-xs text-gray-500">{tr('villageSettings.safeVillage')}</span>
      <select class="ml-2 px-2 py-1 border rounded text-sm"
        value={get('AttackEvasionSafeVillageID')}
        onchange={(e) => set('AttackEvasionSafeVillageID', parseInt((e.target as HTMLSelectElement).value))}>
        <option value={0}>-- {tr('villageSettings.selectSafeVillage')} --</option>
        {#each villageList.filter(v => v.id !== villageId) as v}
          <option value={v.id}>{v.name} ({v.x}|{v.y})</option>
        {/each}
      </select>
    </label>
    <label class="flex items-center gap-2 text-sm">
      <input type="checkbox" checked={get('AttackEvasionEvacResources') === 1}
        onchange={(e) => set('AttackEvasionEvacResources', (e.target as HTMLInputElement).checked ? 1 : 0)} />
      {tr('villageSettings.evacuateResources')}
    </label>
    <RangeInput label={tr('villageSettings.checkInterval')} unit={tr('villageSettings.sec')}
      min={get('AttackEvasionCheckIntervalMin')} max={get('AttackEvasionCheckIntervalMax')}
      onMinChange={(v) => set('AttackEvasionCheckIntervalMin', v)} onMaxChange={(v) => set('AttackEvasionCheckIntervalMax', v)} />
  </div>

  <button class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm" onclick={handleSave}>
    {tr('villageSettings.save')}
  </button>
</div>
