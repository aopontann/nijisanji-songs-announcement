import { defineConfig } from "astro/config";

// https://astro.build/config
export default defineConfig({
  output: "static",
  // adapter: cloudflare({ mode: "directory" }),
});

// policy