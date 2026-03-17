import { AuthProvider } from '@/components/providers/AuthProvider'
import AIChatbot from '@/components/chat/AIChatbot'

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <AuthProvider>
      {children}
      <AIChatbot />
    </AuthProvider>
  )
}
