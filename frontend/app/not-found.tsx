'use client'

import React from 'react'
import Link from 'next/link'
import { motion } from 'framer-motion'
import { PushpakaViman } from '@/components/ui/PushpakaViman'
import { ErrorLayout } from '@/components/layout/ErrorLayout'
import { Home, Search } from 'lucide-react'

export default function NotFound() {
  return (
    <ErrorLayout>
      <div className="space-y-8">
        <div className="relative flex justify-center h-64">
          <PushpakaViman state="lost" />
          <motion.div 
            initial={{ opacity: 0 }}
            animate={{ opacity: [0, 1, 0] }}
            transition={{ duration: 4, repeat: Infinity }}
            className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 text-white/5 font-black text-9xl pointer-events-none select-none"
          >
            404
          </motion.div>
        </div>

        <div className="space-y-4">
          <h1 className="text-4xl md:text-5xl font-bold bg-gradient-to-b from-white to-slate-400 bg-clip-text text-transparent">
            Pushpaka Viman is Lost
          </h1>
          <p className="text-slate-400 text-lg max-w-md mx-auto leading-relaxed">
            Even the most powerful chariot can lose its way. The heavenly path you're looking for doesn't exist.
          </p>
        </div>

        <div className="flex flex-col sm:flex-row items-center justify-center gap-4 pt-4">
          <Link 
            href="/dashboard" 
            className="group flex items-center gap-2 px-6 py-3 bg-brand-600 hover:bg-brand-500 text-white rounded-full font-semibold transition-all shadow-lg shadow-brand-500/20 active:scale-95"
          >
            <Home size={18} />
            Back to Dashboard
          </Link>
          <button 
            onClick={() => window.history.back()}
            className="flex items-center gap-2 px-6 py-3 bg-white/5 hover:bg-white/10 text-[#cccccc] rounded-full border border-white/10 transition-all active:scale-95"
          >
            Go Back
          </button>
        </div>
      </div>
    </ErrorLayout>
  )
}
