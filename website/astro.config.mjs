import { defineConfig } from 'astro/config';
import tailwind from '@astrojs/tailwind';
import node from '@astrojs/node';

export default defineConfig({
  site: 'https://vikukumar.github.io/Pushpaka/',
  base: '/Pushpaka/',
  integrations: [tailwind()],
  output: 'static',
  adapter: node({
    mode: 'middleware'
  })
});
