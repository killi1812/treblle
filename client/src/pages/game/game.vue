<route lang="yaml">
meta:
  layout: game
</route>

<template>
  <div class="d-flex flex-row align-center justify-center ">
    <div id="game-container"></div>
    <div class="pa-5">
      <div>
        <v-btn @click="changeScene">Change Scene</v-btn>
      </div>
      <div>
        <v-btn :disabled="canMoveSprite" @click="moveSprite">Toggle Movement</v-btn>
      </div>
      <div class="spritePosition">Sprite Position:
        <pre>{{ spritePosition }}</pre>
      </div>
      <div>
        <v-btn @click="addSprite">Add New Sprite</v-btn>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue';
import Phaser from 'phaser';
import { EventBus } from '@/game/EventBus';
import StartGame from '@/game/main';
import { MainMenu } from '@/game/scenes/MainMenu';

// Save the current scene instance
const sceneGlobal = ref();
const game = ref();
const canMoveSprite = ref(false)

const spritePosition = ref({ x: 0, y: 0 });

function changeScene() {
  const mainMenue = toRaw(sceneGlobal.value) as MainMenu;
  if (mainMenue)
    //  Call the changeScene method defined in the `MainMenu`, `Game` and `GameOver` Scenes
    mainMenue.changeScene();

}

function moveSprite() {
  if (game.value === undefined)
    return

  const scene = toRaw(sceneGlobal.value) as MainMenu;
  if (!scene)
    return


  // Get the update logo position
  (scene as MainMenu).moveLogo(({ x, y }) => {
    spritePosition.value = { x, y };
  });
}

function addSprite() {
  const scene = toRaw(sceneGlobal.value) as Phaser.Scene;
  if (!scene)
    return

  // Add a new sprite to the current scene at a random position
  const x = Phaser.Math.Between(64, scene.scale.width - 64);
  const y = Phaser.Math.Between(64, scene.scale.height - 64);

  // `add.sprite` is a Phaser GameObjectFactory method and it returns a Sprite Game Object instance
  const star = scene.add.sprite(x, y, 'star');

  //  ... which you can then act upon. Here we create a Phaser Tween to fade the star sprite in and out.
  //  You could, of course, do this from within the Phaser Scene code, but this is just an example
  //  showing that Phaser objects and systems can be acted upon from outside of Phaser itself.
  scene.add.tween({
    targets: star,
    duration: 500 + Math.random() * 1000,
    alpha: 0,
    yoyo: true,
    repeat: -1
  });

}

// Hooks
onMounted(() => {
  game.value = StartGame('game-container');
  EventBus.on('current-scene-ready', (scene_instance: Phaser.Scene) => {
    sceneGlobal.value = scene_instance;
    canMoveSprite.value = !(sceneGlobal.value instanceof MainMenu);
  });
});

onUnmounted(() => {
  if (game.value) {
    game.value.destroy(true);
    game.value = null;
  }
});

defineExpose({ scene: sceneGlobal, game });
</script>
