import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte'; // <-- The CORRECT import path!

/** @type {import('@sveltejs/kit').Config} */
const config = {
  // Consult https://svelte.dev/docs/integrations#preprocessors
  // for more information about preprocessors
  preprocess: vitePreprocess(),

  kit: {
    // This is the most important part. We are forcing the static adapter.
    adapter: adapter({
      // The output directory for the final static files.
      // We explicitly name it 'build' for clarity.
      pages: 'build',
      assets: 'build',
      
      // This creates a single index.html for Single-Page App (SPA) mode.
      // This is the key to making it work with Go's file server.
      fallback: 'index.html',
      
      precompress: false,
      strict: true
    })
  }
};

export default config;
