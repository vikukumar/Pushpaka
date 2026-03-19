'use client'

import React, { useEffect } from 'react'
import { motion } from 'framer-motion'
import { PushpakaViman } from '@/components/ui/PushpakaViman'
import { ErrorLayout } from '@/components/layout/ErrorLayout'
import { RefreshCcw, ShieldAlert } from 'lucide-react'

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string }
  reset: () => void
}) {
  useEffect(() => {
    console.error('Frontend Error:', error)
  }, [error])

  return (
    <ErrorLayout>
      <div className="space-y-8">
        <div className="relative flex justify-center h-64">
          <PushpakaViman state="collapse" />
          <motion.div 
            initial={{ scale: 0.8, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            className="absolute -top-4 -right-4 bg-red-500/20 text-red-400 p-3 rounded-2xl border border-red-500/30 backdrop-blur-xl animate-bounce"
          >
            <ShieldAlert size={32} />
          </motion.div>
        </div>

        <div className="space-y-4">
          <h1 className="text-4xl md:text-5xl font-bold bg-gradient-to-b from-white to-red-400 bg-clip-text text-transparent">
            Celestial Collapse
          </h1>
          <p className="text-slate-400 text-lg max-w-md mx-auto leading-relaxed">
            The Pushpaka Viman has encountered an internal turbulence. Our heavenly engineers have been notified.
          </p>
        </div>

        <div className="flex flex-col items-center gap-6 pt-4">
          <button 
            onClick={() => reset()}
            className="group flex items-center gap-2 px-8 py-3 bg-white text-slate-950 hover:bg-slate-200 rounded-full font-bold transition-all shadow-xl active:scale-95"
          >
            <RefreshCcw size={18} className="group-hover:rotate-180 transition-transform duration-500" />
            Repair Viman & Retry
          </button>
          
          <div className="flex flex-col items-center gap-2 opacity-40 hover:opacity-100 transition-opacity">
            <span className="text-[10px] uppercase tracking-[0.2em] font-bold text-slate-500">Error Identification</span>
            <code className="bg-white/5 border border-white/10 px-3 py-1 rounded text-[10px] text-slate-400 font-mono">
              {error.digest || 'Internal Server Turbulence'}
            </code>
          </div>
        </div>
      </div>
    </ErrorLayout>
  )
}
