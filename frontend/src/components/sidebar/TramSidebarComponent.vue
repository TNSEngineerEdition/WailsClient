<script setup lang="ts">
import SidebarComponent from "@components/sidebar/SidebarComponent.vue"
import { simulation } from "@wails/go/models";

const model = defineModel<boolean>({ required: true })

const props = defineProps<{
  tramID: number,
  tramDetails: simulation.TramDetails
}>()
</script>

<template>
  <SidebarComponent
    v-model="model"
    position="left"
    :title="`#` + props.tramID.toString()"
    title-icon="mdi-tram"
  >
  <div class="sidebarDetails">
    <span><b>Linka</b>: {{ props.tramDetails.route }}</span>
    <span><b>Smer</b>: {{ props.tramDetails.trip_head_sign }}</span>
    <span><b>Rychlost</b>: {{ props.tramDetails.speed }} km/h</span>

    <div class="mt-4">
      <b>Zastavky:</b>
      <ul>
        <li v-for="(stop, index) in props.tramDetails.stop_names" :key="index">
          {{ stop }}
        </li>
      </ul>
    </div>
  </div>

  </SidebarComponent>
</template>

<style lang="scss" scoped>
.sidebarDetails {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
</style>
