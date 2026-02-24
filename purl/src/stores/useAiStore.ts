import { computed } from "vue"
import { defineStore } from "pinia"
import { useTicketStore } from "./useTicketStore"

// ── Types ───────────────────────────────────────────────

export interface AiSuggestion {
  headline: string
  body: string
  action: string
  replyText: string
}

// ── Store ────────────────────────────────────────────────

export const useAiStore = defineStore("ai", () => {
  // TODO: load suggestions from the AI API instead of hardcoding them
  const suggestionsList: AiSuggestion[] = [
    {
      headline: "Config 404 — matches known post-update bug",
      body: "Sarah's error is identical to 3 tickets resolved last week after the v2.4 rollout. A missing config.json is caused by the new deploy script skipping static asset copy. Send the one-line fix and mark resolved.",
      action: "Send fix & resolve",
      replyText: "Hi Sarah! This is a known issue with the v2.4 update — the deploy script misses copying config.json. Run this in your project root: `cp node_modules/@purl/defaults/config.json public/`. That should fix it immediately. Let me know if you need anything else!",
    },
    {
      headline: "CSV export bug — patch ships in 24h",
      body: "This is a confirmed bug in v2.4.1 affecting all accounts on Chrome. Engineering has a fix merging today, deploying tomorrow morning. Recommend acknowledging and setting the expectation.",
      action: "Acknowledge & set ETA",
      replyText: "Hi Mike! This is a confirmed bug in v2.4.1 — our team already has a fix and it's deploying tomorrow morning. I'll follow up as soon as it's live. Sorry for the inconvenience!",
    },
    {
      headline: "Billing overage — likely proration edge case",
      body: "Orion Labs upgraded mid-cycle on Mar 3rd. The $600 difference matches a prorated annual add-on. Check their billing history to confirm, then share the breakdown.",
      action: "Pull billing history",
      replyText: "Hi Priya! I looked into this — the $600 difference is a prorated charge for the annual add-on activated on March 3rd. I've attached the itemized breakdown. Let me know if anything looks off and I'm happy to escalate to billing.",
    },
    {
      headline: "Rate limit counter bug — known issue",
      body: "The dashboard quota display has a caching lag of ~2h, so real usage can exceed what's shown. Check their actual usage in the admin panel and consider a temporary limit increase.",
      action: "Check usage & offer increase",
      replyText: "Hi James! There's a known 2-hour caching lag in the dashboard quota display, so your real-time usage can exceed what's shown there. I checked your account directly and you've hit 94% of your limit. I've bumped your limit by 20% for the next 48 hours while we sort out a permanent solution.",
    },
  ]

  const ticketStore = useTicketStore()

  // Maps real ticket IDs to suggestions based on load order
  const suggestions = computed(() => {
    const map: Record<string, AiSuggestion> = {}
    ticketStore.tickets.forEach((t, i) => {
      if (i < suggestionsList.length) map[t.id] = suggestionsList[i]
    })
    return map
  })

  return {
    suggestions,
  }
})
