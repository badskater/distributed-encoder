interface Props { status: string }

const badgeVars: Record<string, [string, string]> = {
  running:          ['--th-badge-running-bg',  '--th-badge-running-text'],
  completed:        ['--th-badge-success-bg',  '--th-badge-success-text'],
  done:             ['--th-badge-success-bg',  '--th-badge-success-text'],
  failed:           ['--th-badge-error-bg',    '--th-badge-error-text'],
  error:            ['--th-badge-error-bg',    '--th-badge-error-text'],
  cancelled:        ['--th-badge-neutral-bg',  '--th-badge-neutral-text'],
  queued:           ['--th-badge-warning-bg',  '--th-badge-warning-text'],
  assigned:         ['--th-badge-assigned-bg', '--th-badge-assigned-text'],
  pending:          ['--th-badge-warning-bg',  '--th-badge-warning-text'],
  idle:             ['--th-badge-success-bg',  '--th-badge-success-text'],
  offline:          ['--th-badge-error-bg',    '--th-badge-error-text'],
  draining:         ['--th-badge-draining-bg', '--th-badge-draining-text'],
  pending_approval: ['--th-badge-approval-bg', '--th-badge-approval-text'],
  ready:            ['--th-badge-success-bg',  '--th-badge-success-text'],
  new:              ['--th-badge-neutral-bg',  '--th-badge-neutral-text'],
  analysing:        ['--th-badge-running-bg',  '--th-badge-running-text'],
  encoding:         ['--th-badge-assigned-bg', '--th-badge-assigned-text'],
}

export default function StatusBadge({ status }: Props) {
  const vars = badgeVars[status] ?? ['--th-badge-neutral-bg', '--th-badge-neutral-text']
  return (
    <span
      className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium"
      style={{ backgroundColor: `var(${vars[0]})`, color: `var(${vars[1]})` }}
    >
      {status.replace(/_/g, ' ')}
    </span>
  )
}
