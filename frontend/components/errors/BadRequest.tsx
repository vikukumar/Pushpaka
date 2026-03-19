'use client'

import React from 'react'
import { motion } from 'framer-motion'
import { PushpakaViman } from '@/components/ui/PushpakaViman'
import { ErrorLayout } from '@/components/layout/ErrorLayout'
import { ArrowLeftRight, CornerUpLeft } from 'lucide-react'

export default function BadRequest() {
  return (
    <ErrorLayout>
      <div className="space-y-8">
        <div className="relative flex justify-center h-64">
          <PushpakaViman state="wrong-way" />
          <motion.div 
            initial={{ rotate: 0 }}
            animate={{ rotate: 360 }}
            transition={{ duration: 10, repeat: Infinity, ease: "linear" }}
            className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 border-4 border-dashed border-white/5 w-80 h-80 rounded-full"
          />
        </div>

        <div className="space-y-4">
          <h1 className="text-4xl md:text-5xl font-bold bg-gradient-to-b from-white to-orange-400 bg-clip-text text-transparent uppercase tracking-tight">
            Wrong Direction
          </h1>
          <p className="text-slate-400 text-lg max-w-md mx-auto leading-relaxed">
            The heavenly coordinates provided are invalid. Your request has steered the Pushpaka Viman off course.
          </p>
        </div>

        <div className="flex flex-col items-center gap-6 pt-4">
          <button 
            onClick={() => window.history.back()}
            className="group flex items-center gap-3 px-8 py-3 bg-orange-500/10 hover:bg-orange-500/20 text-orange-400 rounded-full font-bold transition-all border border-orange-500/30 active:scale-95"
          >
            <CornerUpLeft size={20} className="group-hover:-translate-x-1 transition-transform" />
            Recalculate & Go Back
          </button>
          
          <div className="flex items-center gap-2 text-[10px] text-slate-600 font-mono uppercase tracking-widest">
            <ArrowLeftRight size={12} />
            HTTP 400 Error: Bad Request
          </div>
        </div>
      </div>
    </ErrorLayout>
  )
}
