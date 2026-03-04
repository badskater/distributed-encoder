import { useState, useEffect, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import * as api from '../api/client'
import type { Job } from '../types'
import StatusBadge from '../components/StatusBadge'
import ProgressBar from '../components/ProgressBar'
import { useAutoRefresh } from '../hooks/useAutoRefresh'

function basename(p: string) {
  return p.split(/[\\/]/).pop() ?? p
}

function fmtDate(s: string) {
  return new Date(s).toLocaleString()
}

const STATUSES = ['', 'queued', 'assigned', 'running', 'completed', 'failed', 'cancelled']

export default function Jobs() {
  const [jobs, setJobs] = useState<Job[]>([])
  const [status, setStatus] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const navigate = useNavigate()

  const load = useCallback(async () => {
    try {
      const j = await api.listJobs(status || undefined)
      setJobs(j)
      setError('')
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Failed to load')
    } finally {
      setLoading(false)
    }
  }, [status])

  useEffect(() => { setLoading(true); load() }, [load])
  useAutoRefresh(load)

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-th-text">Jobs</h1>
        <select
          value={status}
          onChange={e => setStatus(e.target.value)}
          className="bg-th-input-bg border border-th-input-border rounded px-3 py-1.5 text-sm text-th-text focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
          {STATUSES.map(s => (
            <option key={s} value={s}>{s || 'All statuses'}</option>
          ))}
        </select>
      </div>

      {error && <p className="text-red-600 text-sm">{error}</p>}

      {loading ? <p className="text-th-text-muted">Loading…</p> : (
        <div className="bg-th-surface rounded-lg shadow overflow-hidden">
          <table className="min-w-full divide-y divide-th-border text-sm">
            <thead className="bg-th-surface-muted">
              <tr>
                {['ID', 'Source', 'Status', 'Progress', 'Priority', 'Created'].map(h => (
                  <th key={h} className="px-4 py-2 text-left text-xs font-medium text-th-text-muted uppercase">{h}</th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-th-border-subtle">
              {jobs.map(j => (
                <tr
                  key={j.id}
                  onClick={() => navigate(`/jobs/${j.id}`)}
                  className="hover:bg-th-surface-muted cursor-pointer"
                >
                  <td className="px-4 py-2 font-mono text-blue-600">{j.id.slice(0, 8)}</td>
                  <td className="px-4 py-2 max-w-xs truncate text-th-text-secondary">{basename(j.source_path)}</td>
                  <td className="px-4 py-2"><StatusBadge status={j.status} /></td>
                  <td className="px-4 py-2 w-36">
                    <ProgressBar value={j.tasks_completed} max={j.tasks_total} />
                    <span className="text-xs text-th-text-subtle">{j.tasks_completed}/{j.tasks_total}</span>
                  </td>
                  <td className="px-4 py-2 text-th-text-secondary">{j.priority}</td>
                  <td className="px-4 py-2 text-th-text-muted whitespace-nowrap">{fmtDate(j.created_at)}</td>
                </tr>
              ))}
              {jobs.length === 0 && (
                <tr><td colSpan={6} className="px-4 py-4 text-center text-th-text-subtle">No jobs found</td></tr>
              )}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
