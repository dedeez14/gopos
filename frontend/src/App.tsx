// Peta rute + penjaga sesi. Rute privat dibungkus RutePrivat → AdminLayout.

import { Navigate, Route, Routes } from 'react-router-dom';
import type { ReactElement } from 'react';
import AdminLayout from './layout/AdminLayout';
import DashboardPage from './pages/DashboardPage';
import LoginPage from './pages/LoginPage';
import ProdukPage from './pages/ProdukPage';
import TransaksiPage from './pages/TransaksiPage';
import UsersPage from './pages/UsersPage';
import { auth } from './lib/auth';

function RutePrivat({ children }: { children: ReactElement }) {
  return auth.sudahMasuk() ? children : <Navigate to="/login" replace />;
}

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route
        element={
          <RutePrivat>
            <AdminLayout />
          </RutePrivat>
        }
      >
        <Route path="/" element={<DashboardPage />} />
        <Route path="/produk" element={<ProdukPage />} />
        <Route path="/transaksi" element={<TransaksiPage />} />
        <Route path="/users" element={<UsersPage />} />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
