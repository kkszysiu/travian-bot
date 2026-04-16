import { writable } from 'svelte/store'

export interface LogEntry {
  message: string
  level: string
  time: string
}

export const logs = writable<LogEntry[]>([])
