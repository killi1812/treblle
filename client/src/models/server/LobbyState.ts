import type { User } from "../discord/user";

export interface LobbyState {
  players: (User | null)[];
  spectators: User[];
}
