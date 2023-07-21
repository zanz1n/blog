import { defineConfig } from "astro/config";
import vercel from "@astrojs/vercel/edge";

import svelte from "@astrojs/svelte";

export default defineConfig({
  integrations: [svelte()],
  output: "server",
  adapter: vercel()
});
