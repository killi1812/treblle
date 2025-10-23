<template>
  <v-card>
    <v-card-title class="d-flex justify-center">
      <v-avatar size="128" class="mt-4">
        <v-img :src="avatar" alt="User Avatar" />
      </v-avatar>
    </v-card-title>
    <v-card-text class="text-center">
      <h2 class="text-h5">{{ user.username }}</h2>
    </v-card-text>
  </v-card>

</template>
<script lang="ts">

export interface UserPreviewProps {
  user: User
}

</script>
<script setup lang="ts">
import type { User } from '@/models/discord/user';

const props = defineProps<UserPreviewProps>()
const avatar = computed(() => `https://cdn.discordapp.com/avatars/${props.user.id}/${props.user.avatar}.webp?size=128`)

</script>
