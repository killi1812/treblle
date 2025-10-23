<template>
  <v-container>
    <h1>Game Test</h1>

    <h2>
      discard pile
    </h2>
    <div>
      <div v-if="discardPile" class="d-flex flex-row ga-1 flex-wrap">
        <div v-for="tile in discardPile">
          {{ tile.suite }}-{{ tile.value }}
        </div>
      </div>
    </div>
    <hr class="mb-3 mt-3" />
    <div class="d-flex flex-row ga-1 flex-wrap">
      <h3>
        Wall Size:
        <span>
          {{ gameSocketService.gameState.value.publicState?.wallSize ?? 0 }} |
        </span>
      </h3>
      <h3>
        Blocked Wall Size:
        <span>
          {{ gameSocketService.gameState.value.publicState?.blockedWallSize ?? 0 }} |
        </span>
      </h3>
      <h3>
        Bonus Tile:
        <span v-if="gameSocketService.gameState.value.publicState?.bonusTile">
          {{ gameSocketService.gameState.value.publicState?.bonusTile.suite }}-
          {{ gameSocketService.gameState.value.publicState?.bonusTile.value }}
        </span>
      </h3>
    </div>

    <hr class="mb-3 mt-3" />
    <h3>
      Active tile:
      <span v-if="activeTile">
        {{ activeTile.suite }}-
        {{ activeTile.value }}
      </span>
    </h3>

    <hr class="mb-3 mt-3" />
    <div>
      <h2>
        Hand
        <span v-if="myHand">
          : {{ myHand.mainHand.length + myHand.revealedHand.length }}
        </span>
      </h2>
      <div v-if="myHand" class="d-flex flex-row ga-1 flex-wrap">
        <div v-for="tile in myHand.mainHand">
          <v-btn @click="discardTile(tile.tileId)" :disabled="!gameSocketService.isConnected.value">
            {{ tile.suite }}-{{ tile.value }}
          </v-btn>
        </div>
      </div>
      <hr class="mb-3 mt-3" />
      <div v-for="tile in myHand?.revealedHand">
        <v-btn :disabled="!gameSocketService.isConnected.value">
          {{ tile.suite }}-{{ tile.value }}
        </v-btn>
      </div>

    </div>
    <hr class="mb-3 mt-3" />
    <div>
      <h2>
        Bonus Hand
        <span v-if="myHand?.bonusTiles">
          : {{ myHand.bonusTiles.length }}
        </span>
      </h2>
      <div v-if="myHand" class="d-flex flex-row ga-1 flex-wrap">
        <div v-for="tile in myHand.bonusTiles">
          <v-btn :disabled="!gameSocketService.isConnected.value">
            {{ tile.suite }}-{{ tile.value }}
          </v-btn>
        </div>
      </div>
    </div>


    <hr class="mb-3 mt-4" />
    <h2>
      Actions
    </h2>
    <div class="d-flex flex-row ga-1 flex-wrap">
      <v-btn @click="gameSocketService.drawTile()" :disabled="!gameSocketService.isConnected.value">Draw Tile</v-btn>
      <v-btn @click="gameSocketService.siezeTile()" :disabled="!gameSocketService.isConnected.value">Sieze Tile</v-btn>
    </div>
  </v-container>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue';
import { gameSocketService } from '@/api/server/game';
import { useAppStore } from '@/stores/app';
import { useDiscordSdk } from '@/plugins/discord';

const app = useAppStore();
const discordSdk = useDiscordSdk();

const activeTile = computed(() => gameSocketService.gameState.value.publicState?.activeTile)
const myHand = computed(() => gameSocketService.gameState.value.playerState)
const discardPile = computed(() => gameSocketService.gameState.value?.publicState?.discardPile)


watch([gameSocketService.gameState], () => {
  console.log(gameSocketService.gameState)
})

onMounted(() => {
  gameSocketService.connect(discordSdk.instanceId);
});

onUnmounted(() => {
  gameSocketService.disconnect();
});


function discardTile(tileId: number) {
  if (!app.user)
    return

  gameSocketService.discardTile(tileId);
}

</script>
