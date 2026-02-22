import { defineConfig } from "@hey-api/openapi-ts"

export default defineConfig({
  input: "../api/docs/swagger.json",
  output: {
    path: "src/generated",
    postProcess: ["prettier"],
  },
  client: "@hey-api/client-fetch",
})
