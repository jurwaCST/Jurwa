import { Route, Routes } from 'react-router-dom'
import { AuthProvider } from './auth/AuthContext'
import RequireRole from './auth/RequireRole'
import Layout from './components/Layout'
import AdminDashboard from './pages/AdminDashboard'
import Home from './pages/Home'
import Login from './pages/Login'
import NotFound from './pages/NotFound'
import MyReports from './pages/MyReports'
import PresidentDashboard from './pages/PresidentDashboard'
import Register from './pages/Register'
import ReportDetail from './pages/ReportDetail'
import Notifications from './pages/Notifications'
import ReportForm from './pages/ReportForm'
import StaffView from './pages/StaffView'

/**
 * App holds the routes. The router itself is in main.tsx (BrowserRouter) so
 * tests can use a MemoryRouter instead.
 */
export default function App() {
  return (
    <AuthProvider>
      <Routes>
        <Route element={<Layout />}>
          <Route index element={<Home />} />
          <Route path="login" element={<Login />} />
          <Route path="register" element={<Register />} />
          <Route
            path="report/new"
            element={
              <RequireRole roles={['STUDENT']}>
                <ReportForm />
              </RequireRole>
            }
          />
          <Route
            path="reports"
            element={
              <RequireRole roles={['STUDENT']}>
                <MyReports />
              </RequireRole>
            }
          />
          <Route
            path="reports/:reportId"
            element={
              <RequireRole roles={['STUDENT']}>
                <ReportDetail />
              </RequireRole>
            }
          />
          <Route
            path="notifications"
            element={
              <RequireRole roles={['STUDENT']}>
                <Notifications />
              </RequireRole>
            }
          />
          <Route
            path="staff"
            element={
              <RequireRole roles={['CATEGORY_HEAD', 'STAFF']}>
                <StaffView />
              </RequireRole>
            }
          />
          <Route
            path="admin"
            element={
              <RequireRole roles={['SYSTEM_ADMIN']}>
                <AdminDashboard />
              </RequireRole>
            }
          />
          <Route
            path="president"
            element={
              <RequireRole roles={['PRESIDENT']}>
                <PresidentDashboard />
              </RequireRole>
            }
          />
          <Route path="*" element={<NotFound />} />
        </Route>
      </Routes>
    </AuthProvider>
  )
}
