import type { GameState, GameMessage, PrivatePlayerState, PublicState } from '@/models/server/Game';
import { useAppStore } from '@/stores/app';
import { ref } from 'vue';

// Determine WebSocket protocol based on window location protocol
const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
const WS_URL = `${protocol}//${window.location.host}/api/ws/game`;

class GameSocketService {
  private socket: WebSocket | null = null;
  private store = useAppStore()
  private userId = computed(() => this.store.user?.id ?? '0')
  public gameState = ref<GameState>({ playerState: null, publicState: null });
  public isConnected = ref(false);

  public connect(gameId: string) {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      console.log('Game WebSocket is already connected.');
      return;
    }

    this.socket = new WebSocket(`${WS_URL}/${this.userId.value}/${gameId}`);

    this.socket.onopen = () => {
      console.log('WebSocket connected to game');
      this.isConnected.value = true;
    };

    this.socket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        // TODO: see about better message switching when more messages come
        if ('mainHand' in data) { // It's a PlayerState
          this.gameState.value.playerState = data as PrivatePlayerState;
        } else { // It's a PublicState
          this.gameState.value.publicState = data as PublicState;
        }
        console.log('Game state updated:', this.gameState.value);
      } catch (error) {
        console.error('Error parsing game state:', error);
      }
    };

    this.socket.onclose = () => {
      console.log('WebSocket disconnected from game');
      this.isConnected.value = false;
      this.socket = null;
    };

    this.socket.onerror = (error) => {
      console.error('WebSocket error:', error);
    };
  }

  private sendMessage(message: GameMessage) {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      this.socket.send(JSON.stringify(message));
    } else {
      console.error('WebSocket is not connected.');
    }
  }

  public drawTile() {
    this.sendMessage({ action: 'draw_tile', userId: this.userId.value });
  }

  public discardTile(tileId: number) {
    this.sendMessage({ action: 'discard_tile', userId: this.userId.value, tileId });
  }

  public siezeTile() {
    this.sendMessage({ action: 'seize', userId: this.userId.value });
  }

  public disconnect() {
    if (this.socket) {
      this.socket.close(1001);
    }
  }
}

export const gameSocketService = new GameSocketService();
