// Layout for individual deployment detail pages.
export function generateStaticParams() {
  return [{ id: '_' }]
}

export default function DeploymentLayout({ children }: { children: React.ReactNode }) {
  return <>{children}</>
}
