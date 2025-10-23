import type { User } from '@/models/discord/user';
import type { LobbyMessage } from '@/models/server/LobbyMessage';
import type { LobbyState } from '@/models/server/LobbyState';
import { useAppStore } from '@/stores/app';
import { ref } from 'vue';

// Determine WebSocket protocol based on window location protocol
const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
const WS_URL = `${protocol}//${window.location.host}/api/ws/lobby`;

/**
 * Manages the WebSocket connection to the game lobby server.
 * It handles connecting, disconnecting, sending messages, and maintaining the real-time state of the lobby.
 * This service is implemented as a singleton to ensure only one WebSocket connection is active at a time.
 */
class LobbySocketService {
  /** The WebSocket instance. Null if not connected. */
  private socket: WebSocket | null = null;
  private store = useAppStore()
  /** A reactive reference to the current state of the lobby, including players and spectators. */
  public lobbyState = ref<LobbyState>({ players: Array(4).fill(null), spectators: [] });
  /** A reactive boolean indicating the current connection status of the WebSocket. */
  public isConnected = ref(false);


  /**
   * Establishes a WebSocket connection to the server and sets up event listeners
   * for open, message, close, and error events.
   */
  public connect(lobbyId: string) {
    // Avoid multiple connections
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      console.log('WebSocket is already connected.');
      return;
    }


    this.socket = new WebSocket(`${WS_URL}/${this.store.user?.id}/${lobbyId}`);

    this.socket.onopen = () => {
      console.log('WebSocket connected to lobby');
      this.isConnected.value = true;
    };

    this.socket.onmessage = (event) => {
      try {
        const state: LobbyState = JSON.parse(event.data);
        // The Go server sends `null` for empty player slots, which is correct.
        this.lobbyState.value = state;
      } catch (error) {
        console.error('Error parsing lobby state:', error);
      }
    };

    this.socket.onclose = () => {
      console.log('WebSocket disconnected from lobby');
      this.isConnected.value = false;
      this.socket = null;
    };

    this.socket.onerror = (error) => {
      console.error('WebSocket error:', error);
    };
  }

  /**
   * Sends a JSON message to the WebSocket server if the connection is open.
   * @param {LobbyMessage} message - The message object to send.
   */
  private sendMessage(message: LobbyMessage) {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      this.socket.send(JSON.stringify(message));
    } else {
      console.error('WebSocket is not connected.');
    }
  }

  /**
   * Sends a message for the current user to join a specific player slot.
   * @param {User} user - The user joining the lobby.
   * @param {number} slot - The player slot index (0-3) to join.
   */
  public joinAsPlayer(user: User, slot: number) {
    this.sendMessage({ action: 'join_player', user, slot });
  }

  /**
   * Sends a message for the current user to join as a spectator.
   * @param {User} user - The user joining as a spectator.
   */
  public joinAsSpectator(user: User) {
    this.sendMessage({ action: 'join_spectator', user });
  }

  /**
   * Sends a message for the current user to leave the lobby (from either player or spectator status).
   * @param {User} user - The user leaving the lobby.
   */
  public leave(user: User) {
    this.sendMessage({ action: 'leave', user });
  }

  /**
   * Closes the WebSocket connection if it is open.
   */
  public disconnect() {
    if (this.socket) {
      this.socket.close(1001);
    }
  }
}

// Export a singleton instance of the service to be used throughout the application.
export const lobbySocketService = new LobbySocketService();


