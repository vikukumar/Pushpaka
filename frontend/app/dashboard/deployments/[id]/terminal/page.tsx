import TerminalWrapper from './TerminalWrapper'

export function generateStaticParams() {
  return [{ id: '_' }]
}

export default function TerminalPage() {
  return <TerminalWrapper />
}
