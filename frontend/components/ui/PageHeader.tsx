import { Header } from '@/components/layout/Header'
import React from 'react'

interface PageHeaderProps {
  title: string
  description?: string
  icon?: React.ReactNode
  actions?: React.ReactNode
}

/**
 * PageHeader renders a sticky top bar using the shared Header component.
 * The icon (if any) is placed on the right side before the actions.
 */
export function PageHeader({ title, description, icon, actions }: PageHeaderProps) {
  const combinedActions = (
    <div className="flex items-center gap-2">
      {icon && (
        <span
          className="hidden md:flex items-center justify-center w-8 h-8 rounded-lg"
          style={{ background: 'rgba(99,102,241,0.1)', border: '1px solid rgba(99,102,241,0.2)' }}
        >
          {icon}
        </span>
      )}
      {actions}
    </div>
  )

  return <Header title={title} subtitle={description} actions={combinedActions} />
}
