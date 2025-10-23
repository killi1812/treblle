<template>
  <v-container v-if="app.user">
    <v-row class="d-flex justify-center">
      <v-col cols="12" md="6">
        <UserPreview :user="app.user" />
      </v-col>
    </v-row>
    <v-row class="d-flex justify-center mt-4">
      <v-col cols="12" md="6">
        <UserStats :user-id="app.user.id" />
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup lang="ts">
import { useAppStore } from '@/stores/app';
const app = useAppStore()
</script>
