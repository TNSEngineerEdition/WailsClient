<script lang="ts" setup>
const props = defineProps<{
  saveChanges: () => void
}>()
</script>

<template>
  <v-card class="sidebar" flat>
    <v-card-title class="header">Edit tram network</v-card-title>

    <v-card-text class="content-container">Future content</v-card-text>

    <v-card-actions class="button-box">
      <!-- <v-btn base-color="black" variant="elevated">Cancel</v-btn> -->
      <v-btn base-color="blue" variant="elevated" @click="saveChanges"
        >Save</v-btn
      >
    </v-card-actions>
  </v-card>
</template>

<style lang="scss" scoped>
.sidebar {
  width: 350px;
  display: flex;
  flex-direction: column;
}

.header,
.button-box {
  flex: 0 0 auto;
  height: 56px;
  display: flex;
  align-items: center;
}

.content-container {
  flex: 1 1 auto;
  overflow-y: auto;
}

.button-box {
  justify-content: center;
  gap: 10px;
}
</style>
