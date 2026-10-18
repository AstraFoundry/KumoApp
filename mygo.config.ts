import { defineConfig } from "mygo-cli";

export default defineConfig({
  name: "Kumo",
  identifier: "io.kumo.KumoApp",
  version: "0.2.0",
  copyright: "© 2026 AstraFoundry",
  // `mygo dev` runs devCommand and loads devUrl; `mygo build` runs
  // buildCommand and embeds frontendDist into the app.
  devUrl: "http://localhost:5173",
  devCommand: "bun run dev:web",
  buildCommand: "bun run build:web",
  frontendDist: "dist",
  bindings: "src/shared/api/mygo.ts",
  out: "build",
  i18n: { dir: "locales", source: "en" },
  urlSchemes: ["kumo", "clash"],
  macos: {
    minimumSystemVersion: "13.0",
    infoPlist: {
      LSApplicationCategoryType: "public.app-category.utilities",
      CFBundleLocalizations: ["en", "zh-Hans"],
    },
  },
});
