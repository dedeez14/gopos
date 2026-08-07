// Sistem desain Tuléh POS — satu sumber token untuk seluruh panel.
//
// Arah: "kehangatan retail Indonesia + kepercayaan finansial".
//   pine/jade  = uang, tepercaya (sengaja bukan biru korporat)
//   turmeric   = kehangatan rempah, dipakai HEMAT (promo, aktif, CTA kunci)
//   sage       = netral yang menenangkan untuk dipandang seharian
// Token ini menyebar ke SEMUA komponen AntD (tabel, input, tag, modal) —
// satu perubahan di sini merapikan setiap halaman sekaligus.

import type { ThemeConfig } from 'antd';

export const warna = {
  pine: '#0C3B2A', // sidebar & panel brand
  pine2: '#10513A',
  jade: '#147A54', // primary
  jadeHover: '#0F6446',
  jadeActive: '#0C5238',
  turmeric: '#E0952B', // aksen
  turmericSoft: '#F6C97D',
  ink: '#17241E',
  muted: '#5C6B62',
  line: '#E4EAE4',
  surface: '#FFFFFF',
  bg: '#EAF1EC', // sage
} as const;

export const temaTuleh: ThemeConfig = {
  token: {
    colorPrimary: warna.jade,
    colorInfo: warna.jade,
    colorLink: warna.jade,
    colorSuccess: '#1C7C4A',
    colorWarning: warna.turmeric,
    colorError: '#C0472F',
    colorTextBase: warna.ink,
    colorText: warna.ink,
    colorTextSecondary: warna.muted,
    colorBgLayout: warna.bg,
    colorBorder: warna.line,
    colorBorderSecondary: '#EDF2ED',
    borderRadius: 10,
    fontFamily: "'Plus Jakarta Sans', system-ui, -apple-system, sans-serif",
    fontSize: 14,
    controlHeight: 40,
    lineWidth: 1,
  },
  components: {
    Layout: {
      siderBg: warna.pine,
      headerBg: warna.surface,
      headerHeight: 64,
      headerPadding: '0 20px',
      bodyBg: 'transparent',
    },
    Menu: {
      darkItemBg: 'transparent',
      darkPopupBg: warna.pine2,
      darkItemColor: 'rgba(255,255,255,0.72)',
      darkItemHoverBg: 'rgba(255,255,255,0.07)',
      darkItemHoverColor: '#FFFFFF',
      darkItemSelectedBg: 'rgba(224,149,43,0.18)',
      darkItemSelectedColor: warna.turmericSoft,
      itemBorderRadius: 10,
      itemMarginInline: 10,
      itemHeight: 44,
      iconSize: 17,
    },
    Card: {
      borderRadiusLG: 16,
      paddingLG: 22,
      colorBorderSecondary: warna.line,
    },
    Table: {
      headerBg: '#F1F6F1',
      headerColor: '#3C4A42',
      headerSplitColor: 'transparent',
      borderColor: '#EDF2ED',
      rowHoverBg: '#F4F9F4',
      cellPaddingBlock: 14,
      headerBorderRadius: 12,
    },
    Button: {
      controlHeight: 40,
      fontWeight: 600,
      primaryShadow: '0 6px 16px -8px rgba(20,122,84,0.55)',
      borderRadius: 10,
    },
    Input: { controlHeight: 42, borderRadius: 10, activeShadow: '0 0 0 3px rgba(20,122,84,0.12)' },
    InputNumber: { controlHeight: 42, borderRadius: 10 },
    Select: { controlHeight: 42, borderRadius: 10 },
    DatePicker: { controlHeight: 42, borderRadius: 10 },
    Modal: { borderRadiusLG: 18, titleFontSize: 18 },
    Tag: { borderRadiusSM: 999 },
    Statistic: { titleFontSize: 13, contentFontSize: 30 },
    Segmented: { borderRadius: 10, controlHeight: 40 },
    Drawer: { colorBgElevated: warna.pine },
  },
};
