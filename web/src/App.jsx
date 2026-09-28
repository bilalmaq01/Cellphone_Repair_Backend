import { Navigate, NavLink, Route, Routes, useNavigate } from 'react-router-dom'
import { clearAuth, getRole, isLoggedIn } from './api'
import Login from './pages/Login'
import Dashboard from './pages/Dashboard'
import Intake from './pages/Intake'
import RepairDetail from './pages/RepairDetail'
import Employees from './pages/Employees'
import StatusCheck from './pages/StatusCheck'

function RequireAuth({ children }) {
  return isLoggedIn() ? children : <Navigate to="/login" replace />
}

function RequireAdmin({ children }) {
  if (!isLoggedIn()) return <Navigate to="/login" replace />
  return getRole() === 'admin' ? children : <Navigate to="/dashboard" replace />
}

function Nav() {
  const navigate = useNavigate()
  if (!isLoggedIn()) return null

  function logout() {
    clearAuth()
    navigate('/login')
  }

  return (
    <nav className="nav">
      <NavLink to="/dashboard">Dashboard</NavLink>
      <NavLink to="/intake">New Intake</NavLink>
      {getRole() === 'admin' && <NavLink to="/employees">Employees</NavLink>}
      <span className="spacer" />
      <a href="#" onClick={(e) => { e.preventDefault(); logout() }}>Log out</a>
    </nav>
  )
}

export default function App() {
  return (
    <>
      <Nav />
      <div className="container">
        <Routes>
          <Route path="/" element={<Navigate to={isLoggedIn() ? '/dashboard' : '/login'} replace />} />
          <Route path="/login" element={<Login />} />
          <Route path="/status" element={<StatusCheck />} />
          <Route path="/dashboard" element={<RequireAuth><Dashboard /></RequireAuth>} />
          <Route path="/intake" element={<RequireAuth><Intake /></RequireAuth>} />
          <Route path="/repairs/:id" element={<RequireAuth><RepairDetail /></RequireAuth>} />
          <Route path="/employees" element={<RequireAdmin><Employees /></RequireAdmin>} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </div>
    </>
  )
}
