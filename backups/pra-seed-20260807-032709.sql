--
-- PostgreSQL database dump
--

\restrict Y6MurkJocTOsMP6HLUOgfxmzrpxDrCtgLc8pyAn3cKCLJNn5Ct30UPnsaRcaUaS

-- Dumped from database version 16.14
-- Dumped by pg_dump version 16.14

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: holds; Type: TABLE; Schema: public; Owner: tuleh
--

CREATE TABLE public.holds (
    id bigint NOT NULL,
    label character varying(80),
    payload jsonb NOT NULL,
    user_id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.holds OWNER TO tuleh;

--
-- Name: holds_id_seq; Type: SEQUENCE; Schema: public; Owner: tuleh
--

CREATE SEQUENCE public.holds_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.holds_id_seq OWNER TO tuleh;

--
-- Name: holds_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: tuleh
--

ALTER SEQUENCE public.holds_id_seq OWNED BY public.holds.id;


--
-- Name: kategoris; Type: TABLE; Schema: public; Owner: tuleh
--

CREATE TABLE public.kategoris (
    id bigint NOT NULL,
    nama character varying(100) NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.kategoris OWNER TO tuleh;

--
-- Name: kategoris_id_seq; Type: SEQUENCE; Schema: public; Owner: tuleh
--

CREATE SEQUENCE public.kategoris_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.kategoris_id_seq OWNER TO tuleh;

--
-- Name: kategoris_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: tuleh
--

ALTER SEQUENCE public.kategoris_id_seq OWNED BY public.kategoris.id;


--
-- Name: pelanggans; Type: TABLE; Schema: public; Owner: tuleh
--

CREATE TABLE public.pelanggans (
    id bigint NOT NULL,
    nama character varying(150) NOT NULL,
    telepon character varying(20),
    email character varying(150),
    catatan character varying(255),
    aktif boolean NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.pelanggans OWNER TO tuleh;

--
-- Name: pelanggans_id_seq; Type: SEQUENCE; Schema: public; Owner: tuleh
--

CREATE SEQUENCE public.pelanggans_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.pelanggans_id_seq OWNER TO tuleh;

--
-- Name: pelanggans_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: tuleh
--

ALTER SEQUENCE public.pelanggans_id_seq OWNED BY public.pelanggans.id;


--
-- Name: pengeluarans; Type: TABLE; Schema: public; Owner: tuleh
--

CREATE TABLE public.pengeluarans (
    id bigint NOT NULL,
    tanggal date NOT NULL,
    keterangan character varying(255) NOT NULL,
    nominal numeric(15,2) NOT NULL,
    user_id bigint,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.pengeluarans OWNER TO tuleh;

--
-- Name: pengeluarans_id_seq; Type: SEQUENCE; Schema: public; Owner: tuleh
--

CREATE SEQUENCE public.pengeluarans_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.pengeluarans_id_seq OWNER TO tuleh;

--
-- Name: pengeluarans_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: tuleh
--

ALTER SEQUENCE public.pengeluarans_id_seq OWNED BY public.pengeluarans.id;


--
-- Name: produks; Type: TABLE; Schema: public; Owner: tuleh
--

CREATE TABLE public.produks (
    id bigint NOT NULL,
    kode character varying(30) NOT NULL,
    nama character varying(150) NOT NULL,
    barcode character varying(60),
    tipe character varying(10) DEFAULT 'BARANG'::character varying NOT NULL,
    satuan character varying(30) DEFAULT 'pcs'::character varying NOT NULL,
    harga_beli numeric(15,2) DEFAULT 0 NOT NULL,
    harga_jual numeric(15,2) DEFAULT 0 NOT NULL,
    harga_promo numeric(15,2),
    promo_mulai date,
    promo_selesai date,
    favorit boolean NOT NULL,
    kelola_stok boolean NOT NULL,
    stok numeric(15,3) DEFAULT 0 NOT NULL,
    kategori_id bigint,
    aktif boolean NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.produks OWNER TO tuleh;

--
-- Name: produks_id_seq; Type: SEQUENCE; Schema: public; Owner: tuleh
--

CREATE SEQUENCE public.produks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.produks_id_seq OWNER TO tuleh;

--
-- Name: produks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: tuleh
--

ALTER SEQUENCE public.produks_id_seq OWNED BY public.produks.id;


--
-- Name: sesi_kasirs; Type: TABLE; Schema: public; Owner: tuleh
--

CREATE TABLE public.sesi_kasirs (
    id bigint NOT NULL,
    nomor character varying(30) NOT NULL,
    user_id bigint NOT NULL,
    status character varying(10) NOT NULL,
    kas_awal numeric(15,2) NOT NULL,
    kas_akhir numeric,
    kas_sistem numeric,
    selisih numeric,
    catatan character varying(255),
    dibuka_pada timestamp with time zone,
    ditutup_pada timestamp with time zone,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.sesi_kasirs OWNER TO tuleh;

--
-- Name: sesi_kasirs_id_seq; Type: SEQUENCE; Schema: public; Owner: tuleh
--

CREATE SEQUENCE public.sesi_kasirs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.sesi_kasirs_id_seq OWNER TO tuleh;

--
-- Name: sesi_kasirs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: tuleh
--

ALTER SEQUENCE public.sesi_kasirs_id_seq OWNED BY public.sesi_kasirs.id;


--
-- Name: stok_logs; Type: TABLE; Schema: public; Owner: tuleh
--

CREATE TABLE public.stok_logs (
    id bigint NOT NULL,
    produk_id bigint NOT NULL,
    jenis character varying(10) NOT NULL,
    jumlah numeric(15,3) NOT NULL,
    stok_sesudah numeric(15,3) NOT NULL,
    keterangan character varying(255),
    sesi_kasir_id bigint,
    user_id bigint,
    created_at timestamp with time zone
);


ALTER TABLE public.stok_logs OWNER TO tuleh;

--
-- Name: stok_logs_id_seq; Type: SEQUENCE; Schema: public; Owner: tuleh
--

CREATE SEQUENCE public.stok_logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.stok_logs_id_seq OWNER TO tuleh;

--
-- Name: stok_logs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: tuleh
--

ALTER SEQUENCE public.stok_logs_id_seq OWNED BY public.stok_logs.id;


--
-- Name: transaksi_items; Type: TABLE; Schema: public; Owner: tuleh
--

CREATE TABLE public.transaksi_items (
    id bigint NOT NULL,
    transaksi_id bigint NOT NULL,
    produk_id bigint NOT NULL,
    nama_produk character varying(150) NOT NULL,
    satuan character varying(30) NOT NULL,
    harga numeric(15,2) NOT NULL,
    kuantitas numeric(15,3) NOT NULL,
    diskon_persen numeric(5,2) NOT NULL,
    pajak_persen numeric(5,2) NOT NULL,
    subtotal numeric(15,2) NOT NULL
);


ALTER TABLE public.transaksi_items OWNER TO tuleh;

--
-- Name: transaksi_items_id_seq; Type: SEQUENCE; Schema: public; Owner: tuleh
--

CREATE SEQUENCE public.transaksi_items_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.transaksi_items_id_seq OWNER TO tuleh;

--
-- Name: transaksi_items_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: tuleh
--

ALTER SEQUENCE public.transaksi_items_id_seq OWNED BY public.transaksi_items.id;


--
-- Name: transaksis; Type: TABLE; Schema: public; Owner: tuleh
--

CREATE TABLE public.transaksis (
    id bigint NOT NULL,
    nomor character varying(30) NOT NULL,
    sesi_kasir_id bigint NOT NULL,
    user_id bigint NOT NULL,
    idempotency_key character varying(80),
    tanggal timestamp with time zone,
    subtotal numeric(15,2) NOT NULL,
    total_diskon numeric(15,2) NOT NULL,
    diskon_persen numeric(5,2) NOT NULL,
    diskon_nominal numeric(15,2) NOT NULL,
    total_pajak numeric(15,2) NOT NULL,
    grand_total numeric(15,2) NOT NULL,
    dibayar numeric(15,2) NOT NULL,
    kembalian numeric(15,2) NOT NULL,
    tipe_pembayaran character varying(10) NOT NULL,
    status character varying(12) NOT NULL,
    catatan character varying(500),
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    pelanggan_id bigint
);


ALTER TABLE public.transaksis OWNER TO tuleh;

--
-- Name: transaksis_id_seq; Type: SEQUENCE; Schema: public; Owner: tuleh
--

CREATE SEQUENCE public.transaksis_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.transaksis_id_seq OWNER TO tuleh;

--
-- Name: transaksis_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: tuleh
--

ALTER SEQUENCE public.transaksis_id_seq OWNED BY public.transaksis.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: tuleh
--

CREATE TABLE public.users (
    id bigint NOT NULL,
    nama character varying(150) NOT NULL,
    email character varying(150) NOT NULL,
    password_hash character varying(100) NOT NULL,
    role character varying(20) DEFAULT 'KASIR'::character varying NOT NULL,
    aktif boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.users OWNER TO tuleh;

--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: tuleh
--

CREATE SEQUENCE public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.users_id_seq OWNER TO tuleh;

--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: tuleh
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: holds id; Type: DEFAULT; Schema: public; Owner: tuleh
--

ALTER TABLE ONLY public.holds ALTER COLUMN id SET DEFAULT nextval('public.holds_id_seq'::regclass);


--
-- Name: kategoris id; Type: DEFAULT; Schema: public; Owner: tuleh
--

ALTER TABLE ONLY public.kategoris ALTER COLUMN id SET DEFAULT nextval('public.kategoris_id_seq'::regclass);


--
-- Name: pelanggans id; Type: DEFAULT; Schema: public; Owner: tuleh
--

ALTER TABLE ONLY public.pelanggans ALTER COLUMN id SET DEFAULT nextval('public.pelanggans_id_seq'::regclass);


--
-- Name: pengeluarans id; Type: DEFAULT; Schema: public; Owner: tuleh
--

ALTER TABLE ONLY public.pengeluarans ALTER COLUMN id SET DEFAULT nextval('public.pengeluarans_id_seq'::regclass);


--
-- Name: produks id; Type: DEFAULT; Schema: public; Owner: tuleh
--

ALTER TABLE ONLY public.produks ALTER COLUMN id SET DEFAULT nextval('public.produks_id_seq'::regclass);


--
-- Name: sesi_kasirs id; Type: DEFAULT; Schema: public; Owner: tuleh
--

ALTER TABLE ONLY public.sesi_kasirs ALTER COLUMN id SET DEFAULT nextval('public.sesi_kasirs_id_seq'::regclass);


--
-- Name: stok_logs id; Type: DEFAULT; Schema: public; Owner: tuleh
--

ALTER TABLE ONLY public.stok_logs ALTER COLUMN id SET DEFAULT nextval('public.stok_logs_id_seq'::regclass);


--
-- Name: transaksi_items id; Type: DEFAULT; Schema: public; Owner: tuleh
--

ALTER TABLE ONLY public.transaksi_items ALTER COLUMN id SET DEFAULT nextval('public.transaksi_items_id_seq'::regclass);


--
-- Name: transaksis id; Type: DEFAULT; Schema: public; Owner: tuleh
--

ALTER TABLE ONLY public.transaksis ALTER COLUMN id SET DEFAULT nextval('public.transaksis_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: tuleh
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Data for Name: holds; Type: TABLE DATA; Schema: public; Owner: tuleh
--

COPY public.holds (id, label, payload, user_id, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: kategoris; Type: TABLE DATA; Schema: public; Owner: tuleh
--

COPY public.kategoris (id, nama, created_at, updated_at) FROM stdin;
1	Minuman	2026-08-06 12:34:25.93972+00	2026-08-06 12:34:25.93972+00
2	Layanan Laundry	2026-08-07 03:11:05.199079+00	2026-08-07 03:11:05.199079+00
3	Menu Kopi	2026-08-07 03:11:05.203291+00	2026-08-07 03:11:05.203291+00
4	Sembako	2026-08-07 03:11:05.227523+00	2026-08-07 03:11:05.227523+00
5	Makanan	2026-08-07 03:11:05.235831+00	2026-08-07 03:11:05.235831+00
\.


--
-- Data for Name: pelanggans; Type: TABLE DATA; Schema: public; Owner: tuleh
--

COPY public.pelanggans (id, nama, telepon, email, catatan, aktif, created_at, updated_at) FROM stdin;
1	Bu Sari	6281287870509	\N		t	2026-08-07 02:42:11.014828+00	2026-08-07 02:42:11.014828+00
2	Uji	\N	\N		t	2026-08-07 03:11:05.24421+00	2026-08-07 03:11:29.873953+00
3	Bu Rina	6281234500001	\N		t	2026-08-07 03:11:05.246063+00	2026-08-07 03:11:29.875989+00
4	Pak Joko	6281234500002	\N		t	2026-08-07 03:11:05.247201+00	2026-08-07 03:11:29.877561+00
5	Kak Sari	6281234500003	\N		t	2026-08-07 03:11:05.248739+00	2026-08-07 03:11:29.878855+00
6	Bang Ucok	6281234500004	\N		t	2026-08-07 03:11:05.24983+00	2026-08-07 03:11:29.88015+00
7	Ibu Fatimah	6281234500005	\N		t	2026-08-07 03:11:05.250955+00	2026-08-07 03:11:29.881145+00
8	Mas Andi	6281234500006	\N		t	2026-08-07 03:11:05.251974+00	2026-08-07 03:11:29.882173+00
9	Teh Lina	6281234500007	\N		t	2026-08-07 03:11:05.252832+00	2026-08-07 03:11:29.88314+00
10	Pak Haji Salim	6281234500008	\N		t	2026-08-07 03:11:05.253607+00	2026-08-07 03:11:29.884246+00
11	Pelanggan Umum	\N	\N		t	2026-08-07 03:11:05.254854+00	2026-08-07 03:11:29.885893+00
\.


--
-- Data for Name: pengeluarans; Type: TABLE DATA; Schema: public; Owner: tuleh
--

COPY public.pengeluarans (id, tanggal, keterangan, nominal, user_id, created_at, updated_at) FROM stdin;
1	2026-08-07	Bayar listrik toko	150000.00	1	2026-08-07 02:42:11.171386+00	2026-08-07 02:42:11.171386+00
\.


--
-- Data for Name: produks; Type: TABLE DATA; Schema: public; Owner: tuleh
--

COPY public.produks (id, kode, nama, barcode, tipe, satuan, harga_beli, harga_jual, harga_promo, promo_mulai, promo_selesai, favorit, kelola_stok, stok, kategori_id, aktif, created_at, updated_at) FROM stdin;
3	P-1786019760653	Gunting Rambut	\N	JASA	pcs	0.00	25000.00	\N	\N	\N	f	f	0.000	\N	t	2026-08-06 12:36:00.65931+00	2026-08-06 12:36:00.65931+00
2	P-1786019665994	Cuci Setrika Kiloan	\N	JASA	pcs	0.00	7000.00	\N	\N	\N	f	f	0.000	\N	t	2026-08-06 12:34:25.995618+00	2026-08-06 12:36:00.73051+00
20	DL-005	Cuci Sepatu	\N	JASA	Pcs	12000.00	25000.00	\N	\N	\N	f	f	0.000	2	t	2026-08-07 03:11:05.221847+00	2026-08-07 03:11:29.819477+00
21	DL-006	Karpet per Meter	\N	JASA	Kg	10000.00	20000.00	\N	\N	\N	f	f	0.000	2	t	2026-08-07 03:11:05.223818+00	2026-08-07 03:11:29.820536+00
22	DM-001	Aqua Botol 600ml	8998989100014	BARANG	Pcs	2500.00	4000.00	\N	\N	\N	f	t	47.000	1	t	2026-08-07 03:11:05.225351+00	2026-08-07 03:11:29.824403+00
4	P-1786019760766	Es Teh Manis	\N	BARANG	pcs	0.00	5000.00	\N	\N	\N	f	t	-5.000	1	t	2026-08-06 12:36:00.772343+00	2026-08-07 02:43:03.478896+00
23	DM-002	Teh Botol Sosro	8998989100021	BARANG	Pcs	3000.00	5000.00	\N	\N	\N	f	t	32.000	1	t	2026-08-07 03:11:05.226384+00	2026-08-07 03:11:29.827003+00
24	DM-003	Indomie Goreng	8998989100038	BARANG	Pcs	2700.00	3500.00	\N	\N	\N	f	t	105.000	4	t	2026-08-07 03:11:05.228562+00	2026-08-07 03:11:29.828767+00
25	DM-004	Beras Premium 5kg	8998989100045	BARANG	Pcs	62000.00	72000.00	\N	\N	\N	f	t	19.000	4	t	2026-08-07 03:11:05.229694+00	2026-08-07 03:11:29.830151+00
1	P-1786019665966	Kopi Susu Aren	\N	BARANG	cup	8000.00	15000.00	12000.00	2026-08-01	2026-08-31	t	t	40.000	1	t	2026-08-06 12:34:25.970035+00	2026-08-07 02:57:02.638939+00
5	26-P-00003	SusuUHT	\N	BARANG	pcs	3500.00	5000.00	\N	\N	\N	f	t	0.000	\N	t	2026-08-07 03:11:05.194948+00	2026-08-07 03:11:29.787945+00
6	26-P-00004	SusuTest2	\N	BARANG	pcs	0.00	6000.00	\N	\N	\N	f	t	0.000	2	t	2026-08-07 03:11:05.200612+00	2026-08-07 03:11:29.795458+00
7	26-P-00005	SusuTest3	\N	BARANG	pcs	0.00	6000.00	\N	\N	\N	f	t	0.000	2	t	2026-08-07 03:11:05.201992+00	2026-08-07 03:11:29.797009+00
8	DK-001	Kopi Susu Gula Aren	\N	BARANG	Gelas	8000.00	18000.00	\N	\N	\N	f	f	0.000	3	t	2026-08-07 03:11:05.205045+00	2026-08-07 03:11:29.799951+00
9	DK-002	Americano	\N	BARANG	Gelas	6000.00	15000.00	\N	\N	\N	f	f	0.000	3	t	2026-08-07 03:11:05.20622+00	2026-08-07 03:11:29.80134+00
10	DK-003	Cappuccino	\N	BARANG	Gelas	8500.00	20000.00	\N	\N	\N	f	f	0.000	3	t	2026-08-07 03:11:05.207142+00	2026-08-07 03:11:29.802729+00
11	DK-004	Es Teh Manis	\N	BARANG	Gelas	2500.00	8000.00	\N	\N	\N	f	f	0.000	3	t	2026-08-07 03:11:05.208126+00	2026-08-07 03:11:29.804045+00
12	DK-005	Croissant	\N	BARANG	Pcs	7000.00	16000.00	\N	\N	\N	f	f	0.000	3	t	2026-08-07 03:11:05.209205+00	2026-08-07 03:11:29.805403+00
13	DK-006	Roti Bakar Coklat	\N	BARANG	Porsi	6000.00	14000.00	\N	\N	\N	f	f	0.000	3	t	2026-08-07 03:11:05.210141+00	2026-08-07 03:11:29.806323+00
14	DK-007	Nasi Goreng Spesial	\N	BARANG	Porsi	12000.00	25000.00	\N	\N	\N	f	f	0.000	3	t	2026-08-07 03:11:05.210965+00	2026-08-07 03:11:29.807891+00
15	DK-008	Mie Goreng Telur	\N	BARANG	Porsi	9000.00	20000.00	\N	\N	\N	f	f	0.000	3	t	2026-08-07 03:11:05.21218+00	2026-08-07 03:11:29.809464+00
16	DL-001	Cuci Kering Kiloan	\N	JASA	Kg	4000.00	7000.00	\N	\N	\N	f	f	0.000	2	t	2026-08-07 03:11:05.213305+00	2026-08-07 03:11:29.813839+00
17	DL-002	Cuci Setrika Kiloan	\N	JASA	Kg	5000.00	9000.00	\N	\N	\N	f	f	0.000	2	t	2026-08-07 03:11:05.214161+00	2026-08-07 03:11:29.816606+00
18	DL-003	Setrika Saja	\N	JASA	Kg	3000.00	5000.00	\N	\N	\N	f	f	0.000	2	t	2026-08-07 03:11:05.215209+00	2026-08-07 03:11:29.817425+00
19	DL-004	Bed Cover	\N	JASA	Pcs	15000.00	30000.00	\N	\N	\N	f	f	0.000	2	t	2026-08-07 03:11:05.216304+00	2026-08-07 03:11:29.818543+00
26	DM-005	Gula Pasir 1kg	8998989100052	BARANG	Pcs	15500.00	18000.00	\N	\N	\N	f	t	29.000	4	t	2026-08-07 03:11:05.230606+00	2026-08-07 03:11:29.831227+00
27	DM-006	Minyak Goreng 1L	8998989100069	BARANG	Pcs	16500.00	19500.00	\N	\N	\N	f	t	23.000	4	t	2026-08-07 03:11:05.231512+00	2026-08-07 03:11:29.832574+00
28	DM-007	Kopi Sachet 10s	8998989100076	BARANG	Pcs	11000.00	14000.00	\N	\N	\N	f	t	38.000	1	t	2026-08-07 03:11:05.233039+00	2026-08-07 03:11:29.833598+00
29	DM-008	Sabun Mandi Batang	8998989100083	BARANG	Pcs	3200.00	4500.00	\N	\N	\N	f	t	47.000	4	t	2026-08-07 03:11:05.234299+00	2026-08-07 03:11:29.834748+00
30	DM-009	Chitato 68g	8998989100090	BARANG	Pcs	8500.00	11000.00	\N	\N	\N	f	t	28.000	5	t	2026-08-07 03:11:05.237508+00	2026-08-07 03:11:29.837198+00
31	DM-010	Roti Tawar	8998989100106	BARANG	Pcs	12000.00	15000.00	\N	\N	\N	f	t	14.000	5	t	2026-08-07 03:11:05.23888+00	2026-08-07 03:11:29.838543+00
\.


--
-- Data for Name: sesi_kasirs; Type: TABLE DATA; Schema: public; Owner: tuleh
--

COPY public.sesi_kasirs (id, nomor, user_id, status, kas_awal, kas_akhir, kas_sistem, selisih, catatan, dibuka_pada, ditutup_pada, created_at, updated_at) FROM stdin;
1	SK-20260807-00001	2	TUTUP	100000.00	119000	120000	-1000		2026-08-07 02:13:53.787974+00	2026-08-07 02:13:54.025166+00	2026-08-07 02:13:53.78843+00	2026-08-07 02:13:54.025956+00
2	SK-20260807-00002	2	TUTUP	50000.00	0	65000	-65000		2026-08-07 02:42:11.063694+00	2026-08-07 02:57:02.758185+00	2026-08-07 02:42:11.063982+00	2026-08-07 02:57:02.759053+00
\.


--
-- Data for Name: stok_logs; Type: TABLE DATA; Schema: public; Owner: tuleh
--

COPY public.stok_logs (id, produk_id, jenis, jumlah, stok_sesudah, keterangan, sesi_kasir_id, user_id, created_at) FROM stdin;
1	1	MASUK	50.000	48.000	Belanja mingguan	\N	1	2026-08-07 02:57:02.442482+00
2	1	OPNAME	-8.000	40.000	Hitung akhir hari	\N	1	2026-08-07 02:57:02.473848+00
3	1	JUAL	-2.000	38.000	POS-20260807-000005	2	2	2026-08-07 02:57:02.567942+00
4	1	BATAL	2.000	40.000	Batal POS-20260807-000005	2	2	2026-08-07 02:57:02.640191+00
\.


--
-- Data for Name: transaksi_items; Type: TABLE DATA; Schema: public; Owner: tuleh
--

COPY public.transaksi_items (id, transaksi_id, produk_id, nama_produk, satuan, harga, kuantitas, diskon_persen, pajak_persen, subtotal) FROM stdin;
1	1	1	Kopi Susu Aren	cup	12000.00	2.000	0.00	0.00	24000.00
2	2	4	Es Teh Manis	pcs	5000.00	1.000	0.00	0.00	5000.00
3	3	4	Es Teh Manis	pcs	5000.00	3.000	0.00	0.00	15000.00
4	4	4	Es Teh Manis	pcs	5000.00	1.000	0.00	0.00	5000.00
5	5	1	Kopi Susu Aren	cup	12000.00	2.000	0.00	0.00	24000.00
\.


--
-- Data for Name: transaksis; Type: TABLE DATA; Schema: public; Owner: tuleh
--

COPY public.transaksis (id, nomor, sesi_kasir_id, user_id, idempotency_key, tanggal, subtotal, total_diskon, diskon_persen, diskon_nominal, total_pajak, grand_total, dibayar, kembalian, tipe_pembayaran, status, catatan, created_at, updated_at, pelanggan_id) FROM stdin;
1	POS-20260807-000001	1	2	uji-kunci-1	2026-08-07 02:13:53.869663+00	24000.00	4000.00	0.00	4000.00	0.00	20000.00	50000.00	30000.00	TUNAI	SELESAI		2026-08-07 02:13:53.871068+00	2026-08-07 02:13:53.873231+00	\N
2	POS-20260807-000002	1	2	\N	2026-08-07 02:13:53.973872+00	5000.00	0.00	0.00	0.00	0.00	5000.00	5000.00	0.00	QRIS	SELESAI		2026-08-07 02:13:53.976523+00	2026-08-07 02:13:53.977436+00	\N
3	POS-20260807-000003	2	2	\N	2026-08-07 02:42:11.102471+00	15000.00	0.00	0.00	0.00	0.00	15000.00	15000.00	0.00	TUNAI	SELESAI		2026-08-07 02:42:11.103945+00	2026-08-07 02:42:11.107004+00	1
4	POS-20260807-000004	2	2	\N	2026-08-07 02:43:03.473599+00	5000.00	0.00	0.00	0.00	0.00	5000.00	5000.00	0.00	QRIS	SELESAI		2026-08-07 02:43:03.475719+00	2026-08-07 02:43:03.477835+00	1
5	POS-20260807-000005	2	2	\N	2026-08-07 02:57:02.557034+00	24000.00	0.00	0.00	0.00	0.00	24000.00	24000.00	0.00	TUNAI	DIBATALKAN		2026-08-07 02:57:02.559111+00	2026-08-07 02:57:02.639026+00	\N
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: tuleh
--

COPY public.users (id, nama, email, password_hash, role, aktif, created_at, updated_at) FROM stdin;
1	Admin Tuléh	admin@tuleh.id	$2a$10$UtrBBjT0oMg04Z2cgmzw4e8pI7/aAuF6wzKmIaf9fDY98cN3wfUZW	OWNER	t	2026-08-06 09:27:38.343734+00	2026-08-06 09:27:38.343734+00
2	Kasir Uji	kasir@tuleh.id	$2a$10$ZcIWSWW9YxMDn4KoJkuXyOd.aQwxYvJRpQHZUAXSRpx6sBFICXTBS	KASIR	t	2026-08-06 09:28:10.32029+00	2026-08-06 09:28:10.32029+00
\.


--
-- Name: holds_id_seq; Type: SEQUENCE SET; Schema: public; Owner: tuleh
--

SELECT pg_catalog.setval('public.holds_id_seq', 1, true);


--
-- Name: kategoris_id_seq; Type: SEQUENCE SET; Schema: public; Owner: tuleh
--

SELECT pg_catalog.setval('public.kategoris_id_seq', 5, true);


--
-- Name: pelanggans_id_seq; Type: SEQUENCE SET; Schema: public; Owner: tuleh
--

SELECT pg_catalog.setval('public.pelanggans_id_seq', 11, true);


--
-- Name: pengeluarans_id_seq; Type: SEQUENCE SET; Schema: public; Owner: tuleh
--

SELECT pg_catalog.setval('public.pengeluarans_id_seq', 1, true);


--
-- Name: produks_id_seq; Type: SEQUENCE SET; Schema: public; Owner: tuleh
--

SELECT pg_catalog.setval('public.produks_id_seq', 31, true);


--
-- Name: sesi_kasirs_id_seq; Type: SEQUENCE SET; Schema: public; Owner: tuleh
--

SELECT pg_catalog.setval('public.sesi_kasirs_id_seq', 2, true);


--
-- Name: stok_logs_id_seq; Type: SEQUENCE SET; Schema: public; Owner: tuleh
--

SELECT pg_catalog.setval('public.stok_logs_id_seq', 4, true);


--
-- Name: transaksi_items_id_seq; Type: SEQUENCE SET; Schema: public; Owner: tuleh
--

SELECT pg_catalog.setval('public.transaksi_items_id_seq', 5, true);


--
-- Name: transaksis_id_seq; Type: SEQUENCE SET; Schema: public; Owner: tuleh
--

SELECT pg_catalog.setval('public.transaksis_id_seq', 5, true);


--
-- Name: users_id_seq; Type: SEQUENCE SET; Schema: public; Owner: tuleh
--

SELECT pg_catalog.setval('public.users_id_seq', 2, true);


--
-- Name: holds holds_pkey; Type: CONSTRAINT; Schema: public; Owner: tuleh
--

ALTER TABLE ONLY public.holds
    ADD CONSTRAINT holds_pkey PRIMARY KEY (id);


--
-- Name: kategoris kategoris_pkey; Type: CONSTRAINT; Schema: public; Owner: tuleh
--

ALTER TABLE ONLY public.kategoris
    ADD CONSTRAINT kategoris_pkey PRIMARY KEY (id);


--
-- Name: pelanggans pelanggans_pkey; Type: CONSTRAINT; Schema: public; Owner: tuleh
--

ALTER TABLE ONLY public.pelanggans
    ADD CONSTRAINT pelanggans_pkey PRIMARY KEY (id);


--
-- Name: pengeluarans pengeluarans_pkey; Type: CONSTRAINT; Schema: public; Owner: tuleh
--

ALTER TABLE ONLY public.pengeluarans
    ADD CONSTRAINT pengeluarans_pkey PRIMARY KEY (id);


--
-- Name: produks produks_pkey; Type: CONSTRAINT; Schema: public; Owner: tuleh
--

ALTER TABLE ONLY public.produks
    ADD CONSTRAINT produks_pkey PRIMARY KEY (id);


--
-- Name: sesi_kasirs sesi_kasirs_pkey; Type: CONSTRAINT; Schema: public; Owner: tuleh
--

ALTER TABLE ONLY public.sesi_kasirs
    ADD CONSTRAINT sesi_kasirs_pkey PRIMARY KEY (id);


--
-- Name: stok_logs stok_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: tuleh
--

ALTER TABLE ONLY public.stok_logs
    ADD CONSTRAINT stok_logs_pkey PRIMARY KEY (id);


--
-- Name: transaksi_items transaksi_items_pkey; Type: CONSTRAINT; Schema: public; Owner: tuleh
--

ALTER TABLE ONLY public.transaksi_items
    ADD CONSTRAINT transaksi_items_pkey PRIMARY KEY (id);


--
-- Name: transaksis transaksis_pkey; Type: CONSTRAINT; Schema: public; Owner: tuleh
--

ALTER TABLE ONLY public.transaksis
    ADD CONSTRAINT transaksis_pkey PRIMARY KEY (id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: tuleh
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: idx_holds_user_id; Type: INDEX; Schema: public; Owner: tuleh
--

CREATE INDEX idx_holds_user_id ON public.holds USING btree (user_id);


--
-- Name: idx_kategoris_nama; Type: INDEX; Schema: public; Owner: tuleh
--

CREATE UNIQUE INDEX idx_kategoris_nama ON public.kategoris USING btree (nama);


--
-- Name: idx_pelanggans_nama; Type: INDEX; Schema: public; Owner: tuleh
--

CREATE INDEX idx_pelanggans_nama ON public.pelanggans USING btree (nama);


--
-- Name: idx_pelanggans_telepon; Type: INDEX; Schema: public; Owner: tuleh
--

CREATE INDEX idx_pelanggans_telepon ON public.pelanggans USING btree (telepon);


--
-- Name: idx_pengeluarans_tanggal; Type: INDEX; Schema: public; Owner: tuleh
--

CREATE INDEX idx_pengeluarans_tanggal ON public.pengeluarans USING btree (tanggal);


--
-- Name: idx_pengeluarans_user_id; Type: INDEX; Schema: public; Owner: tuleh
--

CREATE INDEX idx_pengeluarans_user_id ON public.pengeluarans USING btree (user_id);


--
-- Name: idx_produks_barcode; Type: INDEX; Schema: public; Owner: tuleh
--

CREATE INDEX idx_produks_barcode ON public.produks USING btree (barcode);


--
-- Name: idx_produks_kategori_id; Type: INDEX; Schema: public; Owner: tuleh
--

CREATE INDEX idx_produks_kategori_id ON public.produks USING btree (kategori_id);


--
-- Name: idx_produks_kode; Type: INDEX; Schema: public; Owner: tuleh
--

CREATE UNIQUE INDEX idx_produks_kode ON public.produks USING btree (kode);


--
-- Name: idx_produks_nama; Type: INDEX; Schema: public; Owner: tuleh
--

CREATE INDEX idx_produks_nama ON public.produks USING btree (nama);


--
-- Name: idx_sesi_kasirs_nomor; Type: INDEX; Schema: public; Owner: tuleh
--

CREATE UNIQUE INDEX idx_sesi_kasirs_nomor ON public.sesi_kasirs USING btree (nomor);


--
-- Name: idx_sesi_kasirs_status; Type: INDEX; Schema: public; Owner: tuleh
--

CREATE INDEX idx_sesi_kasirs_status ON public.sesi_kasirs USING btree (status);


--
-- Name: idx_sesi_kasirs_user_id; Type: INDEX; Schema: public; Owner: tuleh
--

CREATE INDEX idx_sesi_kasirs_user_id ON public.sesi_kasirs USING btree (user_id);


--
-- Name: idx_stok_logs_jenis; Type: INDEX; Schema: public; Owner: tuleh
--

CREATE INDEX idx_stok_logs_jenis ON public.stok_logs USING btree (jenis);


--
-- Name: idx_stok_logs_produk_id; Type: INDEX; Schema: public; Owner: tuleh
--

CREATE INDEX idx_stok_logs_produk_id ON public.stok_logs USING btree (produk_id);


--
-- Name: idx_stok_logs_sesi_kasir_id; Type: INDEX; Schema: public; Owner: tuleh
--

CREATE INDEX idx_stok_logs_sesi_kasir_id ON public.stok_logs USING btree (sesi_kasir_id);


--
-- Name: idx_stok_logs_user_id; Type: INDEX; Schema: public; Owner: tuleh
--

CREATE INDEX idx_stok_logs_user_id ON public.stok_logs USING btree (user_id);


--
-- Name: idx_transaksi_items_produk_id; Type: INDEX; Schema: public; Owner: tuleh
--

CREATE INDEX idx_transaksi_items_produk_id ON public.transaksi_items USING btree (produk_id);


--
-- Name: idx_transaksi_items_transaksi_id; Type: INDEX; Schema: public; Owner: tuleh
--

CREATE INDEX idx_transaksi_items_transaksi_id ON public.transaksi_items USING btree (transaksi_id);


--
-- Name: idx_transaksis_idempotency_key; Type: INDEX; Schema: public; Owner: tuleh
--

CREATE UNIQUE INDEX idx_transaksis_idempotency_key ON public.transaksis USING btree (idempotency_key);


--
-- Name: idx_transaksis_nomor; Type: INDEX; Schema: public; Owner: tuleh
--

CREATE UNIQUE INDEX idx_transaksis_nomor ON public.transaksis USING btree (nomor);


--
-- Name: idx_transaksis_pelanggan_id; Type: INDEX; Schema: public; Owner: tuleh
--

CREATE INDEX idx_transaksis_pelanggan_id ON public.transaksis USING btree (pelanggan_id);


--
-- Name: idx_transaksis_sesi_kasir_id; Type: INDEX; Schema: public; Owner: tuleh
--

CREATE INDEX idx_transaksis_sesi_kasir_id ON public.transaksis USING btree (sesi_kasir_id);


--
-- Name: idx_transaksis_status; Type: INDEX; Schema: public; Owner: tuleh
--

CREATE INDEX idx_transaksis_status ON public.transaksis USING btree (status);


--
-- Name: idx_transaksis_user_id; Type: INDEX; Schema: public; Owner: tuleh
--

CREATE INDEX idx_transaksis_user_id ON public.transaksis USING btree (user_id);


--
-- Name: idx_users_email; Type: INDEX; Schema: public; Owner: tuleh
--

CREATE UNIQUE INDEX idx_users_email ON public.users USING btree (email);


--
-- Name: holds fk_holds_user; Type: FK CONSTRAINT; Schema: public; Owner: tuleh
--

ALTER TABLE ONLY public.holds
    ADD CONSTRAINT fk_holds_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: pengeluarans fk_pengeluarans_user; Type: FK CONSTRAINT; Schema: public; Owner: tuleh
--

ALTER TABLE ONLY public.pengeluarans
    ADD CONSTRAINT fk_pengeluarans_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: produks fk_produks_kategori; Type: FK CONSTRAINT; Schema: public; Owner: tuleh
--

ALTER TABLE ONLY public.produks
    ADD CONSTRAINT fk_produks_kategori FOREIGN KEY (kategori_id) REFERENCES public.kategoris(id);


--
-- Name: sesi_kasirs fk_sesi_kasirs_user; Type: FK CONSTRAINT; Schema: public; Owner: tuleh
--

ALTER TABLE ONLY public.sesi_kasirs
    ADD CONSTRAINT fk_sesi_kasirs_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: stok_logs fk_stok_logs_produk; Type: FK CONSTRAINT; Schema: public; Owner: tuleh
--

ALTER TABLE ONLY public.stok_logs
    ADD CONSTRAINT fk_stok_logs_produk FOREIGN KEY (produk_id) REFERENCES public.produks(id);


--
-- Name: stok_logs fk_stok_logs_sesi_kasir; Type: FK CONSTRAINT; Schema: public; Owner: tuleh
--

ALTER TABLE ONLY public.stok_logs
    ADD CONSTRAINT fk_stok_logs_sesi_kasir FOREIGN KEY (sesi_kasir_id) REFERENCES public.sesi_kasirs(id);


--
-- Name: stok_logs fk_stok_logs_user; Type: FK CONSTRAINT; Schema: public; Owner: tuleh
--

ALTER TABLE ONLY public.stok_logs
    ADD CONSTRAINT fk_stok_logs_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: transaksi_items fk_transaksis_items; Type: FK CONSTRAINT; Schema: public; Owner: tuleh
--

ALTER TABLE ONLY public.transaksi_items
    ADD CONSTRAINT fk_transaksis_items FOREIGN KEY (transaksi_id) REFERENCES public.transaksis(id);


--
-- Name: transaksis fk_transaksis_pelanggan; Type: FK CONSTRAINT; Schema: public; Owner: tuleh
--

ALTER TABLE ONLY public.transaksis
    ADD CONSTRAINT fk_transaksis_pelanggan FOREIGN KEY (pelanggan_id) REFERENCES public.pelanggans(id);


--
-- Name: transaksis fk_transaksis_user; Type: FK CONSTRAINT; Schema: public; Owner: tuleh
--

ALTER TABLE ONLY public.transaksis
    ADD CONSTRAINT fk_transaksis_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- PostgreSQL database dump complete
--

\unrestrict Y6MurkJocTOsMP6HLUOgfxmzrpxDrCtgLc8pyAn3cKCLJNn5Ct30UPnsaRcaUaS

