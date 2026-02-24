import { defineStore } from "pinia"

export const useUserStore = defineStore("user", () => {
  // TODO: load from auth session instead of hardcoding
  const CURRENT_USER = "Alex Chen"

  return {
    CURRENT_USER,
  }
})
