# User Manual Penggunaan Aplikasi Sistem Akuntansi

Manual ini menjelaskan cara penggunaan setiap modul aplikasi secara non-teknis, dari penyiapan awal (termasuk modal awal di Kas & Bank) hingga menghasilkan laporan keuangan yang kredibel.

Catatan: Nama menu di aplikasi Anda mungkin sedikit berbeda. Sesuaikan istilah pada panduan ini dengan label yang tampil di layar Anda.

---

## 1) Tujuan Manual
- Membantu pengguna operasional (kasir, akunting, manajer) menjalankan aplikasi dari nol hingga menarik laporan keuangan.
- Menjaga konsistensi data sehingga laporan (Neraca, Laba-Rugi, Arus Kas) akurat dan dapat dipertanggungjawabkan.

---

## 2) Peran Pengguna & Hak Akses
- Admin Keuangan: Mengatur master (akun kas/bank, COA) dan seluruh transaksi, rekonsiliasi, serta akses semua laporan.
- Staf Kasir/Operasional: Mencatat transaksi harian (setoran, penarikan, transfer), melihat saldo.
- Manajer/Direktur: Melihat ringkasan kas, memeriksa rekonsiliasi, dan mengunduh laporan keuangan.

Prinsip umum: Berikan akses seperlunya sesuai tanggung jawab pengguna (least privilege).

---

## 3) Konsep Dasar (Bahasa Sehari-hari)
- Kas & Bank: Tempat mencatat uang tunai dan rekening bank perusahaan.
- Modal Awal: Saldo awal kas/bank saat mulai memakai sistem (diinput satu kali saat setup).
- Transaksi Harian: Setoran (penerimaan), Penarikan (pengeluaran), Transfer antar rekening.
- Rekonsiliasi Bank: Mencocokkan saldo sistem dengan rekening koran.
- Laporan Keuangan: Neraca, Laba-Rugi, Arus Kas, Buku Besar, Neraca Saldo — tersusun otomatis dari transaksi yang dicatat.

---

## 4) Peta Modul Aplikasi
- Dashboard: Ringkasan cepat kondisi bisnis dan notifikasi penting.
- Auth & Profil: Login, cek profil.
- Peran & Izin: Mengatur akses setiap pengguna.
- Kontak: Data pelanggan/pemasok.
- Produk & Inventory: Produk, kategori, satuan, stok, penyesuaian, opname.
- Akun (COA): Daftar akun akuntansi yang jadi kerangka laporan.
- Kas & Bank: Akun kas/bank, modal awal, transaksi, rekonsiliasi, ringkasan saldo.
- Penjualan & Pembayaran: Dokumen penjualan dan penerimaan pembayaran pelanggan.
- Pembelian & Pembayaran: Dokumen pembelian dan pembayaran ke pemasok.
- Gudang & Lokasi: Pengelolaan lokasi stok (jika multi-gudang).
- Aset Tetap: Pengelolaan aset dan depresiasi (bila digunakan).
- Jurnal & Drilldown: Penelusuran angka laporan hingga ke jurnal/transaksi sumber.
- Laporan Keuangan: Neraca, Laba-Rugi, Arus Kas, Buku Besar, Neraca Saldo.
- Notifikasi: Pemberitahuan kegiatan penting/approval.

---

## 5) Alur Kerja Inti (Ringkas)
1. Tambahkan akun Kas/Bank dan isi Modal Awal (sekali saat mulai).
2. Catat transaksi harian (Setoran, Penarikan, Transfer).
3. Lakukan rekonsiliasi bank berkala (mingguan/bulanan).
4. Tarik laporan keuangan untuk periode yang diinginkan.

---

## 6) Panduan Per Modul

### 6.1 Dashboard
Tujuan: Melihat ringkasan kas, penjualan/pembelian, stok kritis, dan notifikasi.
Cara pakai:
- Buka Dashboard setelah login.
- Perhatikan indikator: saldo kas & bank, penjualan bulan ini, stok menipis.
- Klik indikator untuk masuk ke modul terkait (mis. stok menipis → Produk).

### 6.2 Auth & Profil
Tujuan: Akses aman dan informasi akun.
Cara pakai:
- Login dengan akun Anda.
- Cek Profil untuk melihat peran dan izin Anda.
Tips keamanan:
- Jangan membagikan kredensial. Ganti password berkala sesuai kebijakan perusahaan.

### 6.3 Peran & Izin
Tujuan: Mengatur siapa bisa melakukan apa.
Cara pakai (Admin):
- Buka menu Pengguna/Izin.
- Atur peran (mis. finance, admin, kasir) dan izin per modul (lihat kebutuhan kerja).
Tips:
- Terapkan prinsip “least privilege”.

### 6.4 Kontak (Pelanggan & Pemasok)
Tujuan: Memudahkan pengisian transaksi dan analisis.
Cara pakai:
- Tambah kontak dengan tipe (Pelanggan/Pemasok), isi data dasar (alamat, kontak, NPWP bila ada).
- Gunakan pencarian saat membuat penjualan/pembelian untuk mempercepat input.

