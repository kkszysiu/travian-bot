<script lang="ts">
  import { GetTroopTypes } from '../../../../wailsjs/go/app/App'
  import { t } from '../../i18n'

  let { tribe, building = 0, value, onchange }: {
    tribe: number
    building?: number
    value: number
    onchange: (value: number) => void
  } = $props()

  let troops = $state<{ type: number; name: string }[]>([])
  let loading = $state(false)
  let tr = $state<(key: string, params?: Record<string, string | number>) => string>((k) => k)
  t.subscribe(v => tr = v)

  $effect(() => {
    loadTroops(tribe, building)
  })

  async function loadTroops(t: number, b: number) {
    loading = true
    try {
      troops = await GetTroopTypes(t, b) || []
    } catch (e) {
      console.error('Failed to load troop types', e)
      troops = []
    } finally {
      loading = false
    }
  }
</script>

<select
  class="w-full px-2 py-1 border rounded text-sm bg-white"
  {value}
  disabled={loading}
  onchange={(e) => onchange(parseInt((e.target as HTMLSelectElement).value))}
>
  <option value="0">{tr('troopSelector.none')}</option>
  {#each troops as troop (troop.type)}
    <option value={troop.type}>{troop.name}</option>
  {/each}
</select>
