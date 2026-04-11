import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import path from "path";

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@": path.resolve(import.meta.dirname, "client", "src"),
      "@shared": path.resolve(import.meta.dirname, "shared"),
    },
  },
  root: path.resolve(import.meta.dirname, "client"),
  build: {
    outDir: path.resolve(import.meta.dirname, "server-go/dist/public"),
    emptyOutDir: true,
  },
  server: {
    host: "127.0.0.1",
    port: parseInt(process.env.VITE_PORT || "9005"),
    proxy: {
      "/api": {
        target: `http://127.0.0.1:${process.env.PORT || "9002"}`,
        ws: true,
      },
      "/files": `http://127.0.0.1:${process.env.PORT || "9002"}`,
    },
  },
});
