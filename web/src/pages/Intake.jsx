import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../api'
import SignaturePad from '../SignaturePad'

const EMPTY = { phone: '', name: '', email: '', device: '', issue: '' }

// Update these sample terms to match the shop's repair policy.
const TERMS =
  'I authorize the shop to perform the requested repair and diagnostic work. ' +
  'I understand data loss may occur and the shop is not liable for lost data. ' +
  'Devices not picked up within 90 days may be considered abandoned.'

export default function Intake() {
  const [form, setForm] = useState(EMPTY)
  const [signature, setSignature] = useState('')
  const [agreed, setAgreed] = useState(false)
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)
  const navigate = useNavigate()

  function update(field) {
    return (e) => setForm({ ...form, [field]: e.target.value })
  }

  async function submit(e) {
    e.preventDefault()
    setError('')
    if (!agreed) return setError('The customer must agree to the terms.')
    if (!signature) return setError('A customer signature is required.')
    setBusy(true)
    try {
      const repair = await api.intake({
        ...form,
        signature,
        terms: TERMS,
        signed_by_name: form.name,
      })
      navigate(`/repairs/${repair.id}`)
    } catch (err) {
      setError(err.message)
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="card">
      <h1>New Repair Intake</h1>
      <form onSubmit={submit}>
        <label>Customer phone *</label>
        <input value={form.phone} onChange={update('phone')} placeholder="(555) 123-4567" required />
        <label>Customer name *</label>
        <input value={form.name} onChange={update('name')} required />
        <label>Email (optional)</label>
        <input type="email" value={form.email} onChange={update('email')} />
        <label>Device *</label>
        <input value={form.device} onChange={update('device')} placeholder="iPhone 13" required />
        <label>Issue *</label>
        <textarea value={form.issue} onChange={update('issue')} placeholder="Describe the problem" required />

        <label>Repair authorization</label>
        <div className="muted" style={{ border: '1px solid var(--border)', borderRadius: 8, padding: 10, maxHeight: 120, overflow: 'auto' }}>
          {TERMS}
        </div>
        <label style={{ display: 'flex', gap: 8, alignItems: 'center', fontWeight: 400 }}>
          <input type="checkbox" style={{ width: 'auto' }} checked={agreed} onChange={(e) => setAgreed(e.target.checked)} />
          Customer agrees to the terms above
        </label>

        <label>Customer signature *</label>
        <SignaturePad onChange={setSignature} />

        {error && <div className="error">{error}</div>}
        <button disabled={busy}>{busy ? 'Saving…' : 'Create repair'}</button>
      </form>
    </div>
  )
}
