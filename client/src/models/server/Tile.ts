export type TileCategory = "suited" | "honor" | "bonus";

export type TileSuite =
  // Suited
  | "dot"
  | "bamboo"
  | "character"
  // Winds
  | "east"
  | "south"
  | "west"
  | "north"
  // Dragons
  | "white"
  | "green"
  | "red"
  // Flowers
  | "plum"
  | "orchid"
  | "chrysanthemum"
  | "bamboo_bonus"
  // Seasons
  | "spring"
  | "summer"
  | "autumn"
  | "winter";

export interface Tile {
  tileId: number;
  category: TileCategory;
  suite: TileSuite;
  value: number;
  visible: boolean
}

// TODO: map dragons, winds, flowers, and seasons from the numbers to values
