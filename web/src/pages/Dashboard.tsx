import { useState, useEffect, useCallback } from 'react'
import { Link } from 'react-router-dom'
import * as api from '../api/client'
import type { Job, Agent } from '../types'
import StatusBadge from '../components/StatusBadge'
import ProgressBar from '../components/ProgressBar'
import { useAutoRefresh } from '../hooks/useAutoRefresh'

function basename(p: string) {
  return p.split(/[\\/]/).pop() ?? p
}

function fmtDate(s: string) {
  return new Date(s).toLocaleString()
}

export default function Dashboard() {
  const [jobs, setJobs] = useState<Job[]>([])
  const [agents, setAgents] = useState<Agent[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const load = useCallback(async () => {
    try {
      const [j, a] = await Promise.all([api.listJobs(), api.listAgents()])
      setJobs(j)
      setAgents(a)
      setError('')
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Failed to load')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { load() }, [load])
  useAutoRefresh(load)

  const running = jobs.filter(j => j.status === 'running').length
  const idleAgents = agents.filter(a => a.status === 'idle').length
  const offlineAgents = agents.filter(a => a.status === 'offline').length
  const recent = [...jobs].sort((a, b) => b.created_at.localeCompare(a.created_at)).slice(0, 10)

  if (loading) return <p className="text-th-text-muted">Loading…</p>

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-th-text">Dashboard</h1>

      {error && <p className="text-red-600 text-sm">{error}</p>}

      <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
        {[
          { label: 'Total Jobs', value: jobs.length },
          { label: 'Running Jobs', value: running },
          { label: 'Idle Agents', value: idleAgents },
          { label: 'Offline Agents', value: offlineAgents },
        ].map(card => (
          <div key={card.label} className="bg-th-surface rounded-lg shadow p-4">
            <p className="text-sm text-th-text-muted">{card.label}</p>
            <p className="text-3xl font-bold text-th-text mt-1">{card.value}</p>
          </div>
        ))}
      </div>

      <div className="bg-th-surface rounded-lg shadow">
        <div className="px-4 py-3 border-b border-th-border">
          <h2 className="text-sm font-semibold text-th-text-secondary">Recent Jobs</h2>
        </div>
        <table className="min-w-full divide-y divide-th-border text-sm">
          <thead className="bg-th-surface-muted">
            <tr>
              {['ID', 'Source', 'Status', 'Progress', 'Created'].map(h => (
                <th key={h} className="px-4 py-2 text-left text-xs font-medium text-th-text-muted uppercase">{h}</th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-th-border-subtle">
            {recent.map(j => (
              <tr key={j.id} className="hover:bg-th-surface-muted">
                <td className="px-4 py-2 font-mono">
                  <Link to={`/jobs/${j.id}`} className="text-blue-600 hover:underline">{j.id.slice(0, 8)}</Link>
                </td>
                <td className="px-4 py-2 max-w-xs truncate text-th-text-secondary">{basename(j.source_path)}</td>
                <td className="px-4 py-2"><StatusBadge status={j.status} /></td>
                <td className="px-4 py-2 w-32">
                  <ProgressBar value={j.tasks_completed} max={j.tasks_total} />
                  <span className="text-xs text-th-text-subtle">{j.tasks_completed}/{j.tasks_total}</span>
                </td>
                <td className="px-4 py-2 text-th-text-muted whitespace-nowrap">{fmtDate(j.created_at)}</td>
              </tr>
            ))}
            {recent.length === 0 && (
              <tr><td colSpan={5} className="px-4 py-4 text-center text-th-text-subtle">No jobs yet</td></tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
}
