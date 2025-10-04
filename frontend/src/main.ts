import { createApp } from "vue"
import App from "./App.vue"
import "@plugins/leaflet"
import vuetify from "@plugins/vuetify"
import router from "@plugins/router"

const app = createApp(App)
app.use(vuetify)
app.use(router)
app.mount("#app")
