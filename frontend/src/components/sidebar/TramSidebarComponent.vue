<script setup lang="ts">
import SidebarComponent from "@components/sidebar/SidebarComponent.vue"

const model = defineModel<boolean>({ required: true })
</script>

<template>
  <SidebarComponent
    v-model="model"
    position="left"
    title="Tram #2137"
    title-icon="mdi-tram"
  >
    This is tram sidebar content
  </SidebarComponent>
</template>
