<script lang="ts">
  import { GetVillages, GetBuildings, GetQueueBuildings, GetJobs, RefreshVillage } from '../../../../wailsjs/go/app/App'
  import { villages, selectedVillageId, type VillageListItem } from '../../stores/villages'
  import { buildings, queueBuildings, jobs, type BuildingItem, type QueueBuildingItem, type JobItem } from '../../stores/buildings'
  import BuildTab from '../village/BuildTab.svelte'
  import TransferRulesTab from '../village/TransferRulesTab.svelte'
  import VillageSettingTab from '../village/VillageSettingTab.svelte'
  import { EventsOn, EventsOff } from '../../../../wailsjs/runtime/runtime'
  import { t } from '../../i18n'

  let { accountId }: { accountId: number } = $props()

  let villageList = $state<VillageListItem[]>([])
  let selectedId = $state<number | null>(null)
  let activeSubTab = $state<string>('build')
  let tr = $state<(key: string, params?: Record<string, string | number>) => string>((k) => k)
  t.subscribe(v => tr = v)

  villages.subscribe(v => villageList = v)
  selectedVillageId.subscribe(v => selectedId = v)

  async function loadVillages() {
    try {
      const list = await GetVillages(accountId)
      villages.set(list || [])
    } catch (e) { console.error(e) }
  }

  async function loadVillageData(vid: number) {
    try {
      const [b, q, j] = await Promise.all([
        GetBuildings(vid),
        GetQueueBuildings(vid),
        GetJobs(vid)
      ])
      buildings.set(b || [])
      queueBuildings.set(q || [])
      jobs.set(j || [])
    } catch (e) { console.error(e) }
  }

  function selectVillage(id: number) {
    selectedVillageId.set(id)
    loadVillageData(id)
  }

  $effect(() => {
    loadVillages()
    EventsOn('villages:modified', () => loadVillages())
    EventsOn('buildings:modified', () => { if (selectedId) loadVillageData(selectedId) })
    EventsOn('jobs:modified', () => { if (selectedId) loadVillageData(selectedId) })
    EventsOn('transfer_rules:modified', () => {})
    return () => {
      EventsOff('villages:modified', 'buildings:modified', 'jobs:modified', 'transfer_rules:modified')
    }
  })

  const subTabIds = ['build', 'transferRules', 'villageSettings'] as const
  const subTabLabelKeys: Record<string, string> = {
    build: 'villageTab.build',
    transferRules: 'villageTab.transferRules',
    villageSettings: 'villageTab.villageSettings',
  }
</script>

<div class="flex h-full gap-4">
  <!-- Village List (left ~20%) -->
  <div class="w-48 flex-shrink-0 border rounded overflow-y-auto">
    <h3 class="text-sm font-medium p-2 border-b bg-gray-50">{tr('villageTab.villages')}</h3>
    {#each villageList as village (village.id)}
      <div
        class="flex items-center border-b border-gray-50 hover:bg-blue-50 group"
        class:bg-blue-100={selectedId === village.id}
      >
        <button
          class="flex-1 text-left px-3 py-2 text-sm"
          onclick={() => selectVillage(village.id)}
        >
          <span class:text-red-500={village.isUnderAttack}>{village.name}</span>
          <span class="text-xs text-gray-400">({village.x}|{village.y})</span>
        </button>
        <button
          class="hidden group-hover:block px-2 py-1 text-xs text-gray-400 hover:text-blue-500"
          title={tr('villageTab.refreshTooltip')}
          onclick={() => RefreshVillage(accountId, village.id)}
        >&#8635;</button>
      </div>
    {/each}
    {#if villageList.length === 0}
      <p class="p-3 text-xs text-gray-400">{tr('villageTab.noVillages')}</p>
    {/if}
  </div>

  <!-- Village Content (right ~80%) -->
  <div class="flex-1 flex flex-col overflow-hidden">
    {#if selectedId != null}
      <!-- Sub-tab bar -->
      <div class="flex border-b border-gray-200 mb-3">
        {#each subTabIds as tabId}
          <button
            class="px-3 py-1.5 text-sm font-medium transition-colors"
            class:text-blue-600={activeSubTab === tabId}
            class:border-b-2={activeSubTab === tabId}
            class:border-blue-600={activeSubTab === tabId}
            class:text-gray-500={activeSubTab !== tabId}
            onclick={() => activeSubTab = tabId}
          >
            {tr(subTabLabelKeys[tabId])}
          </button>
        {/each}
      </div>

      <div class="flex-1 overflow-auto">
        {#if activeSubTab === 'build'}
          <BuildTab villageId={selectedId} />
        {:else if activeSubTab === 'transferRules'}
          <TransferRulesTab villageId={selectedId} {accountId} />
        {:else if activeSubTab === 'villageSettings'}
          <VillageSettingTab villageId={selectedId} {accountId} />
        {/if}
      </div>
    {:else}
      <div class="flex items-center justify-center h-full text-gray-400">
        <p>{tr('villageTab.selectVillage')}</p>
      </div>
    {/if}
  </div>
</div>
