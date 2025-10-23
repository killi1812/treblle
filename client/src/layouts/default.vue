<template>
  <div>
    <AppHeader />
    <v-main>
      <router-view v-slot="{ Component, route }">
        <v-scroll-x-transition>
          <component :is="Component" :key="route" />
        </v-scroll-x-transition>
      </router-view>
    </v-main>
  </div>
</template>

<script lang="ts" setup>
//
</script>
