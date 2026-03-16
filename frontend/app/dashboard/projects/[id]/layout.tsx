// This layout covers all routes under /dashboard/projects/[id]/.
// generateStaticParams here tells Next.js there are no pre-renderable IDs;
// the Go SPA handler (index.html fallback) handles routing at runtime.
export function generateStaticParams() {
  return [{ id: '_' }]
}

export default function ProjectLayout({ children }: { children: React.ReactNode }) {
  return <>{children}</>
}
