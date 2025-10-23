<template>
  <v-container class="fill-height" max-width="900">
    <div>
      <v-img class="mb-4" height="150" src="@/assets/rocket.png" alt="discord" />
      <div class="mb-8 text-center">
        <div class="text-body-2 font-weight-light mb-n1">Welcome to</div>
        <h1 class="text-h2 font-weight-bold">Mahjong</h1>
      </div>
      <div class="d-flex justify-center flex-column">
        <span class="text-body-2 font-weight-light ">
          Activity Channel:
        </span>
        <span class="text-body-2 font-weight-bold mb-n1">{{ channelName }}</span>
        <div>
          <v-img v-if="guildImgUrl" class="icon" :src="guildImgUrl" alt="Guild Image" width="128" height="128" />
        </div>
        <v-btn to="/game/lobby">Play</v-btn>
      </div>

    </div>
  </v-container>
</template>

<script setup lang="ts">
import { getGuilds } from '@/api/discord/myData';
import type { Guild } from '@/models/discord/guild';
import { useDiscordSdk } from '@/plugins/discord';
import { useDiscordStore } from '@/stores/discord';


// Definistions
const discordSdk = useDiscordSdk()
const dStore = useDiscordStore()

const channelName = ref<string>("Unknown")
const guildImgUrl = ref<string | undefined>(undefined)

// functions
async function appendVoiceChannelName() {
  // Requesting the channel in GDMs (when the guild ID is null) requires
  // the dm_channels.read scope which requires Discord approval.
  if (discordSdk.channelId != null && discordSdk.guildId != null) {
    // Over RPC collect info about the channel
    const channel = await discordSdk.commands.getChannel({ channel_id: discordSdk.channelId });
    if (channel.name != null) {
      channelName.value = channel.name
    }
  }
}

async function appendGuildAvatar() {
  const guilds = await getGuilds()
  if (!guilds)
    return

  const currentGuild = guilds.find((g: Guild) => g.id === discordSdk.guildId);
  dStore.currentGuild = currentGuild

  // TODO: see about moving to discordStore
  if (currentGuild != null)
    guildImgUrl.value = `https://cdn.discordapp.com/icons/${currentGuild.id}/${currentGuild.icon}.webp?size=128`
}

// Hooks
onMounted(async () => {
  await appendVoiceChannelName()
  await appendGuildAvatar()
})


</script>

<style lang="css" scoped>
.icon {
  border-radius: 50%;
}
</style>
