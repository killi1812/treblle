import type { Guild } from "@/models/discord/guild";
import discordApi from "./discordAxios";

/**
 *
 * Retrives discord guilds (servers)
 */
export async function getGuilds(): Promise<Guild[] | undefined> {
  try {
    const rez = await discordApi.get<Guild[]>(`/v10/users/@me/guilds`)
    if (rez.status == 200)
      return rez.data
  }
  catch (error: any) {
  }
}
