import type { User } from "../discord/user"

/**
 * Defines the possible actions that can be sent to the lobby WebSocket server.
 */
export type LobbyAction = "join_player" | "join_spectator" | "leave"

/**
 * Represents the structure of a message sent to the lobby WebSocket server.
 */
export interface LobbyMessage {
  /** The action to be performed by the server. */
  action: LobbyAction
  /** The user performing the action. */
  user: User
  /** The player slot index (0-3), only required for the 'join_player' action. */
  slot?: number
}
