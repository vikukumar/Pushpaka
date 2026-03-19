'use client'

import { motion, Variants } from 'framer-motion'
import { cn } from '@/lib/utils'

export type VimanState = 'idle' | 'lost' | 'collapse' | 'wrong-way'

interface PushpakaVimanProps {
  state?: VimanState
  className?: string
}

export function PushpakaViman({ state = 'idle', className }: PushpakaVimanProps) {
  // Variants for the overall container
  const containerVariants: Variants = {
    idle: {
      y: [0, -20, 0],
      rotate: [0, 1, -1, 0],
      transition: {
        duration: 4,
        repeat: Infinity,
        ease: "easeInOut"
      }
    },
    lost: {
      x: [0, 100, -100, 200, -200],
      y: [0, -50, 50, -100, 100],
      opacity: [1, 0.8, 0.5, 0.2, 0],
      scale: [1, 0.9, 0.8, 0.7, 0.6],
      transition: {
        duration: 8,
        repeat: Infinity,
        ease: "linear"
      }
    },
    collapse: {
      rotate: [0, 5, -10, 20, -30, 45],
      y: [0, 10, -10, 50, 100, 500],
      opacity: [1, 1, 1, 0.8, 0.5, 0],
      transition: {
        duration: 3,
        ease: "easeIn"
      }
    },
    'wrong-way': {
      rotateY: [0, 180],
      x: [0, 1000],
      transition: {
        rotateY: { duration: 0.5 },
        x: { duration: 2, delay: 0.5, ease: "easeIn" }
      }
    }
  }

  return (
    <motion.div
      variants={containerVariants}
      animate={state}
      className={cn("relative w-64 h-64", className)}
    >
      {/* Glow effect */}
      <div className="absolute inset-0 bg-brand-500/20 blur-3xl rounded-full scale-75 animate-pulse" />
      
      <svg viewBox="0 0 500 500" className="w-full h-full drop-shadow-2xl">
        <defs>
          <linearGradient id="gold" x1="0%" y1="0%" x2="100%" y2="100%">
            <stop offset="0%" style={{ stopColor: '#FFD700', stopOpacity: 1 }} />
            <stop offset="50%" style={{ stopColor: '#FDB931', stopOpacity: 1 }} />
            <stop offset="100%" style={{ stopColor: '#B8860B', stopOpacity: 1 }} />
          </linearGradient>
          <filter id="shadow">
            <feDropShadow dx="0" dy="4" stdDeviation="4" floodOpacity="0.5" />
          </filter>
        </defs>

        {/* Main Body (Lotus Base) */}
        <motion.path
          d="M100 350 Q250 450 400 350 L420 300 Q250 350 80 300 Z"
          fill="url(#gold)"
          initial={{ y: 0 }}
          animate={state === 'collapse' ? { y: 20, rotate: 5 } : {}}
        />

        {/* Central Palace Structure */}
        <motion.path
          d="M150 300 L150 200 Q250 100 350 200 L350 300"
          fill="url(#gold)"
          stroke="#B8860B"
          strokeWidth="2"
        />

        {/* Dome */}
        <motion.path
          d="M180 200 Q250 50 320 200 Z"
          fill="#FFD700"
          opacity="0.8"
        />

        {/* Pillars */}
        <line x1="180" y1="300" x2="180" y2="200" stroke="#B8860B" strokeWidth="4" />
        <line x1="250" y1="300" x2="250" y2="180" stroke="#B8860B" strokeWidth="4" />
        <line x1="320" y1="300" x2="320" y2="200" stroke="#B8860B" strokeWidth="4" />

        {/* Windows */}
        <rect x="235" y="220" width="30" height="40" rx="5" fill="#4B3621" />
        <circle cx="210" cy="160" r="10" fill="#4B3621" />
        <circle cx="290" cy="160" r="10" fill="#4B3621" />

        {/* Flags */}
        <motion.path
          d="M320 100 L350 115 L320 130 Z"
          fill="#FF4500"
          animate={{
            rotate: [0, 10, -10, 0],
            skewX: [0, 5, -5, 0]
          }}
          transition={{ duration: 2, repeat: Infinity }}
        />

        {/* Ornaments */}
        <circle cx="250" cy="80" r="5" fill="#FFFFFF" className="animate-pulse" />
        
        {/* Clouds around the Viman */}
        <motion.g
          animate={{ x: [-10, 10, -10] }}
          transition={{ duration: 3, repeat: Infinity }}
        >
          <circle cx="80" cy="380" r="20" fill="white" opacity="0.6" />
          <circle cx="110" cy="400" r="30" fill="white" opacity="0.4" />
          <circle cx="390" cy="380" r="25" fill="white" opacity="0.6" />
          <circle cx="430" cy="400" r="35" fill="white" opacity="0.4" />
        </motion.g>
      </svg>
    </motion.div>
  )
}
