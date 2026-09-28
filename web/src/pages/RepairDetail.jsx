import { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { api, STATUSES, humanStatus } from '../api'

export default function RepairDetail() {
  const { id } = useParams()
  const [repair, setRepair] = useState(null)
  const [authz, setAuthz] = useState(null)
  const [status, setStatus] = useState('')
  const [error, setError] = useState('')
  const [msg, setMsg] = useState('')
  const [busy, setBusy] = useState(false)

  useEffect(() => {
    api
      .getRepair(id)
      .then((r) => { setRepair(r); setStatus(r.status) })
      .catch((err) => setError(err.message))
    // Load the signed authorization (404 = none captured; not an error).
    api
      .getAuthorization(id)
      .then(setAuthz)
      .catch(() => setAuthz(null))
  }, [id])

  async function saveStatus(e) {
    e.preventDefault()
    setError('')
    setMsg('')
    setBusy(true)
    try {
      const updated = await api.updateStatus(id, status)
      setRepair(updated)
      setMsg('Status updated — customer notified.')
    } catch (err) {
      setError(err.message)
    } finally {
      setBusy(false)
    }
  }

  if (error && !repair) return <div className="card"><div className="error">{error}</div></div>
  if (!repair) return <p className="muted">Loading…</p>

  return (
    <div>
      <p><Link to="/dashboard">← Back to dashboard</Link></p>
      <div className="card">
        <h1>Repair #{repair.id}</h1>
        <p><strong>{repair.device}</strong></p>
        <p className="muted">{repair.customer_phone}</p>
        <p>{repair.issue_description}</p>
        <p>Current status: <span className="badge">{humanStatus(repair.status)}</span></p>
      </div>

      {authz && (
        <div className="card">
          <h2>Signed authorization</h2>
          <p className="muted">
            Signed by <strong>{authz.signed_by_name}</strong> on{' '}
            {new Date(authz.signed_at).toLocaleString()}
          </p>
          {authz.signature_url ? (
            <img
              src={authz.signature_url}
              alt="Customer signature"
              style={{ maxWidth: '100%', border: '1px solid var(--border)', borderRadius: 8, background: '#fff' }}
            />
          ) : (
            <p className="muted">(signature image unavailable)</p>
          )}
          {authz.terms_text && (
            <details style={{ marginTop: 10 }}>
              <summary className="muted">Terms agreed to</summary>
              <p className="muted" style={{ marginTop: 8 }}>{authz.terms_text}</p>
            </details>
          )}
        </div>
      )}

      <div className="card">
        <h2>Update status</h2>
        <form onSubmit={saveStatus}>
          <select value={status} onChange={(e) => setStatus(e.target.value)}>
            {STATUSES.map((s) => (
              <option key={s} value={s}>{humanStatus(s)}</option>
            ))}
          </select>
          {error && <div className="error">{error}</div>}
          {msg && <div className="success">{msg}</div>}
          <button disabled={busy || status === repair.status}>
            {busy ? 'Saving…' : 'Update & notify customer'}
          </button>
        </form>
      </div>
    </div>
  )
}
