<script lang="ts" setup>
defineProps<{
  saveChanges: () => void
}>()
</script>

<template>
  <v-footer tag="header">
    <v-row no-gutters>
      <v-col>
        <div class="button-box">
          <v-btn variant="elevated" prepend-icon="mdi-backburger" to="/">
            Menu
          </v-btn>
          <v-btn variant="elevated" base-color="blue" @click="saveChanges">
            Save changes
          </v-btn>
        </div>
      </v-col>
    </v-row>
  </v-footer>
</template>

<style lang="scss" scoped>
.button-box {
  display: flex;
  justify-content: right;
  align-items: center;
  height: 100%;
  gap: 10px;
}
</style>
