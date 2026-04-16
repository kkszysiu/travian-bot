<script lang="ts">
  import { buildings, queueBuildings, jobs, type BuildingItem, type QueueBuildingItem, type JobItem } from '../../stores/buildings'
  import { AddNormalBuildJob, AddResourceBuildJob, DeleteJob, DeleteAllJobs, MoveJob, GetJobs, GetBuildings, GetQueueBuildings, GetStorage, ExportJobs, ImportJobs, GetAvailableNewBuildings } from '../../../../wailsjs/go/app/App'
  import { EventsOn, EventsOff } from '../../../../wailsjs/runtime/runtime'
  import { t } from '../../i18n'

  let { villageId }: { villageId: number } = $props()

  let buildingList = $state<BuildingItem[]>([])
  let queueList = $state<QueueBuildingItem[]>([])
  let jobList = $state<JobItem[]>([])
  let storage = $state<{ wood: number; clay: number; iron: number; crop: number; warehouse: number; granary: number; freeCrop: number } | null>(null)
  let tr = $state<(key: string, params?: Record<string, string | number>) => string>((k) => k)
  t.subscribe(v => tr = v)

  buildings.subscribe(v => buildingList = v)
  queueBuildings.subscribe(v => queueList = v)
  jobs.subscribe(v => jobList = v)

  async function loadStorage() {
    try { storage = await GetStorage(villageId) } catch { storage = null }
  }
  $effect(() => {
    loadStorage()
    EventsOn('storage:modified', () => loadStorage())
    return () => {
      EventsOff('storage:modified')
    }
  })

  // Normal build form - use location as key (unique per building slot)
  let normalLocation = $state(0)
  let normalLevel = $state(1)

  // New construction on empty site
  let newBuildingType = $state(0)
  let availableBuildings = $state<{ type: number; name: string }[]>([])

  // Detect if selected location is an empty site and load available buildings
  let selectedIsEmptySite = $derived(
    buildingList.find(b => b.location === normalLocation)?.type === 0
  )

  $effect(() => {
    if (selectedIsEmptySite && villageId > 0) {
      GetAvailableNewBuildings(villageId).then(list => {
        availableBuildings = list || []
        if (list && list.length > 0) newBuildingType = list[0].type
      }).catch(() => { availableBuildings = [] })
    }
  })

  // Resource build form
  let resourcePlan = $state(0)
  let resourceLevel = $state(10)

  let jobError = $state('')

  async function addNormalJob() {
    jobError = ''
    const b = buildingList.find(b => b.location === normalLocation)
    if (!b) return
    const type = b.type === 0 ? newBuildingType : b.type
    if (type === 0) return
    try {
      await AddNormalBuildJob({ villageId, type, level: normalLevel, location: b.location })
      await refreshJobs()
    } catch (e: any) { jobError = e?.message || String(e) }
  }

  async function addResourceJob() {
    jobError = ''
    try {
      await AddResourceBuildJob({ villageId, plan: resourcePlan, level: resourceLevel })
      await refreshJobs()
    } catch (e: any) { jobError = e?.message || String(e) }
  }

  async function deleteJob(id: number) {
    try {
      await DeleteJob(id)
      await refreshJobs()
    } catch (e: any) { jobError = e?.message || String(e) }
  }

  async function moveJob(id: number, direction: string) {
    try {
      await MoveJob(id, direction)
      await refreshJobs()
    } catch (e: any) { jobError = e?.message || String(e) }
  }

  async function refreshJobs() {
    const j = await GetJobs(villageId)
    jobs.set(j || [])
  }

  // Quick-add upgrade jobs from building list
  async function upgradeOne(building: BuildingItem) {
    if (building.type === 0) return // Skip empty sites
    const targetLevel = building.level + 1
    if (targetLevel > building.maxLevel) return
    await AddNormalBuildJob({ villageId, type: building.type, level: targetLevel, location: building.location })
    await refreshJobs()
  }

  async function upgradeToMax(building: BuildingItem) {
    if (building.type === 0) return // Skip empty sites
    if (building.level >= building.maxLevel) return
    await AddNormalBuildJob({ villageId, type: building.type, level: building.maxLevel, location: building.location })
    await refreshJobs()
  }

  // Import / Export / Delete All
  let fileInput: HTMLInputElement
  let importing = $state(false)
  let exporting = $state(false)

  async function exportJobs() {
    exporting = true
    try {
      const jsonData = await ExportJobs(villageId)
      const blob = new Blob([jsonData], { type: 'application/json' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `jobs-${villageId}.json`
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      URL.revokeObjectURL(url)
    } catch (e) {
      console.error('Export failed:', e)
    } finally {
      exporting = false
    }
  }

  function triggerImport() {
    fileInput?.click()
  }

  async function handleImportFile(event: Event) {
    const target = event.target as HTMLInputElement
    const file = target.files?.[0]
    if (!file) return

    importing = true
    try {
      const text = await file.text()
      await ImportJobs(villageId, text)
      await refreshJobs()
    } catch (e) {
      console.error('Import failed:', e)
    } finally {
      importing = false
      target.value = ''
    }
  }

  let confirmingDeleteAll = $state(false)
  let confirmTimer: ReturnType<typeof setTimeout> | null = null

  function handleDeleteAll() {
    if (jobList.length === 0) return
    if (!confirmingDeleteAll) {
      confirmingDeleteAll = true
      confirmTimer = setTimeout(() => { confirmingDeleteAll = false }, 3000)
      return
    }
    if (confirmTimer) clearTimeout(confirmTimer)
    confirmingDeleteAll = false
    doDeleteAll()
  }

  async function doDeleteAll() {
    try {
      await DeleteAllJobs(villageId)
      await refreshJobs()
    } catch (e: any) {
      jobError = e?.message || String(e)
    }
  }
</script>

<div class="flex gap-4 h-full">
  <!-- Left: Queue + Storage + Buildings -->
  <div class="w-64 flex-shrink-0 space-y-3 overflow-y-auto">
    <!-- Storage -->
    {#if storage}
      <div class="border rounded p-2 text-xs space-y-1 bg-gray-50">
        <div class="flex justify-between"><span class="text-yellow-700">{tr('buildTab.wood')}</span><span>{storage.wood?.toLocaleString()}</span></div>
        <div class="flex justify-between"><span class="text-orange-700">{tr('buildTab.clay')}</span><span>{storage.clay?.toLocaleString()}</span></div>
        <div class="flex justify-between"><span class="text-gray-600">{tr('buildTab.iron')}</span><span>{storage.iron?.toLocaleString()}</span></div>
        <div class="flex justify-between"><span class="text-green-700">{tr('buildTab.crop')}</span><span>{storage.crop?.toLocaleString()}</span></div>
        <div class="flex justify-between text-gray-400 border-t pt-1 mt-1">
          <span>{tr('buildTab.warehouse')}</span><span>{storage.warehouse?.toLocaleString()}</span>
        </div>
        <div class="flex justify-between text-gray-400"><span>{tr('buildTab.granary')}</span><span>{storage.granary?.toLocaleString()}</span></div>
        <div class="flex justify-between text-gray-400"><span>{tr('buildTab.freeCrop')}</span><span>{storage.freeCrop?.toLocaleString()}</span></div>
      </div>
    {/if}

    <!-- Building Queue -->
    <div class="border rounded">
      <h4 class="text-xs font-medium p-2 border-b bg-gray-50">{tr('buildTab.buildingQueue')}</h4>
      {#each queueList as item}
        <div class="px-2 py-1 text-xs border-b border-gray-50">
          <span>{item.typeName} Lv.{item.level}</span>
          <span class="text-gray-400 float-right">{item.completeTime}</span>
        </div>
      {/each}
      {#if queueList.length === 0}
        <p class="p-2 text-xs text-gray-400">{tr('buildTab.empty')}</p>
      {/if}
    </div>

    <!-- Buildings -->
    <div class="border rounded">
      <h4 class="text-xs font-medium p-2 border-b bg-gray-50">{tr('buildTab.buildings')}</h4>
      <div class="max-h-64 overflow-y-auto">
        {#each buildingList as b}
          <div class="px-2 py-1 text-xs border-b border-gray-50 flex items-center justify-between group"
            style="background-color: {b.color}20">
            <span style="color: {b.color}">
              [{b.location}] {b.typeName}
            </span>
            <span class="flex items-center gap-1">
              <span class="text-gray-500">Lv.{b.level}{b.isUnderConstruction ? ' *' : ''}</span>
              {#if b.type > 0 && b.level < b.maxLevel}
                <span class="hidden group-hover:flex gap-0.5">
                  <button class="px-1 text-blue-500 hover:text-blue-700" title="Upgrade +1" onclick={() => upgradeOne(b)}>+1</button>
                  <button class="px-1 text-green-500 hover:text-green-700" title="Upgrade to max" onclick={() => upgradeToMax(b)}>max</button>
                </span>
              {/if}
            </span>
          </div>
        {/each}
      </div>
    </div>
  </div>

  <!-- Middle: Build Forms -->
  <div class="w-56 flex-shrink-0 space-y-3">
    {#if jobError}<div class="p-2 bg-red-100 text-red-700 text-xs rounded">{jobError}</div>{/if}
    <!-- Normal Build -->
    <div class="border rounded p-3 space-y-2">
      <h4 class="text-xs font-medium">{tr('buildTab.normalBuilding')}</h4>
      <select bind:value={normalLocation} class="w-full px-2 py-1 border rounded text-xs">
        {#each buildingList as b}
          <option value={b.location}>[{b.location}] {b.typeName}{b.type > 0 ? ` Lv.${b.level}` : ''}</option>
        {/each}
      </select>
      {#if selectedIsEmptySite}
        <select bind:value={newBuildingType} class="w-full px-2 py-1 border rounded text-xs">
          {#each availableBuildings as bt}
            <option value={bt.type}>{bt.name}</option>
          {/each}
        </select>
      {/if}
      <input type="number" bind:value={normalLevel} min="1" max="20"
        class="w-full px-2 py-1 border rounded text-xs" placeholder={tr('buildTab.level')} />
      <button class="w-full px-2 py-1 text-xs bg-blue-500 text-white rounded hover:bg-blue-600" onclick={addNormalJob}>
        {tr('buildTab.addNormalBuild')}
      </button>
    </div>

    <!-- Resource Build -->
    <div class="border rounded p-3 space-y-2">
      <h4 class="text-xs font-medium">{tr('buildTab.resourceBuilding')}</h4>
      <select bind:value={resourcePlan} class="w-full px-2 py-1 border rounded text-xs">
        <option value={0}>{tr('buildTab.allResources')}</option>
        <option value={1}>{tr('buildTab.excludeCrop')}</option>
        <option value={2}>{tr('buildTab.onlyCrop')}</option>
      </select>

      <input type="number" bind:value={resourceLevel} min="1" max="20"
        class="w-full px-2 py-1 border rounded text-xs" placeholder={tr('buildTab.level')} />
      <button class="w-full px-2 py-1 text-xs bg-green-500 text-white rounded hover:bg-green-600" onclick={addResourceJob}>
        {tr('buildTab.addResourceBuild')}
      </button>
    </div>
  </div>

  <!-- Right: Jobs List -->
  <div class="flex-1 border rounded overflow-hidden flex flex-col">
    <div class="p-2 border-b bg-gray-50 flex items-center justify-between gap-2">
      <div class="flex items-center gap-1.5">
        <h4 class="text-xs font-medium">{tr('buildTab.jobs')}</h4>
        <span class="text-xs text-gray-400">{jobList.length} {tr('buildTab.items')}</span>
      </div>
      <div class="flex items-center gap-1">
        <input type="file" accept=".json" class="hidden" bind:this={fileInput} onchange={handleImportFile} />
        <button
          class="px-2 py-0.5 text-xs bg-blue-500 text-white rounded hover:bg-blue-600 disabled:opacity-50"
          onclick={triggerImport}
          disabled={importing}
        >
          {importing ? tr('buildTab.importing') : tr('buildTab.import')}
        </button>
        <button
          class="px-2 py-0.5 text-xs bg-green-500 text-white rounded hover:bg-green-600 disabled:opacity-50"
          onclick={exportJobs}
          disabled={exporting || jobList.length === 0}
        >
          {exporting ? tr('buildTab.exporting') : tr('buildTab.export')}
        </button>
        <button
          class="px-2 py-0.5 text-xs text-white rounded disabled:opacity-50 {confirmingDeleteAll ? 'bg-orange-500 hover:bg-orange-600' : 'bg-red-500 hover:bg-red-600'}"
          onclick={handleDeleteAll}
          disabled={jobList.length === 0}
        >
          {confirmingDeleteAll ? tr('buildTab.confirmDeleteAll', { count: jobList.length }) : tr('buildTab.deleteAll')}
        </button>
      </div>
    </div>
    <div class="flex-1 overflow-y-auto">
      {#each jobList as job}
        <div class="px-2 py-1.5 text-xs border-b border-gray-50 flex items-center gap-2 group">
          <span class="flex-1 truncate">{job.display}</span>
          <div class="hidden group-hover:flex gap-1">
            <button class="text-gray-400 hover:text-blue-500" onclick={() => moveJob(job.id, 'top')} title={tr('buildTab.top')}>&#8593;&#8593;</button>
            <button class="text-gray-400 hover:text-blue-500" onclick={() => moveJob(job.id, 'up')} title={tr('buildTab.up')}>&#8593;</button>
            <button class="text-gray-400 hover:text-blue-500" onclick={() => moveJob(job.id, 'down')} title={tr('buildTab.down')}>&#8595;</button>
            <button class="text-gray-400 hover:text-blue-500" onclick={() => moveJob(job.id, 'bottom')} title={tr('buildTab.bottom')}>&#8595;&#8595;</button>
            <button class="text-gray-400 hover:text-red-500" onclick={() => deleteJob(job.id)} title={tr('buildTab.deleteTooltip')}>&#10005;</button>
          </div>
        </div>
      {/each}
      {#if jobList.length === 0}
        <p class="p-3 text-xs text-gray-400 text-center">{tr('buildTab.noJobs')}</p>
      {/if}
    </div>
  </div>
</div>
