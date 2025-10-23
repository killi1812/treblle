<template>
  <v-container>
    <v-row>
      <v-col v-for="(player, index) in players" :key="index" cols="12" sm="6" md="3">
        <v-card class="player-slot" :class="{ 'empty-slot': !player }">
          <v-card-text class="d-flex flex-column align-center justify-center fill-height">
            <template v-if="player">
              <v-avatar size="80">
                <v-img :src="`https://cdn.discordapp.com/avatars/${player.id}/${player.avatar}.webp?size=80`"
                  :alt="player.username"></v-img>
              </v-avatar>
              <div class="mt-4 text-h6">{{ player.username }}</div>
            </template>
            <template v-else>
              <v-btn color="primary" @click="joinGame(index)">
                Join Slot
              </v-btn>
            </template>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <v-row class="mt-6">
      <v-col class="text-center">
        <v-btn color="success" @click="startGame" :disabled="!isLobbyFull">
          Play
        </v-btn>
      </v-col>
    </v-row>

    <v-row class="mt-6">
      <v-col>
        <v-card>
          <v-card-title>Spectators ({{ spectators.length }})</v-card-title>
          <v-list>
            <v-list-item v-for="spectator in spectators" :key="spectator.id">
              <template v-slot:prepend>
                <v-avatar size="32">
                  <v-img :src="`https://cdn.discordapp.com/avatars/${spectator.id}/${spectator.avatar}.png`"
                    :alt="spectator.username"></v-img>
                </v-avatar>
              </template>
              <v-list-item-title>{{ spectator.username }}</v-list-item-title>
            </v-list-item>
            <v-list-item v-if="!spectators.length">
              <v-list-item-title class="text-center">No spectators yet.</v-list-item-title>
            </v-list-item>
          </v-list>
          <v-card-actions>
            <v-spacer></v-spacer>
            <v-btn color="secondary" @click="joinAsSpectator" :disabled="inSpectators">
              Join as Spectator
            </v-btn>
          </v-card-actions>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue';
import { useDiscordSdk } from '@/plugins/discord';
import { useAppStore } from '@/stores/app';
import { lobbySocketService } from '@/api/server/lobby';

const discordSdk = useDiscordSdk();
const app = useAppStore();
const router = useRouter()

// Get reactive state from the socket service
const players = computed(() => lobbySocketService.lobbyState.value.players);
const spectators = computed(() => lobbySocketService.lobbyState.value.spectators);

// const isLobbyFull = computed(() => players.value.every(p => p !== null));
const isLobbyFull = computed(() => true)

const inSpectators = computed(() => spectators.value.some(s => s.id === app.user?.id))

onMounted(async () => {
  await discordSdk.ready();

  console.log(discordSdk.instanceId)
  lobbySocketService.connect(discordSdk.instanceId);
});

onUnmounted(() => {
  if (app.user) {
    lobbySocketService.leave(app.user);
  }
  lobbySocketService.disconnect();
});

function joinGame(index: number) {
  if (!app.user) return;

  if (players.value[index] === null) {
    lobbySocketService.joinAsPlayer(app.user, index);
  }
};

function joinAsSpectator() {
  if (!app.user) return;
  if (inSpectators.value) {
    return;
  }
  lobbySocketService.joinAsSpectator(app.user);
}

async function startGame() {
  // TODO: game setup on beckend required
  await router.replace("/game/game")
};
</script>

<style scoped>
.player-slot {
  height: 200px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.empty-slot {
  border: 2px dashed grey;
}
</style>
