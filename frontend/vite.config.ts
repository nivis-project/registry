/// <reference types="vitest/config" />
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// The SPA is data-agnostic: in dev it reads the static contract from /registry
// (served out of public/), and in production the same files ship alongside the
// built assets (S3 + CloudFront later). No live backend is required.
export default defineConfig({
  plugins: [react()],
  base: "./",
  test: {
    environment: "jsdom",
    globals: true,
  },
});
