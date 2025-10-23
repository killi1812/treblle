import type { Tile } from "./Tile";

export type GameAction = "draw_tile" | "seize" | "discard_tile" | "pass_turn";

export interface GameMessage {
  action: GameAction;
  userId: string;
  slot?: number;
  tileId?: number
}


export interface PrivatePlayerState {
  mainHand: Tile[];
  revealedHand: Tile[];
  bonusTiles: Tile[];
}

export interface PublicPlayerState {
  mainHandLen: number;
  revealedHand: Tile[];
  bonusTiles: Tile[];
}

export interface PublicState {
  wallSize: number;
  blockedWallSize: number;
  bonusTile: Tile;
  discardPile: Tile[];
  activeTile?: Tile;
  playerHands: Record<string, PublicPlayerState>
}

export interface GameState {
  playerState: PrivatePlayerState | null;
  publicState: PublicState | null;
}

