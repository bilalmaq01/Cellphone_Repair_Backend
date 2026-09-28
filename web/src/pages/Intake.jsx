import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api, humanStatus } from '../api'
import SignaturePad from '../SignaturePad'
import PhotoCapture from '../PhotoCapture'

const EMPTY = { phone: '', first_name: '', email: '', device: '', issue: '' }

// Update these sample terms to match the shop's repair policy.
const TERMS =
  'I authorize the shop to perform the requested repair and diagnostic work. ' +
  'I understand data loss may occur and the shop is not liable for lost data. ' +
  'Devices not picked up within 90 days may be considered abandoned.'

export default function Intake() {
  const [form, setForm] = useState(EMPTY)
  const [step, setStep] = useState('phone')
  const [devices, setDevices] = useState([])
  const [customerFound, setCustomerFound] = useState(false)
  const [signature, setSignature] = useState('')
  const [frontPhoto, setFrontPhoto] = useState('')
  const [backPhoto, setBackPhoto] = useState('')
  const [agreed, setAgreed] = useState(false)
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)
  const navigate = useNavigate()

  function update(field) {
    return (e) => setForm((current) => ({ ...current, [field]: e.target.value }))
  }

  async function lookup(e) {
    e.preventDefault()
    setError('')
    setBusy(true)
    try {
      const result = await api.lookupCustomer(form.phone)
      const customer = result.customer || {}
      setCustomerFound(result.found)
      setDevices(result.devices || [])
      setForm((current) => ({
        ...current,
        phone: result.phone || current.phone,
        first_name: customer.first_name || '',
        email: customer.email || '',
        device: '',
      }))
      setStep('device')
    } catch (err) {
      setError(err.message)
    } finally {
      setBusy(false)
    }
  }

  function chooseDevice(name) {
    setForm((current) => ({ ...current, device: name }))
  }

  function continueToRepair(e) {
    e.preventDefault()
    if (!form.first_name.trim()) return setError('Enter the customer’s first name.')
    if (!form.device.trim()) return setError('Choose a device or enter its model.')
    setError('')
    setStep('repair')
  }

  async function submit(e) {
    e.preventDefault()
    setError('')
    if (!frontPhoto || !backPhoto) return setError('Add photos of the front and back of the phone.')
    if (!agreed) return setError('The customer must agree to the terms.')
    if (!signature) return setError('A customer signature is required.')
    setBusy(true)
    try {
      const repair = await api.intake({
        ...form,
        signature,
        front_photo: frontPhoto,
        back_photo: backPhoto,
        terms: TERMS,
        signed_by_name: form.first_name,
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
      <p className="muted">Step {step === 'phone' ? '1' : step === 'device' ? '2' : '3'} of 3</p>

      {step === 'phone' && (
        <form onSubmit={lookup}>
          <label htmlFor="customer-phone">Customer phone *</label>
          <input id="customer-phone" type="tel" value={form.phone} onChange={update('phone')} placeholder="(555) 123-4567" required />
          {error && <div className="error">{error}</div>}
          <button disabled={busy}>{busy ? 'Looking up…' : 'Find customer'}</button>
        </form>
      )}

      {step === 'device' && (
        <form onSubmit={continueToRepair}>
          <p className="muted">{customerFound ? 'Customer profile found.' : 'No profile found. Create one with the details below.'}</p>
          <label htmlFor="first-name">First name *</label>
          <input id="first-name" value={form.first_name} onChange={update('first_name')} required />
          <label htmlFor="email">Email (optional)</label>
          <input id="email" type="email" value={form.email} onChange={update('email')} />

          {devices.length > 0 && (
            <>
              <h2 className="section-heading">Devices on this account</h2>
              <div className="device-list">
                {devices.map((device) => (
                  <button
                    type="button"
                    key={device.id}
                    className={`device-choice${form.device === device.model_name ? ' selected' : ''}`}
                    onClick={() => chooseDevice(device.model_name)}
                  >
                    <strong>{device.model_name}</strong>
                    {device.repairs?.length > 0 ? (
                      <span className="device-history">
                        {device.repairs.map((repair) => (
                          <span key={repair.id}>
                            {new Date(repair.created_at).toLocaleDateString()} · {humanStatus(repair.status)} · {repair.issue_description}
                          </span>
                        ))}
                      </span>
                    ) : <span className="device-history">No past repairs</span>}
                  </button>
                ))}
              </div>
            </>
          )}

          <label htmlFor="device">{devices.length ? 'Or add another device model' : 'Device model *'}</label>
          <input id="device" value={form.device} onChange={update('device')} placeholder="iPhone 12" required />
          {error && <div className="error">{error}</div>}
          <button type="button" className="secondary" onClick={() => { setStep('phone'); setError('') }}>Back to phone</button>
          <button>Continue</button>
        </form>
      )}

      {step === 'repair' && (
        <form onSubmit={submit}>
          <p className="muted">{form.first_name} · {form.phone} · {form.device}</p>
          <label htmlFor="issue">Issue *</label>
          <textarea id="issue" value={form.issue} onChange={update('issue')} placeholder="Describe the problem" required />

          <h2 className="section-heading">Phone photos</h2>
          <p className="muted">Take clear photos of the front and back of the phone.</p>
          <div className="photo-fields">
            <PhotoCapture label="Front" value={frontPhoto} onChange={setFrontPhoto} />
            <PhotoCapture label="Back" value={backPhoto} onChange={setBackPhoto} />
          </div>

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
          <button type="button" className="secondary" onClick={() => { setStep('device'); setError('') }}>Back to device</button>
          <button disabled={busy}>{busy ? 'Saving…' : 'Create repair'}</button>
        </form>
      )}
    </div>
  )
}
