import { useState, useEffect, useCallback } from 'react'
import * as api from '../../api/client'
import type { Template } from '../../types'

function fmtDate(s: string) {
  return new Date(s).toLocaleString()
}

export default function Templates() {
  const [templates, setTemplates] = useState<Template[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [showForm, setShowForm] = useState(false)
  const [form, setForm] = useState({ name: '', type: 'bat' as Template['type'], description: '', content: '' })
  const [saving, setSaving] = useState(false)

  const load = useCallback(async () => {
    try {
      const t = await api.listTemplates()
      setTemplates(t)
      setError('')
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Failed to load')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { load() }, [load])

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    setSaving(true)
    try {
      await api.createTemplate(form)
      setShowForm(false)
      setForm({ name: '', type: 'bat', description: '', content: '' })
      load()
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Failed to create')
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async (id: string) => {
    if (!confirm('Delete this template?')) return
    try {
      await api.deleteTemplate(id)
      load()
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Failed to delete')
    }
  }

  if (loading) return <p className="text-th-text-muted">Loading…</p>

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-th-text">Templates</h1>
        <button
          onClick={() => setShowForm(!showForm)}
          className="bg-blue-600 text-white px-3 py-1.5 rounded text-sm font-medium hover:bg-blue-700"
        >
          {showForm ? 'Cancel' : 'Add Template'}
        </button>
      </div>

      {error && <p className="text-red-600 text-sm">{error}</p>}

      {showForm && (
        <form onSubmit={handleCreate} className="bg-th-surface rounded-lg shadow p-4 space-y-3">
          <h2 className="text-sm font-semibold text-th-text-secondary">New Template</h2>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-xs text-th-text-muted mb-1">Name</label>
              <input value={form.name} onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
                className="w-full bg-th-input-bg border border-th-input-border rounded px-2 py-1.5 text-sm text-th-text" required />
            </div>
            <div>
              <label className="block text-xs text-th-text-muted mb-1">Type</label>
              <select value={form.type} onChange={e => setForm(f => ({ ...f, type: e.target.value as Template['type'] }))}
                className="w-full bg-th-input-bg border border-th-input-border rounded px-2 py-1.5 text-sm text-th-text">
                <option value="bat">.bat</option>
                <option value="avs">.avs</option>
                <option value="vpy">.vpy</option>
              </select>
            </div>
          </div>
          <div>
            <label className="block text-xs text-th-text-muted mb-1">Description</label>
            <input value={form.description} onChange={e => setForm(f => ({ ...f, description: e.target.value }))}
              className="w-full bg-th-input-bg border border-th-input-border rounded px-2 py-1.5 text-sm text-th-text" />
          </div>
          <div>
            <label className="block text-xs text-th-text-muted mb-1">Content</label>
            <textarea value={form.content} onChange={e => setForm(f => ({ ...f, content: e.target.value }))}
              rows={8}
              className="w-full bg-th-input-bg border border-th-input-border rounded px-2 py-1.5 text-sm font-mono text-th-text" required />
          </div>
          <button type="submit" disabled={saving}
            className="bg-blue-600 text-white px-3 py-1.5 rounded text-sm font-medium hover:bg-blue-700 disabled:opacity-50">
            {saving ? 'Saving…' : 'Create Template'}
          </button>
        </form>
      )}

      <div className="bg-th-surface rounded-lg shadow overflow-hidden">
        <table className="min-w-full divide-y divide-th-border text-sm">
          <thead className="bg-th-surface-muted">
            <tr>
              {['Name', 'Type', 'Description', 'Created', ''].map(h => (
                <th key={h} className="px-4 py-2 text-left text-xs font-medium text-th-text-muted uppercase">{h}</th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-th-border-subtle">
            {templates.map(t => (
              <tr key={t.id} className="hover:bg-th-surface-muted">
                <td className="px-4 py-2 font-medium text-th-text">{t.name}</td>
                <td className="px-4 py-2">
                  <span className="font-mono text-xs bg-th-surface-muted px-1.5 py-0.5 rounded text-th-text-muted">.{t.type}</span>
                </td>
                <td className="px-4 py-2 text-th-text-muted">{t.description ?? '—'}</td>
                <td className="px-4 py-2 text-th-text-muted whitespace-nowrap">{fmtDate(t.created_at)}</td>
                <td className="px-4 py-2">
                  <button onClick={() => handleDelete(t.id)}
                    className="text-xs text-red-600 hover:underline">Delete</button>
                </td>
              </tr>
            ))}
            {templates.length === 0 && (
              <tr><td colSpan={5} className="px-4 py-4 text-center text-th-text-subtle">No templates</td></tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
}
