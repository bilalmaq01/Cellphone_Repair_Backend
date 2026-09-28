import { useEffect, useState } from 'react'
import { api } from '../api'

const EMPTY = { email: '', password: '', role: 'employee' }

export default function Employees() {
  const [employees, setEmployees] = useState([])
  const [form, setForm] = useState(EMPTY)
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  function load() {
    api.listEmployees().then(setEmployees).catch((err) => setError(err.message))
  }
  useEffect(load, [])

  function update(field) {
    return (e) => setForm({ ...form, [field]: e.target.value })
  }

  async function create(e) {
    e.preventDefault()
    setError('')
    setBusy(true)
    try {
      await api.createEmployee(form)
      setForm(EMPTY)
      load()
    } catch (err) {
      setError(err.message)
    } finally {
      setBusy(false)
    }
  }

  async function toggleActive(emp) {
    setError('')
    try {
      await api.updateEmployee(emp.id, { active: !emp.active })
      load()
    } catch (err) {
      setError(err.message)
    }
  }

  return (
    <div>
      <h1>Employees</h1>
      {error && <div className="error">{error}</div>}

      <div className="card">
        {employees.map((emp) => (
          <div key={emp.id} className="repair-row">
            <div>
              <div style={{ fontWeight: 600 }}>{emp.email}</div>
              <div className="muted">{emp.role}{emp.active ? '' : ' · deactivated'}</div>
            </div>
            <button
              className="secondary"
              style={{ width: 'auto', margin: 0 }}
              onClick={() => toggleActive(emp)}
            >
              {emp.active ? 'Deactivate' : 'Reactivate'}
            </button>
          </div>
        ))}
      </div>

      <div className="card">
        <h2>Add employee</h2>
        <form onSubmit={create}>
          <label>Email</label>
          <input type="email" value={form.email} onChange={update('email')} required />
          <label>Temporary password</label>
          <input value={form.password} onChange={update('password')} required />
          <label>Role</label>
          <select value={form.role} onChange={update('role')}>
            <option value="employee">Employee</option>
            <option value="admin">Admin</option>
          </select>
          <button disabled={busy}>{busy ? 'Creating…' : 'Create account'}</button>
        </form>
      </div>
    </div>
  )
}