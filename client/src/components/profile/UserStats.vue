<template>
  <v-card>
    <v-card-title>
      Under Progress
    </v-card-title>
    <v-card-text>
      <p>This section is currently under construction. More features will be added soon!</p>
    </v-card-text>
  </v-card>
</template>
<script lang="ts">
export interface UserStatsProps {
  userId: string
}
</script>
<script setup lang="ts">

defineProps<UserStatsProps>()
</script>
