# 🔧 Node.js Deprecation Warning Fix

## ⚠️ Problem
Frontend menampilkan deprecation warning:
```
(node:16583) [DEP0060] DeprecationWarning: The util._extend API is deprecated. Please use Object.assign() instead.
```

## 🔍 Root Cause
Warning ini disebabkan oleh dependency lama (kemungkinan dari Chakra UI atau React Icons) yang masih menggunakan `util._extend` API yang sudah deprecated di Node.js modern.

## 🚀 Solutions Available

### Option 1: Selective Suppression (Recommended)
Menggunakan custom script yang hanya suppress DEP0060 warning:
```bash
npm run dev
```

Script ini menggunakan `suppress-deprecation.js` yang hanya menyembunyikan warning DEP0060 tapi tetap menampilkan deprecation warning lainnya.

### Option 2: Complete Suppression
Suppress semua deprecation warnings:
```bash
npm run dev:clean
```

### Option 3: Show All Warnings (Debug Mode)
Untuk debugging atau development yang membutuhkan semua warning:
```bash
npm run dev:verbose
```

## 📋 Available Scripts

| Script | Description | Deprecation Warnings |
|--------|-------------|---------------------|
| `npm run dev` | Default development (selective suppress) | Only DEP0060 hidden |
| `npm run dev:clean` | Clean development (all suppress) | All hidden |
| `npm run dev:verbose` | Verbose development | All shown |
| `npm run build` | Production build (selective suppress) | Only DEP0060 hidden |
| `npm run start` | Production start (selective suppress) | Only DEP0060 hidden |

## 🛠️ Technical Details

### What is DEP0060?
- `util._extend` adalah private API yang akan dihapus di Node.js versi mendatang
- Penggantinya adalah `Object.assign()` yang sudah tersedia sejak ES6
- Warning ini tidak mempengaruhi functionality aplikasi

### Why Not Fix the Source?
- Warning berasal dari third-party dependencies
- Update dependencies bisa menyebabkan breaking changes
- Selective suppression adalah solusi sementara yang aman

### Future Considerations
- Monitor update dari Chakra UI dan React Icons
- Overwatch untuk dependency alternatives yang lebih modern
- Periodic check untuk compatibility dengan Node.js LTS terbaru

## 🔧 Troubleshooting

### If suppression doesn't work:
1. Clear node_modules dan reinstall:
   ```bash
   rm -rf node_modules package-lock.json
   npm install
   ```

2. Check Node.js version compatibility:
   ```bash
   node --version
   # Should be v18 or higher
   ```

### For production deployment:
Use the default build script which handles deprecation warnings:
```bash
npm run build
npm run start
```

## 📝 Notes
- Warning ini adalah **cosmetic issue** dan tidak mempengaruhi functionality
- Aplikasi tetap berjalan normal dengan atau tanpa warning
- Solution ini temporary sampai dependencies updated