<template>
  <v-app-bar color="primary" :elevation="8" height="60">
    <v-app-bar-nav-icon class="text-white" @click="open = !open"></v-app-bar-nav-icon>
    <img id="logo" src="../../assets/logo.png" alt="logo" />
    <v-app-bar-title>Mahjong</v-app-bar-title>
    <v-spacer></v-spacer>
    <v-btn icon @click="router.back()">
      <v-icon>mdi-arrow-left</v-icon>
    </v-btn>

    <v-menu location="bottom">
      <template v-slot:activator="{ props }">
        <v-btn icon v-bind="props">
          <v-icon> mdi-account-circle </v-icon>
        </v-btn>
      </template>
      <v-list>
        <IconListItem to="/me/myProfile" icon="mdi-account">
          MY_PROFILE
        </IconListItem>
      </v-list>
    </v-menu>
  </v-app-bar>

  <v-navigation-drawer class="offset" :rail="!open" rail-width="60" location="left" permanent persistent elevation="2">
    <v-list nav>
      <!-- Information section -->
      <!-- Title -->
      <v-list-item v-show="open" readonly>INFORMATION</v-list-item>
      <v-divider></v-divider>
      <!-- Navigations -->
      <IconListItem to="/" icon="mdi-home" tooltip-text="START.DASHBOARD" :show-tooltip="!open">
        HOME
      </IconListItem>
      <!-- Master Data section -->
      <!-- Title -->
      <v-list-item v-show="open" readonly>MASTER_DATA</v-list-item>
      <v-divider></v-divider>
      <!-- Navigations  -->
      <IconListItem to="/help/help" icon="mdi-help-circle" tooltip-text="START.HELP" :show-tooltip="!open">
        HELP
      </IconListItem>
      <IconListItem to="/help/about" icon="mdi-information" tooltip-text="START.ABOUT" :show-tooltip="!open">
        ABOUT
      </IconListItem>
      <IconListItem to="/help/test" icon="mdi-information" tooltip-text="START.TEST" :show-tooltip="!open">
        TEST
      </IconListItem>
    </v-list>
  </v-navigation-drawer>

</template>


<script lang="ts" setup>

const open = ref(false)
const router = useRouter()

</script>
<style lang="css" scoped>
#logo {
  width: 40px;
  height: 30px;
}

.offset {
  padding-bottom: 40px;
}

.uppercase {
  text-transform: uppercase;
}
</style>