### 6.5 Produk & Inventory
Komponen: Kategori, Satuan, Produk, Penyesuaian Stok, Opname.
Cara pakai:
- Buat Kategori (Bahan Baku, Barang Jadi) dan Satuan (pcs, box, kg).
- Tambah Produk: nama, kode/SKU, kategori, satuan, harga dasar.
- Penyesuaian Stok: koreksi selisih stok (tulis alasan yang jelas).
- Opname: hitung fisik berkala, lalu konfirmasi di sistem.
Tips:
- Gunakan kode/SKU konsisten agar mudah telusur.

### 6.6 Akun (Chart of Accounts/COA)
Tujuan: Kerangka laporan keuangan.
Cara pakai:
- Lihat hierarki akun (Aset, Kewajiban, Ekuitas, Pendapatan, Beban).
- Tambah akun bila diperlukan, hindari duplikasi.
- Pastikan akun “Modal Disetor/Share Capital” tersedia untuk pengakuan modal awal.

### 6.7 Kas & Bank (inti dari modal awal hingga laporan)

#### A) Membuat Akun Kas/Bank + Modal Awal
Kapan: Saat mulai menggunakan sistem atau membuka rekening baru.
Langkah:
- Buka Kas & Bank → Tambah Akun.
- Isi: nama (mis. “Kas Utama”/“Bank Operasional”), tipe (Cash/Bank), mata uang.
- Isi “Saldo Awal” + tanggal saldo awal, dan keterangan (mis. “Modal pemilik”).
- Simpan.
Hasil:
- Akun kas/bank terbentuk dengan saldo awal.
- Sistem otomatis mengakui modal (ekuitas) sehingga Neraca akurat.
Catatan:
- Modal awal diinput sekali saat setup. Tambahan modal berikutnya gunakan transaksi Setoran (bukan mengubah saldo awal).

#### B) Transaksi Harian
1) Setoran (penerimaan kas/bank)
- Gunakan saat ada uang masuk: penjualan tunai, modal tambahan, pengembalian dana.
- Isi tanggal, akun penerima, nominal, keterangan singkat, dan nomor bukti.
- Simpan. Saldo bertambah, laporan akan terbarui.

2) Penarikan (pengeluaran kas/bank)
- Gunakan saat ada uang keluar: biaya operasional, pembelian tunai.
- Isi tanggal, akun sumber, nominal, tujuan/penggunaan dana, nomor bukti.
- Simpan. Saldo berkurang, laporan terbarui.

3) Transfer antar rekening
- Gunakan saat memindahkan dana dari satu akun kas/bank ke akun lainnya.
- Pilih rekening sumber & tujuan, isi tanggal dan nominal.
- Simpan. Saldo sumber turun dan saldo tujuan naik (tidak mengubah total kas perusahaan).

Tips transaksi:
- Gunakan tanggal kejadian sebenarnya.
- Isi catatan dan nomor referensi untuk memudahkan audit.

#### C) Rekonsiliasi Bank
Tujuan: Mencocokkan transaksi di sistem dengan mutasi rekening koran.
Langkah:
- Buka Kas & Bank → Rekonsiliasi.
- Pilih rekening bank dan tanggal cut-off (mis. akhir bulan).
- Masukkan saldo rekening koran pada tanggal tersebut.
- Centang transaksi yang sudah “cleared” di bank.
- Simpan. Selisih seharusnya minimal/nol.
Saran bila ada selisih:
- Periksa transaksi yang belum dicatat (biaya admin bank, bunga, pajak).
- Cek tanggal transaksi dan kemungkinan duplikasi.

#### D) Ringkasan Saldo
- Lihat total kas, total bank, dan saldo per akun untuk pemantauan harian.

### 6.8 Penjualan & Pembayaran
Alur umum: (Penawaran opsional) → Penjualan/Invoice → Terima Pembayaran.
Cara pakai:
- Buat Penjualan: pilih pelanggan, masukkan produk, kuantitas, harga, disertai pajak/diskon bila ada.
- Konfirmasi/simpan sesuai prosedur.
- Terima Pembayaran: pilih invoice, pilih akun penerima (Kas/Bank), isi jumlah & tanggal, simpan.
Dampak:
- Piutang turun, Kas/Bank naik, pendapatan tercatat. Laporan terbarui otomatis.
Tips:
- Uang muka/DP: catat sebagai penerimaan, lalu aplikasikan ke invoice saat terbit.

### 6.9 Pembelian & Pembayaran
Alur umum: PO → Terima Barang (bila ada) → Tagihan → Bayar Pemasok.
Cara pakai:
- Buat Pembelian: pilih pemasok, masukkan item, harga, dan pajak bila ada.
- Simpan/konfirmasi.
- Bayar Pemasok: pilih tagihan, pilih akun sumber (Kas/Bank), isi jumlah & tanggal, simpan.
Dampak:
- Hutang turun, Kas/Bank turun, persediaan/HPP terbarui sesuai alur yang diterapkan.

