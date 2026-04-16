import { writable, derived } from 'svelte/store'
import { selectedAccountId } from './accounts'

export interface AccountStatus {
  status: number
  color: string
}

// Map of accountId -> status
export const statuses = writable<Record<number, AccountStatus>>({})

export const selectedStatus = derived(
  [statuses, selectedAccountId],
  ([$statuses, $id]) => $id != null ? $statuses[$id] ?? { status: 0, color: 'black' } : { status: 0, color: 'black' }
)

export const statusLabels: Record<number, string> = {
  0: 'Offline',
  1: 'Starting',
  2: 'Online',
  3: 'Pausing',
  4: 'Paused',
  5: 'Stopping'
}
