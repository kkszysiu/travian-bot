import { writable } from 'svelte/store'

export interface WaitingState {
  visible: boolean
  message: string
}

export const waiting = writable<WaitingState>({ visible: false, message: '' })

export function showOverlay(message: string = '') {
  waiting.set({ visible: true, message })
}

export function hideOverlay() {
  waiting.set({ visible: false, message: '' })
}

export function changeOverlayMessage(message: string) {
  waiting.update(s => ({ ...s, message }))
}
