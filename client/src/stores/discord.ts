import type { Guild } from '@/models/discord/guild'
import { defineStore } from 'pinia'

export const useDiscordStore = defineStore('discord', () => {
  // Discord auth key
  const _currentGuild = ref<Guild | undefined>(undefined)
  const currentGuild = computed({
    get: () => _currentGuild.value,
    set: (value: Guild) => _currentGuild.value = value
  })

  return {
    currentGuild
  }
})
