// Utilities
import type { User } from '@/models/discord/user'
import { defineStore } from 'pinia'

export const useAppStore = defineStore('app', () => {

  // Discord auth key
  const _authTokenState = ref("")
  const authToken = computed({
    get: () => _authTokenState.value,
    set: (value: string) => _authTokenState.value = value
  })

  // User
  const _userState = ref<User | undefined>(undefined)
  const user = computed({
    get: () => _userState.value,
    set: (value: User) => _userState.value = value
  })

  const userAvatarIcon = computed(() => _userState.value ? `https://cdn.discordapp.com/avatars/${_userState.value.id}/${_userState.value.avatar}.webp?size=128` : undefined)

  return {
    authToken,
    user,
    userAvatarIcon
  }
})
