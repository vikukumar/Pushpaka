/** @type {import('tailwindcss').Config} */
export default {
  content: ['./src/**/*.{astro,html,js,jsx,md,mdx,svelte,ts,tsx,vue}'],
  theme: {
    extend: {
      colors: {
        'brand': {
          '50': '#eef2ff',
          '100': '#e0e7ff',
          '200': '#c7d2fe',
          '300': '#a5b4fc',
          '400': '#818cf8',
          '500': '#6366f1',
          '600': '#4f46e5',
          '700': '#4338ca',
          '800': '#3730a3',
          '900': '#312e81',
          '950': '#1e1b4b',
        },
        'accent': {
          '50': '#ecfeff',
          '100': '#cffafe',
          '200': '#a5f3fc',
          '300': '#67e8f9',
          '400': '#22d3ee',
          '500': '#06b6d4',
          '600': '#0891b2',
          '700': '#0e7490',
          '800': '#155e75',
          '900': '#164e63',
          '950': '#083344',
        },
        'surface': {
          'bg': '#0f172a',
          'elevated': '#1e293b',
          'border': '#334155',
          'muted': '#475569',
          'subtle': '#64748b',
        }
      },
      backgroundImage: {
        'gradient-metallic': 'linear-gradient(135deg, rgba(99,102,241,0.1) 0%, rgba(6,182,212,0.05) 50%, rgba(139,92,246,0.1) 100%)',
        'gradient-radial-brand': 'radial-gradient(circle at 50% 50%, rgba(99,102,241,0.15) 0%, transparent 50%)',
      },
      boxShadow: {
        'metallic': '0 20px 60px -10px rgba(99, 102, 241, 0.3), inset 0 1px 1px rgba(255, 255, 255, 0.1)',
        'metallic-lg': '0 40px 80px -20px rgba(99, 102, 241, 0.4), inset 0 2px 2px rgba(255, 255, 255, 0.15)',
        'accent-glow': '0 0 30px rgba(6, 182, 212, 0.2), inset 0 1px 1px rgba(255, 255, 255, 0.1)',
      },
      animation: {
        'float': 'float 6s ease-in-out infinite',
        'glow': 'glow 3s ease-in-out infinite',
      },
      keyframes: {
        float: {
          '0%, 100%': { transform: 'translateY(0px)' },
          '50%': { transform: 'translateY(-20px)' },
        },
        glow: {
          '0%, 100%': { 'box-shadow': '0 0 20px rgba(99, 102, 241, 0.5)' },
          '50%': { 'box-shadow': '0 0 40px rgba(6, 182, 212, 0.3)' },
        }
      }
    },
  },
  plugins: [],
};
