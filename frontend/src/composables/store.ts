import { reactive } from "vue"

type NodeModifications = {
  neighborMaxSpeed: Record<number, number>
}

export const modifiedNodes = reactive<Record<number, NodeModifications>>({})
