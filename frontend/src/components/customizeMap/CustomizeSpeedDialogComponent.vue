<script lang="ts" setup>
import { ref } from "vue"

const speedDialog = defineModel<boolean>("speedDialog", { default: false })
const currentSpeed = ref<number>(50) // domyślnie 50 km/h

const onCancelCallback = defineModel<() => void>("onCancel", { default: null })
const onSaveCallback = defineModel<(newMaxSpeed: number) => void>(
  "onSpeedSave",
  {
    default: null,
  },
)

function onCancel() {
  onCancelCallback.value()
  speedDialog.value = false
}

function onSpeedSave() {
  onSaveCallback.value(currentSpeed.value)
  speedDialog.value = false
}
</script>

<template>
  <v-dialog
    v-model="speedDialog"
    persistent
    max-width="400px"
    transition="dialog-transition"
  >
    <v-card title="Set max speed">
      <v-card-text>
        <v-number-input
          v-model.number="currentSpeed"
          label="Max speed (km/h)"
          :reverse="false"
          control-variant="split"
          :min="5"
          :max="100"
          :step="5"
          inset
        ></v-number-input>
      </v-card-text>
      <v-card-actions>
        <v-btn text="Cancel" @click="onCancel"></v-btn>
        <v-btn color="primary" text="Save" @click="onSpeedSave"></v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
