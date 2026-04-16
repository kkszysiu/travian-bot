import { writable, derived } from 'svelte/store'

export interface VillageListItem {
  id: number
  name: string
  x: number
  y: number
  isActive: boolean
  isUnderAttack: boolean
  evasionState: number
}

export const villages = writable<VillageListItem[]>([])
export const selectedVillageId = writable<number | null>(null)

export const selectedVillage = derived(
  [villages, selectedVillageId],
  ([$villages, $id]) => $villages.find(v => v.id === $id) ?? null
)
