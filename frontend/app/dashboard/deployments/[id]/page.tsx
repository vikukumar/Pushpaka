// Server-component wrapper: required for output: export with dynamic segments.
// Client logic lives in _Content.tsx which uses useParams() for dynamic routing.
import Content from './_Content'

export function generateStaticParams() {
  return [{ id: '_' }]
}

export default function Page() {
  return <Content />
}
