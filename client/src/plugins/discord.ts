import { authorizeServer } from "@/api/server/auth";
import type { Auth } from "@/models/discord/auth";
import { useAppStore } from "@/stores/app";
import { DiscordSDK } from "@discord/embedded-app-sdk";

import type { App } from 'vue'


const key = "discordSdk"

export async function setupDiscordSdk(app: App) {
  const discordSdk = new DiscordSDK(import.meta.env.VITE_DISCORD_CLIENT_ID);
  await discordSdk.ready();

  const { code } = await discordSdk.commands.authorize({
    client_id: import.meta.env.VITE_DISCORD_CLIENT_ID,
    response_type: "code",
    state: "",
    prompt: "none",
    scope: [
      "identify",
      "guilds",
      "applications.commands"
    ],
  });
  const access_token = await authorizeServer(code)

  const auth: Auth = await discordSdk.commands.authenticate({
    access_token,
  });

  if (auth == null) {
    throw new Error("Authenticate command failed");
  }

  const store = useAppStore()
  if (auth.access_token)
    store.authToken = auth.access_token
  else
    throw Error("No auth token")

  // TODO: save other info later
  store.user = auth.user

  app.provide(key, discordSdk);
}

export const useDiscordSdk = (): DiscordSDK => inject(key)!
