import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { App as AntdApp, ConfigProvider, Empty } from 'antd';
import idID from 'antd/locale/id_ID';
import App from './App';
import { temaTuleh } from './theme';
import dayjs from 'dayjs';
import 'dayjs/locale/id';
import './index.css';

// Nama hari/bulan Bahasa Indonesia untuk seluruh format tanggal (grafik, dll).
dayjs.locale('id');

// Satu QueryClient untuk seluruh app; default konservatif supaya tidak
// membanjiri server (refetch saat fokus dimatikan, retry sekali).
const queryClient = new QueryClient({
  defaultOptions: {
    queries: { refetchOnWindowFocus: false, retry: 1, staleTime: 30_000 },
  },
});

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <QueryClientProvider client={queryClient}>
      <ConfigProvider
        locale={idID}
        theme={temaTuleh}
        renderEmpty={() => (
          <Empty
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            description="Belum ada data di sini."
            style={{ padding: '20px 0' }}
          />
        )}
      >
        {/* AntdApp menyediakan konteks message/modal ber-tema. */}
        <AntdApp>
          <BrowserRouter>
            <App />
          </BrowserRouter>
        </AntdApp>
      </ConfigProvider>
    </QueryClientProvider>
  </React.StrictMode>,
);
