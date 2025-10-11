import HomeComponent from "@components/home/HomeComponent.vue"
import SimulationComponent from "@components/simulation/SimulationComponent.vue"
import { createRouter, createWebHashHistory, RouteRecordRaw } from "vue-router"

const routes: RouteRecordRaw[] = [
  {
    path: "/",
    name: "Home",
    component: HomeComponent,
  },
  {
    path: "/simulation",
    name: "Simulation",
    component: SimulationComponent,
  },
]

export default createRouter({
  history: createWebHashHistory(),
  routes,
})
