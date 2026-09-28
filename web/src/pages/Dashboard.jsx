import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, humanStatus } from '../api'

function badgeClass(status) {
  if (status === 'COMPLETED') return 'badge done'
  if (status === 'CANCELLED') return 'badge cancelled'
  return 'badge'
}

export default function Dashboard() {
  const [openOnly, setOpenOnly] = useState(true)
  const [repairs, setRepairs] = useState([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    let active = true
    setLoading(true)
    api
      .listRepairs(openOnly)
      .then((data) => active && setRepairs(data))
      .catch((err) => active && setError(err.message))
      .finally(() => active && setLoading(false))
    return () => { active = false }
  }, [openOnly])

  return (
    <div>
      <h1>Repairs</h1>
      <div className="toggle">
        <button className={openOnly ? 'active' : ''} onClick={() => setOpenOnly(true)}>Open</button>
        <button className={!openOnly ? 'active' : ''} onClick={() => setOpenOnly(false)}>All</button>
      </div>

      {error && <div className="error">{error}</div>}
      {loading ? (
        <p className="muted">Loading…</p>
      ) : repairs.length === 0 ? (
        <div className="card"><p className="muted">No repairs {openOnly ? 'open' : 'yet'}.</p></div>
      ) : (
        <div className="card">
          {repairs.map((r) => (
            <Link key={r.id} to={`/repairs/${r.id}`} className="repair-row">
              <div>
                <div style={{ fontWeight: 600 }}>#{r.id} · {r.device}</div>
                <div className="muted">{r.customer_phone} — {r.issue_description}</div>
              </div>
              <span className={badgeClass(r.status)}>{humanStatus(r.status)}</span>
            </Link>
          ))}
        </div>
      )}
    </div>
  )
}
