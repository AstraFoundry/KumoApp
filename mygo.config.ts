import { defineConfig } from "mygo-cli";

export default defineConfig({
  name: "Kumo",
  identifier: "io.kumo.KumoApp",
  version: "0.1.0",
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
  // Self-updates need the maintainers' signing key: run `mygo keygen`, add
  //   updates: { publicKey: "<mygo-update.pub>", github: "AstraFoundry/KumoApp" },
  // and store mygo-update.key as the MYGO_UPDATER_PRIVATE_KEY secret.
  macos: {
    minimumSystemVersion: "13.0",
    infoPlist: {
      LSApplicationCategoryType: "public.app-category.utilities",
      CFBundleLocalizations: ["en", "zh-Hans"],
    },
  },
});
