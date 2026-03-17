import { defineConfig } from 'astro/config';
import tailwind from '@astrojs/tailwind';

export default defineConfig({
  site: 'https://vikukumar.github.io/Pushpaka/',
  base: '/Pushpaka/',
  integrations: [tailwind()],
  output: 'static'
});
