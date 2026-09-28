import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { api, setAuth } from '../api'

export default function Login() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)
  const navigate = useNavigate()

  // Set when a request auto-logs the user out on an expired token.
  const expired = new URLSearchParams(window.location.search).get('expired') === '1'

  async function submit(e) {
    e.preventDefault()
    setError('')
    setBusy(true)
    try {
      const { token } = await api.login(email, password)
      setAuth(token)
      const me = await api.me()
      setAuth(token, me.role)
      navigate('/dashboard')
    } catch (err) {
      setError(err.message)
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="card">
      <h1>Employee Login</h1>
      {expired && <div className="error">Your session expired. Please sign in again.</div>}
      <form onSubmit={submit}>
        <label>Email</label>
        <input type="email" value={email} onChange={(e) => setEmail(e.target.value)} required />
        <label>Password</label>
        <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} required />
        {error && <div className="error">{error}</div>}
        <button disabled={busy}>{busy ? 'Signing in…' : 'Sign in'}</button>
      </form>
      <p className="muted" style={{ marginTop: 16 }}>
        Are you a customer? <Link to="/status">Check your repair status</Link>
      </p>
    </div>
  )
}
