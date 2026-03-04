import { useState, useEffect, useCallback } from 'react'
import { useParams, Link } from 'react-router-dom'
import * as api from '../api/client'
import type { Task, LogEntry } from '../types'
import StatusBadge from '../components/StatusBadge'
import { useAutoRefresh } from '../hooks/useAutoRefresh'

function fmtBytes(n: number) {
  if (n >= 1e9) return (n / 1e9).toFixed(1) + ' GB'
  if (n >= 1e6) return (n / 1e6).toFixed(1) + ' MB'
  if (n >= 1e3) return (n / 1e3).toFixed(1) + ' KB'
  return n + ' B'
}

function fmtDate(s: string | null) {
  return s ? new Date(s).toLocaleString() : '—'
}

function Row({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="flex py-2 border-b border-th-border-subtle last:border-0">
      <span className="w-40 text-sm text-th-text-muted shrink-0">{label}</span>
      <span className="text-sm text-th-text">{value}</span>
    </div>
  )
}

const streamColors: Record<string, string> = {
  stdout: 'text-green-400',
  stderr: 'text-red-400',
  agent: 'text-yellow-400',
}

export default function TaskDetail() {
  const { id } = useParams<{ id: string }>()
  const [task, setTask] = useState<Task | null>(null)
  const [logs, setLogs] = useState<LogEntry[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const load = useCallback(async () => {
    if (!id) return
    try {
      const [t, l] = await Promise.all([api.getTask(id), api.listTaskLogs(id)])
      setTask(t)
      setLogs(l)
      setError('')
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Failed to load')
    } finally {
      setLoading(false)
    }
  }, [id])

  useEffect(() => { load() }, [load])
  useAutoRefresh(load)

  if (loading) return <p className="text-th-text-muted">Loading…</p>
  if (error) return <p className="text-red-600">{error}</p>
  if (!task) return <p className="text-th-text-muted">Task not found</p>

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <Link to={`/jobs/${task.job_id}`} className="text-blue-600 hover:underline text-sm">← Job</Link>
        <h1 className="text-xl font-bold text-th-text">Task <span className="font-mono">{task.id.slice(0, 8)}…</span></h1>
        <StatusBadge status={task.status} />
      </div>

      <div className="bg-th-surface rounded-lg shadow p-4">
        <h2 className="text-sm font-semibold text-th-text-secondary mb-2">Task Details</h2>
        <Row label="ID" value={<span className="font-mono text-xs">{task.id}</span>} />
        <Row label="Job" value={<Link to={`/jobs/${task.job_id}`} className="text-blue-600 hover:underline font-mono text-xs">{task.job_id.slice(0, 8)}…</Link>} />
        <Row label="Chunk Index" value={task.chunk_index} />
        <Row label="Agent" value={task.agent_id ? <span className="font-mono text-xs">{task.agent_id}</span> : '—'} />
        <Row label="Exit Code" value={task.exit_code != null ? task.exit_code : '—'} />
        <Row label="Frames Encoded" value={task.frames_encoded > 0 ? task.frames_encoded.toLocaleString() : '—'} />
        <Row label="Avg FPS" value={task.avg_fps > 0 ? task.avg_fps.toFixed(1) : '—'} />
        <Row label="Output Size" value={task.output_size > 0 ? fmtBytes(task.output_size) : '—'} />
        <Row label="Started" value={fmtDate(task.started_at)} />
        <Row label="Completed" value={fmtDate(task.completed_at)} />
        {task.error_msg && <Row label="Error" value={<span className="text-red-600 text-xs">{task.error_msg}</span>} />}
      </div>

      <div className="bg-th-surface rounded-lg shadow">
        <div className="px-4 py-3 border-b border-th-border flex items-center justify-between">
          <h2 className="text-sm font-semibold text-th-text-secondary">Logs ({logs.length})</h2>
        </div>
        <div className="bg-th-log-bg rounded-b-lg overflow-auto max-h-[500px] p-4 font-mono text-xs">
          {logs.length === 0 ? (
            <span className="text-th-text-muted">No logs available</span>
          ) : (
            logs.map(entry => (
              <div key={entry.id} className="flex gap-2 leading-5">
                <span className="text-th-text-muted whitespace-nowrap shrink-0">
                  {new Date(entry.timestamp).toLocaleTimeString()}
                </span>
                <span className={`w-12 shrink-0 ${streamColors[entry.stream] ?? 'text-th-text-subtle'}`}>
                  {entry.stream}
                </span>
                <span className="text-th-log-text break-all">{entry.message}</span>
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  )
}
