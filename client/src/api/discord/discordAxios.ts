import { useAppStore } from '@/stores/app';
import axios from 'axios';

const DISCORD_URL = "https://discord.com/api";
const discordApi = axios.create({
  baseURL: DISCORD_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

discordApi.interceptors.request.use(
  (config) => {
    const authStore = useAppStore();
    const accessToken = authStore.authToken;
    if (accessToken) {

      // TODO: what do i need this for ??
      config.headers = config.headers || {};

      config.headers['Authorization'] = `Bearer ${accessToken}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);


export default discordApi;
