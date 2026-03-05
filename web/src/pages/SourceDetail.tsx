import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import * as api from '../api/client'
import type { Source, AnalysisResult, AnalysisFramePoint } from '../types'
import StatusBadge from '../components/StatusBadge'

function fmtBytes(n: number) {
  if (n >= 1e9) return (n / 1e9).toFixed(1) + ' GB'
  if (n >= 1e6) return (n / 1e6).toFixed(1) + ' MB'
  if (n >= 1e3) return (n / 1e3).toFixed(1) + ' KB'
  return n + ' B'
}

function fmtDuration(sec: number | null | undefined) {
  if (sec == null) return '—'
  const h = Math.floor(sec / 3600)
  const m = Math.floor((sec % 3600) / 60)
  const s = Math.floor(sec % 60)
  return h > 0 ? `${h}h ${m}m ${s}s` : `${m}m ${s}s`
}

function Row({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="flex py-2 border-b border-th-border-subtle last:border-0">
      <span className="w-40 text-sm text-th-text-muted shrink-0">{label}</span>
      <span className="text-sm text-th-text">{value}</span>
    </div>
  )
}

function VmafSparkline({ data }: { data: AnalysisFramePoint[] }) {
  if (!data.length) return null

  const W = 600
  const H = 60
  const PAD = 2

  // Downsample to at most 300 points for rendering performance.
  const step = Math.ceil(data.length / 300)
  const sampled = data.filter((_, i) => i % step === 0)

  const scores = sampled.map(p => p.score ?? 0)
  const minScore = Math.min(...scores)
  const maxScore = Math.max(...scores)
  const range = maxScore - minScore || 1

  const xScale = (W - PAD * 2) / (sampled.length - 1 || 1)
  const yScale = (H - PAD * 2) / range

  const points = sampled
    .map((p, i) => `${PAD + i * xScale},${H - PAD - ((p.score ?? 0) - minScore) * yScale}`)
    .join(' ')

  // Draw a reference line at score=90 if it falls within the visible range.
  const refY = H - PAD - (90 - minScore) * yScale
  const showRef = refY > PAD && refY < H - PAD

  return (
    <div className="mt-3">
      <p className="text-xs text-th-text-muted mb-1">
        Per-frame VMAF ({sampled.length.toLocaleString()} samples
        {data.length > sampled.length ? `, 1 per ${step} frames` : ''})
      </p>
      <svg viewBox={`0 0 ${W} ${H}`} className="w-full h-16 bg-th-surface-muted rounded">
        {showRef && (
          <line
            x1={PAD} y1={refY} x2={W - PAD} y2={refY}
            stroke="currentColor" strokeOpacity="0.25" strokeDasharray="4 3"
            className="text-th-border"
          />
        )}
        <polyline
          fill="none"
          stroke="currentColor"
          strokeWidth="1.5"
          className="text-blue-500"
          points={points}
        />
      </svg>
      {showRef && (
        <p className="text-xs text-th-text-subtle mt-0.5">Dashed line = score 90 (reference quality)</p>
      )}
    </div>
  )
}

export default function SourceDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [source, setSource] = useState<Source | null>(null)
  const [analysis, setAnalysis] = useState<AnalysisResult | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!id) return
    Promise.all([api.getSource(id), api.getAnalysis(id).catch(() => null)])
      .then(([s, a]) => { setSource(s); setAnalysis(a) })
      .catch(e => setError(e instanceof Error ? e.message : 'Failed to load'))
      .finally(() => setLoading(false))
  }, [id])

  if (loading) return <p className="text-th-text-muted">Loading…</p>
  if (error) return <p className="text-red-600">{error}</p>
  if (!source) return <p className="text-th-text-muted">Source not found</p>

  const sum = analysis?.summary

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <button onClick={() => navigate('/sources')} className="text-blue-600 hover:underline text-sm">← Sources</button>
        <h1 className="text-2xl font-bold text-th-text">{source.filename}</h1>
      </div>

      <div className="bg-th-surface rounded-lg shadow p-4 space-y-1">
        <h2 className="text-sm font-semibold text-th-text-secondary mb-2">Source Info</h2>
        <Row label="Filename" value={source.filename} />
        <Row label="Path" value={<span className="font-mono text-xs break-all">{source.path}</span>} />
        <Row label="Size" value={fmtBytes(source.size_bytes)} />
        <Row label="Duration" value={fmtDuration(source.duration_sec)} />
        <Row label="State" value={<StatusBadge status={source.state} />} />
        <Row label="VMAF Score" value={source.vmaf_score != null ? source.vmaf_score.toFixed(2) : '—'} />
      </div>

      {analysis ? (
        <div className="bg-th-surface rounded-lg shadow p-4">
          <h2 className="text-sm font-semibold text-th-text-secondary mb-2">
            Analysis Results
            <span className="ml-2 font-normal text-th-text-subtle">({analysis.type})</span>
          </h2>

          {sum && (
            <div className="space-y-0">
              {sum.mean != null && (
                <Row label="VMAF Mean" value={sum.mean.toFixed(2)} />
              )}
              {sum.min != null && (
                <Row label="VMAF Min" value={sum.min.toFixed(2)} />
              )}
              {sum.max != null && (
                <Row label="VMAF Max" value={sum.max.toFixed(2)} />
              )}
              {sum.psnr != null && (
                <Row label="PSNR" value={sum.psnr.toFixed(2) + ' dB'} />
              )}
              {sum.ssim != null && (
                <Row label="SSIM" value={sum.ssim.toFixed(4)} />
              )}
              {sum.width != null && sum.height != null && (
                <Row label="Resolution" value={`${sum.width}×${sum.height}`} />
              )}
              {sum.duration_sec != null && (
                <Row label="Duration" value={fmtDuration(sum.duration_sec)} />
              )}
              {sum.frame_count != null && (
                <Row label="Frame Count" value={sum.frame_count.toLocaleString()} />
              )}
              {sum.codec && (
                <Row label="Codec" value={sum.codec} />
              )}
              {sum.bit_rate != null && (
                <Row label="Bit Rate" value={fmtBytes(sum.bit_rate) + '/s'} />
              )}
              {sum.scene_count != null && (
                <Row label="Scene Count" value={sum.scene_count.toLocaleString()} />
              )}
            </div>
          )}

          {analysis.frame_data && analysis.frame_data.length > 0 && (
            <VmafSparkline data={analysis.frame_data} />
          )}

          {!sum && (!analysis.frame_data || analysis.frame_data.length === 0) && (
            <p className="text-sm text-th-text-muted">No summary data available for this analysis.</p>
          )}

          <p className="text-xs text-th-text-subtle mt-3">
            Recorded {new Date(analysis.created_at).toLocaleString()}
          </p>
        </div>
      ) : (
        <div className="bg-th-surface rounded-lg shadow p-4">
          <h2 className="text-sm font-semibold text-th-text-secondary mb-1">Analysis Results</h2>
          <p className="text-sm text-th-text-muted">No analysis results yet. Run an analysis job to populate this section.</p>
        </div>
      )}
    </div>
  )
}
