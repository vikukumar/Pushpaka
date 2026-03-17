import { defineConfig } from 'astro/config';
import tailwind from '@astrojs/tailwind';

export default defineConfig({
  site: 'https://pushpaka.vikshro.in/',
  base: '/',
  integrations: [tailwind()],
  output: 'static'
});
