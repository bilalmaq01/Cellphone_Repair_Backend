import { useState } from 'react'
import { Link } from 'react-router-dom'
import { api, humanStatus } from '../api'

function badgeClass(status) {
  if (status === 'COMPLETED') return 'badge done'
  if (status === 'CANCELLED') return 'badge cancelled'
  return 'badge'
}

export default function StatusCheck() {
  const [step, setStep] = useState('phone') // phone -> code -> results
  const [phone, setPhone] = useState('')
  const [code, setCode] = useState('')
  const [repairs, setRepairs] = useState([])
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  async function requestCode(e) {
    e.preventDefault()
    setError(''); setBusy(true)
    try {
      await api.requestOtp(phone)
      setStep('code')
    } catch (err) {
      setError(err.message)
    } finally {
      setBusy(false)
    }
  }

  async function verify(e) {
    e.preventDefault()
    setError(''); setBusy(true)
    try {
      const { repairs } = await api.verifyOtp(phone, code)
      setRepairs(repairs)
      setStep('results')
    } catch (err) {
      setError(err.message)
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="card">
      <h1>Check Your Repair Status</h1>

      {step === 'phone' && (
        <form onSubmit={requestCode}>
          <label>Your phone number</label>
          <input value={phone} onChange={(e) => setPhone(e.target.value)} placeholder="(555) 123-4567" required />
          {error && <div className="error">{error}</div>}
          <button disabled={busy}>{busy ? 'Sending…' : 'Send me a code'}</button>
        </form>
      )}

      {step === 'code' && (
        <form onSubmit={verify}>
          <p className="muted">We texted a code to {phone}. Enter it below.</p>
          <label>Verification code</label>
          <input value={code} onChange={(e) => setCode(e.target.value)} inputMode="numeric" required />
          {error && <div className="error">{error}</div>}
          <button disabled={busy}>{busy ? 'Verifying…' : 'View my repairs'}</button>
        </form>
      )}

      {step === 'results' && (
        <div>
          {repairs.length === 0 ? (
            <p className="muted">No repairs found for that number.</p>
          ) : (
            repairs.map((r) => (
              <div key={r.id} className="repair-row">
                <div>
                  <div style={{ fontWeight: 600 }}>{r.device}</div>
                  <div className="muted">{r.issue}</div>
                </div>
                <span className={badgeClass(r.status)}>{humanStatus(r.status)}</span>
              </div>
            ))
          )}
        </div>
      )}

      <p className="muted" style={{ marginTop: 16 }}>
        <Link to="/login">Employee login</Link>
      </p>
    </div>
  )
}