### 6.10 Gudang & Lokasi
Tujuan: Mengelola stok per lokasi (jika multi-gudang).
Cara pakai:
- Tambah lokasi gudang.
- Lakukan opname/transfer antar lokasi sesuai kebutuhan.
Manfaat:
- Stok akurat per lokasi, memudahkan operasi logistik.

### 6.11 Aset Tetap (opsional)
Tujuan: Mengelola aset dan depresiasi.
Cara pakai:
- Tambah aset (nilai perolehan, tanggal mulai, umur manfaat).
- Proses depresiasi berkala sesuai kebijakan perusahaan.
Manfaat:
- Beban depresiasi dan nilai buku aset tercatat rapi di laporan.

### 6.12 Jurnal & Drilldown
Tujuan: Menelusuri angka laporan ke jurnal/transaksi sumber.
Cara pakai:
- Buka menu Jurnal/Drilldown.
- Filter periode/akun/nomor jurnal.
- Buka detail untuk melihat baris debit/kredit dan referensinya.
Manfaat:
- Transparansi dan kemudahan audit.

### 6.13 Laporan Keuangan
Laporan utama:
- Neraca (Balance Sheet): Posisi aset, kewajiban, ekuitas pada tanggal tertentu.
- Laba-Rugi (Profit & Loss): Performa usaha pada periode.
- Arus Kas (Cash Flow): Aliran kas masuk/keluar.
- Buku Besar (General Ledger): Mutasi per akun.
- Neraca Saldo (Trial Balance): Ringkasan saldo seluruh akun.
Cara pakai:
- Pilih laporan → tentukan periode → tampilkan/unduh.
Tips interpretasi:
- Neraca harus seimbang (Aset = Kewajiban + Ekuitas).
- Analisis Laba-Rugi untuk melihat sumber pertumbuhan/penekanan biaya.

### 6.14 Notifikasi
Tujuan: Mengingatkan hal penting (approval, jatuh tempo, dsb.).
Cara pakai:
- Buka menu Notifikasi.
- Tindak lanjuti sesuai pesan (persetujuan, melengkapi data, proses pembayaran).

---

## 7) Checklist Operasional
Harian:
- Catat semua penerimaan/pengeluaran/transfer kas.
- Catat penjualan & pembayaran pelanggan, pembelian & pembayaran pemasok.
- Cek dashboard dan ringkasan saldo.

Mingguan:
- Opname stok (bila perlu) dan penyesuaian dengan alasan jelas.
- Rekonsiliasi bank sementara untuk transaksi besar.

Bulanan:
- Rekonsiliasi bank semua rekening hingga selisih nol.
- Tinjau akun pendapatan/beban yang melonjak tidak wajar.
- Tarik dan distribusikan laporan: Neraca, Laba-Rugi, Arus Kas, Buku Besar, Neraca Saldo.

---

## 8) Tips Kualitas Data (Agar Laporan Kredibel)
- Selalu gunakan fitur transaksi (hindari ubah saldo manual).
- Gunakan tanggal kejadian sebenarnya, isi referensi & catatan.
- Rekonsiliasi bank tepat waktu; catat biaya admin, bunga, pajak.
- Hindari akun/produk ganda; standarkan penamaan.
- Lakukan review berkala sebelum menutup periode.

---

## 9) Skenario Contoh dari Nol
- Hari 1:
  - Kas & Bank: Tambah “Kas Utama” dengan Modal Awal Rp5.000.000.
  - Produk: Tambah produk & kategori dasar.
  - Kontak: Tambah pelanggan & pemasok utama.
- Hari 2–30:
  - Pencatatan penjualan harian + penerimaan kas (Setoran).
  - Pencatatan pembelian + pembayaran pemasok (Penarikan).
  - Transfer dana dari Kas ke Bank bila diperlukan.
- Akhir Bulan:
  - Rekonsiliasi bank sampai selisih nol.
  - Tarik laporan utama untuk rapat manajemen.

---

## 10) FAQ Singkat
- Saldo bank di sistem tidak sama dengan rekening koran?
  - Lakukan rekonsiliasi; catat transaksi bank yang belum masuk sistem (admin bank, bunga, pajak), cek tanggal dan duplikasi.
- Tidak bisa hapus akun kas/bank?
  - Pastikan saldo akun nol dan tidak ada transaksi aktif yang menggantung.
- Laporan terlihat tidak wajar?
  - Pastikan periode benar, transaksi sudah lengkap, dan rekonsiliasi selesai.
- Ada transaksi salah nominal?
  - Buat pembetulan/penyesuaian sesuai prosedur internal, dengan catatan yang jelas.

---

## 11) Catatan Penutup
Ikuti urutan: Setup akun + modal awal → Transaksi harian → Rekonsiliasi → Laporan. 
Kedisiplinan pada langkah-langkah ini akan memastikan laporan keuangan yang akurat dan kredibel.
