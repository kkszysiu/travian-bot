import { writable } from 'svelte/store'

export interface BuildingItem {
  id: number
  type: number
  typeName: string
  level: number
  maxLevel: number
  isUnderConstruction: boolean
  location: number
  color: string
}

export interface QueueBuildingItem {
  position: number
  location: number
  typeName: string
  level: number
  completeTime: string
}

export interface JobItem {
  id: number
  position: number
  type: number
  content: string
  display: string
}

export const buildings = writable<BuildingItem[]>([])
export const queueBuildings = writable<QueueBuildingItem[]>([])
export const jobs = writable<JobItem[]>([])
