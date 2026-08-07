// Wordmark Tuléh — tile monogram + teks. Selalu di atas latar pine
// (sidebar & login), jadi teksnya selalu putih. `flush` menghapus padding
// bawaan saat pembungkusnya sudah punya padding sendiri (login).

export default function Brand({ flush = false }: { flush?: boolean }) {
  return (
    <div className="sider-brand" style={flush ? { padding: 0 } : undefined}>
      <span className="brand-mark">T</span>
      <span className="brand-word" style={{ color: '#fff' }}>
        Tuléh<em>POS</em>
      </span>
    </div>
  );
}
