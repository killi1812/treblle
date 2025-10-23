<template>
  <v-tooltip v-if="tooltipText" :disabled="!showTooltip" :text="tooltipText">
    <template v-slot:activator="{ props }">
      <v-list-item :to="to" :prepend-icon="genIcon()" v-bind="props">
        <slot>
        </slot>
      </v-list-item>
    </template>
  </v-tooltip>
  <v-list-item v-else :to="to" :prepend-icon="genIcon()" v-bind="props">
    <slot>
    </slot>
  </v-list-item>

</template>

<script setup lang="ts">

const props = defineProps({
  to: {
    type: String,
    requried: true,
  },
  icon: {
    type: String,
    requried: true,
  },
  tooltipText: {
    type: String,
    required: false,
  },
  showTooltip: {
    type: Boolean,
    default: false,
  },
})

const router = useRouter()

const genIcon = () => {
  if (router.currentRoute.value.path == props.to)
    return props.icon
  // Return outline if not on page
  return `${props.icon}-outline`
}

</script>
