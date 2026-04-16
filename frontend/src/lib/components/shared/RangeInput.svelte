<script lang="ts">
  let { label, unit, min, max, onMinChange, onMaxChange }: {
    label: string
    unit: string
    min: number
    max: number
    onMinChange: (v: number) => void
    onMaxChange: (v: number) => void
  } = $props()
</script>

<div class="flex items-center gap-2 text-sm">
  <span class="text-xs text-gray-500 w-24 flex-shrink-0">{label}</span>
  <input type="number" class="w-20 px-2 py-1 border rounded text-sm" value={min}
    onchange={(e) => onMinChange(parseInt((e.target as HTMLInputElement).value) || 0)} />
  <span class="text-gray-400">~</span>
  <input type="number" class="w-20 px-2 py-1 border rounded text-sm" value={max}
    onchange={(e) => onMaxChange(parseInt((e.target as HTMLInputElement).value) || 0)} />
  {#if unit}<span class="text-xs text-gray-400">{unit}</span>{/if}
</div>
