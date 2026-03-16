/** @type {import('tailwindcss').Config} */
module.exports = {
  darkMode: 'class',
  content: [
    './pages/**/*.{js,ts,jsx,tsx,mdx}',
    './components/**/*.{js,ts,jsx,tsx,mdx}',
    './app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      colors: {
        brand: {
          DEFAULT: '#6366f1',
          50:  '#eef2ff',
          100: '#e0e7ff',
          200: '#c7d2fe',
          300: '#a5b4fc',
          400: '#818cf8',
          500: '#6366f1',
          600: '#4f46e5',
          700: '#4338ca',
          800: '#3730a3',
          900: '#312e81',
          950: '#1e1b4b',
        },
        accent: {
          DEFAULT: '#06b6d4',
          300: '#67e8f9',
          400: '#22d3ee',
          500: '#06b6d4',
          600: '#0891b2',
        },
        violet: {
          400: '#a78bfa',
          500: '#8b5cf6',
          600: '#7c3aed',
          700: '#6d28d9',
        },
        surface: {
          DEFAULT: '#0b1120',
          elevated: '#111d30',
          card: '#162038',
          border: '#1d2e4e',
          muted: '#475569',
          subtle: '#64748b',
        },
      },
      fontFamily: {
        sans: ['var(--font-inter)', 'system-ui', 'sans-serif'],
        mono: ['var(--font-mono)', 'JetBrains Mono', 'Fira Code', 'monospace'],
      },
      animation: {
        'shimmer':     'shimmer 2.5s linear infinite',
        'glow-pulse':  'glowPulse 2s ease-in-out infinite',
        'pulse-slow':  'pulse 3s cubic-bezier(0.4,0,0.6,1) infinite',
        'fade-in':     'fadeIn 0.3s ease-in-out',
        'slide-up':    'slideUp 0.3s ease-out',
        'spin-slow':   'spin 3s linear infinite',
      },
      keyframes: {
        shimmer: {
          '0%':   { transform: 'translateX(-100%)' },
          '100%': { transform: 'translateX(100%)' },
        },
        glowPulse: {
          '0%, 100%': { opacity: '0.3' },
          '50%':      { opacity: '1' },
        },
        fadeIn: {
          '0%':   { opacity: '0' },
          '100%': { opacity: '1' },
        },
        slideUp: {
          '0%':   { transform: 'translateY(10px)', opacity: '0' },
          '100%': { transform: 'translateY(0)',    opacity: '1' },
        },
      },
      backgroundImage: {
        'gradient-brand':    'linear-gradient(135deg, #4338ca, #6366f1, #8b5cf6)',
        'gradient-accent':   'linear-gradient(135deg, #0891b2, #06b6d4, #22d3ee)',
        'gradient-metallic': 'linear-gradient(150deg, #1a2844 0%, #111c30 45%, #14213a 100%)',
      },
      boxShadow: {
        'brand-glow':  '0 0 24px -4px rgba(99,102,241,0.5)',
        'accent-glow': '0 0 20px -4px rgba(6,182,212,0.45)',
        'metallic':    'inset 0 1px 0 rgba(255,255,255,0.06), 0 4px 24px -4px rgba(0,0,0,0.5)',
      },
    },
  },
  plugins: [],
}
