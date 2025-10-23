/**
 * main.ts
 *
 * Bootstraps Vuetify and other plugins then mounts the App`
 */

// Plugins
import { registerPlugins } from '@/plugins'

// Components
import App from './App.vue'

// Composables
import { createApp } from 'vue'

// Styles
import '@/styles/style.css'
import { setupDiscordSdk } from './plugins/discord'

const app = createApp(App)

registerPlugins(app)

await setupDiscordSdk(app)

app.mount('#app')
