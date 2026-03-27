import nextVitals from 'eslint-config-next/core-web-vitals'

export default [
  ...nextVitals,

  {
    rules: {
      /**
       * 🔧 Your existing override
       */
      'react-hooks/set-state-in-effect': 'off',

      /**
       * ⚠️ React Compiler strict rules (tune, don’t disable blindly)
       */
      'react-hooks/purity': 'warn', // was error → reduce noise, still visible
      'react-hooks/immutability': 'warn',
      'react-hooks/preserve-manual-memoization': 'warn',

      /**
       * 🧠 Hooks correctness (KEEP strict)
       */
      'react-hooks/rules-of-hooks': 'error',

      /**
       * 🔧 JSX escape issue (optional relax)
       */
      'react/no-unescaped-entities': [
        'warn',
        {
          forbid: ['>', '"'], // allow single quote without error spam
        },
      ],
    },
  },
]