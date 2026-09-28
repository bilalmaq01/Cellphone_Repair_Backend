// API client for the repair-shop backend, plus small auth helpers backed by
// localStorage. Paths are same-origin (/api/*) and proxied to the Go server.

const TOKEN_KEY = 'employee_token'
const ROLE_KEY = 'employee_role'

export function setAuth(token, role) {
  localStorage.setItem(TOKEN_KEY, token)
  if (role) localStorage.setItem(ROLE_KEY, role)
}
export function getToken() {
  return localStorage.getItem(TOKEN_KEY)
}
export function getRole() {
  return localStorage.getItem(ROLE_KEY)
}
export function clearAuth() {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(ROLE_KEY)
}
export function isLoggedIn() {
  return !!getToken()
}

export const STATUSES = [
  'RECEIVED',
  'DIAGNOSING',
  'AWAITING_APPROVAL',
  'IN_PROGRESS',
  'READY_FOR_PICKUP',
  'COMPLETED',
  'CANCELLED',
]

export function humanStatus(s) {
  return (s || '').toLowerCase().replaceAll('_', ' ')
}

async function request(method, path, { body, token } = {}) {
  const headers = {}
  if (body) headers['Content-Type'] = 'application/json'
  if (token) headers['Authorization'] = 'Bearer ' + token

  const res = await fetch(path, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  })
  const text = await res.text()
  const data = text ? JSON.parse(text) : null
  if (!res.ok) {
    const msg = (data && data.error) || `request failed (${res.status})`
    throw new Error(msg)
  }
  return data
}

export const api = {
  // Employee auth
  login: (email, password) => request('POST', '/api/auth/login', { body: { email, password } }),
  me: () => request('GET', '/api/auth/me', { token: getToken() }),

  // Repairs (employee)
  listRepairs: (openOnly) =>
    request('GET', `/api/repairs${openOnly ? '?open=true' : ''}`, { token: getToken() }),
  getRepair: (id) => request('GET', `/api/repairs/${id}`, { token: getToken() }),
  getAuthorization: (id) => request('GET', `/api/repairs/${id}/authorization`, { token: getToken() }),
  updateStatus: (id, status) =>
    request('PATCH', `/api/repairs/${id}/status`, { token: getToken(), body: { status } }),
  intake: (data) => request('POST', '/api/intake', { token: getToken(), body: data }),
  lookupCustomer: (phone) =>
    request('POST', '/api/intake/lookup', { token: getToken(), body: { phone } }),

  // Employees (admin)
  listEmployees: () => request('GET', '/api/employees', { token: getToken() }),
  createEmployee: (data) => request('POST', '/api/employees', { token: getToken(), body: data }),
  updateEmployee: (id, data) =>
    request('PATCH', `/api/employees/${id}`, { token: getToken(), body: data }),

  // Customer status check (public / customer token)
  requestOtp: (phone) => request('POST', '/api/status/request-otp', { body: { phone } }),
  verifyOtp: (phone, code) => request('POST', '/api/status/verify-otp', { body: { phone, code } }),
  customerRepairs: (custToken) => request('GET', '/api/status/repairs', { token: custToken }),
}
