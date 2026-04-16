import { writable } from 'svelte/store'

export interface TaskItem {
  task: string
  executeAt: string
  stage: string
}

export const tasks = writable<TaskItem[]>([])
