import { writable, derived } from 'svelte/store'

export interface AccountListItem {
  id: number
  username: string
  server: string
}

export interface AccessDetail {
  id: number
  username: string
  password: string
  proxyHost: string
  proxyPort: number
  proxyUsername: string
  proxyPassword: string
  useragent: string
  lastUsed: string
}

export interface AccountDetail {
  id: number
  username: string
  server: string
  accesses: AccessDetail[]
}

export const accounts = writable<AccountListItem[]>([])
export const selectedAccountId = writable<number | null>(null)

export const selectedAccount = derived(
  [accounts, selectedAccountId],
  ([$accounts, $id]) => $accounts.find(a => a.id === $id) ?? null
)
