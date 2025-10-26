# Feature: COA Information Tooltip

## 📍 Lokasi
Halaman: `http://localhost:3000/accounts`

## 🎯 Tujuan
Memberikan panduan lengkap kepada user tentang Chart of Accounts (COA) sehingga user dapat:
- Memahami struktur kode akun
- Mengetahui akun-akun penting yang harus ada
- Tahu cara membuat ulang akun yang terhapus
- Mendapat tips best practices

## ✨ Fitur yang Ditambahkan

### 1. **Help Icon Button (❓)**
- **Lokasi**: Di sebelah heading "Chart of Accounts"
- **Fungsi**: Membuka popover informasi lengkap
- **Icon**: `FiHelpCircle` dari react-icons

### 2. **Comprehensive Popover Guide**
Popover berisi 5 section:

#### 🏷️ Struktur Kode Akun
```
1xxx - ASSETS (Aset)
2xxx - LIABILITIES (Kewajiban)
3xxx - EQUITY (Ekuitas/Modal)
4xxx - REVENUE (Pendapatan)
5xxx - EXPENSES (Beban)
```

#### ✅ Contoh Akun yang Harus Ada
Daftar akun-akun critical:
- `1101` - KAS
- `1102` - BANK
- `1201` - PIUTANG USAHA
- `2101` - UTANG USAHA
- `2103` - PPN KELUARAN
- `1240` - PPN MASUKAN
- `4101` - PENDAPATAN PENJUALAN
- `5101` - HARGA POKOK PENJUALAN

#### ⚠️ Tips Penting
- Jangan hapus akun yang sudah punya transaksi
- Header Account tidak bisa dihapus jika punya child
- Gunakan nama UPPERCASE untuk konsistensi
- Backup data sebelum hapus akun penting

#### 🔧 Jika Terhapus Tidak Sengaja
Panduan membuat ulang dengan:
- Kode yang sama
- Nama UPPERCASE
- Type sesuai kategori
- Parent account yang benar

#### 💡 Pro Tip
Penjelasan tentang perbedaan "Add Header Account" vs "Add Account"

### 3. **Button Tooltips**
Tooltips pada tombol aksi:

**Add Header Account:**
> "Buat kategori besar (Header) seperti ASSETS, CURRENT ASSETS, dll. Header tidak bisa digunakan untuk transaksi langsung."

**Add Account:**
> "Buat akun detail seperti KAS, BANK, PIUTANG USAHA yang bisa digunakan untuk mencatat transaksi."

### 4. **Info Banner**
Alert banner di bawah heading yang menampilkan:
- Akun-akun penting yang diperlukan
- Reminder untuk membuat ulang dengan kode dan nama yang sama
- Format: Alert dengan variant "left-accent" dan status "info"

## 🎨 Design Features

### Visual Elements:
- ✅ **Emoji Icons** untuk easy scanning
- ✅ **Color-coded sections** (blue, green, orange, purple)
- ✅ **Code tags** untuk highlight kode akun
- ✅ **Dividers** untuk separation yang jelas
- ✅ **Highlighted tip box** dengan background biru

### Responsiveness:
- Max width: 500px untuk readability
- Proper spacing dengan VStack
- Mobile-friendly popover placement

## 💻 Technical Implementation

### Components Used:
```tsx
- Popover (Chakra UI)
- PopoverTrigger
- PopoverContent
- PopoverHeader
- PopoverBody
- PopoverArrow
- PopoverCloseButton
- Tooltip
- Alert (with left-accent variant)
- Code tags
- UnorderedList
- IconButton (FiHelpCircle)
```

### Key Features:
1. **Non-intrusive**: Tidak mengganggu workflow normal
2. **On-demand**: Info hanya muncul saat user butuh
3. **Comprehensive**: Semua info penting ada di satu tempat
4. **Visual**: Menggunakan emoji dan color untuk kategorisasi

## 📱 User Experience Flow

1. User membuka halaman `/accounts`
2. Melihat info banner dengan ringkasan
3. Jika butuh detail lebih, klik icon help (❓)
4. Popover muncul dengan panduan lengkap
5. Hover tombol untuk tooltip context-specific
6. User dapat membuat akun dengan confidence

## 🔄 Future Enhancements (Optional)

### Potential Additions:
1. **Video tutorial link** dalam popover
2. **"Show me how" interactive guide**
3. **Recently deleted accounts list** untuk quick restore
4. **Account template import** untuk industry-specific COA
5. **Validation hints** saat membuat akun baru
6. **Quick create buttons** untuk common accounts

### Analytics Ideas:
- Track berapa kali help icon diklik
- Identifikasi section mana yang paling berguna
- A/B testing untuk format info yang optimal

## 🧪 Testing Checklist

- [x] Popover muncul saat klik help icon
- [x] Semua section tampil dengan benar
- [x] Emoji rendering correctly
- [x] Code tags styling proper
- [x] Colors sesuai design
- [x] Tooltips pada buttons work
- [x] Info banner visible
- [x] Responsive di mobile
- [x] Dark mode compatibility (inherit from theme)
- [x] Close button works
- [x] Click outside to close

## 📝 User Feedback Expected

### Positive Outcomes:
- ✅ Reduced support tickets tentang "missing accounts"
- ✅ Increased user confidence dalam manage COA
- ✅ Faster onboarding untuk new users
- ✅ Better understanding struktur akuntansi

### Metrics to Track:
- Number of help icon clicks
- Account creation success rate
- Time to create first account
- Error rate saat membuat accounts

## 🎓 Educational Value

Tooltip ini tidak hanya membantu user, tapi juga:
1. **Mengajarkan** struktur COA yang benar
2. **Mencegah** kesalahan umum
3. **Meningkatkan** financial literacy
4. **Mengurangi** dependency pada support team

## 🚀 Deployment Notes

### Requirements:
- Chakra UI components (already installed)
- react-icons (FiHelpCircle)
- No additional dependencies

### Performance:
- Lightweight (popover lazy loaded)
- No impact on page load
- Minimal bundle size increase

### Compatibility:
- Works dengan existing permission system
- Respects canCreate permission
- Theme-aware (dark mode ready)

---

## 📸 Screenshots (Expected UI)

### Desktop View:
```
┌─────────────────────────────────────────────────────┐
│ Chart of Accounts [?]  [Add Header] [Add Account]  │
│ ────────────────────────────────────────────────── │
│ 💡 Chart of Accounts Guidelines                    │
│ Akun-akun penting seperti KAS (1101), BANK...      │
└─────────────────────────────────────────────────────┘
```

### Popover Open:
```
┌──────────────────────────────┐
│ 📚 Panduan Chart of Accounts │
├──────────────────────────────┤
│ 🏷️ Struktur Kode Akun:       │
│ • 1xxx - ASSETS              │
│ • 2xxx - LIABILITIES         │
│ ...                          │
├──────────────────────────────┤
│ ✅ Contoh Akun yang Harus Ada│
│ • 1101 - KAS                 │
│ • 1102 - BANK                │
│ ...                          │
└──────────────────────────────┘
```

---

**Last Updated:** 2025-10-25
**Feature Status:** ✅ Implemented
**Version:** 1.0
