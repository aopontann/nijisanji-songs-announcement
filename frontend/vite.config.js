import { defineConfig } from 'vite';

export default defineConfig({
  base: '/',
  build: {
    rollupOptions: {
      input: [
        'index.html',
        'serviceworker.js',
        'manifest.webmanifest'
      ],
      output: {
        entryFileNames: '[name].js',
      }
    }
  }
});
