import { writable } from 'svelte/store'

export const accountSettings = writable<Record<string, number>>({})
export const villageSettings = writable<Record<string, number>>({})
