<script setup lang="ts">
import SidebarComponent from "@components/sidebar/SidebarComponent.vue"

const model = defineModel<boolean>({ required: true })
</script>

<template>
  <SidebarComponent
    v-model="model"
    position="right"
    title="Stella-Sawickiego 01"
    title-icon="mdi-tram-side"
  >
    This is stop sidebar content
  </SidebarComponent>
</template>
