# Quick Fix: Period Closing Error

## Problem
Period closing fails with error:
```
⚠ WARNING: Account 4101 (PENDAPATAN PENJUALAN) still has balance: 21600000.00
closing validation failed: 1 revenue/expense accounts still have non-zero balance
```

## Root Cause
Backend server is running **OLD CODE**. You need to restart with the latest code from GitHub.

---

## Solution (Choose One)

### Option 1: Automatic Script (Recommended)

**For Mac/Linux:**
```bash
cd backend
./restart_with_latest.sh
```

**For Windows:**
```powershell
cd backend
.\restart_with_latest.ps1
```

This script will automatically:
1. ✅ Stop the backend
2. ✅ Pull latest code
3. ✅ Fix database
4. ✅ Start backend with new code

---

### Option 2: Manual Steps

**1. Stop Backend Server**
- Press `Ctrl+C` in the terminal running backend
- Or close the terminal window

**2. Pull Latest Code**
```bash
cd /path/to/accounting_proj
git pull origin main
```

**3. Fix Database**
```bash
cd backend
go run cmd/verify_and_fix_pc.go
```

**4. Start Backend Again**
```bash
go run main.go
```

---

## Verification

After restart, check:

1. **Backend log should show**: Starting on port 8080
2. **Try period closing again** from UI
3. **Should succeed** without errors

### If Still Fails

Run debug script to see what's wrong:
```bash
cd backend
go run cmd/debug_closing_state.go
```

Then report the output.

---

## Important Notes

⚠️ **Always pull latest code before fixing**
⚠️ **Must restart backend after pulling code**
⚠️ **Database fix script is safe - it only fixes balances**

---

## Support

If you still get errors after following these steps:
1. Run: `go run cmd/debug_closing_state.go`
2. Share the output
3. Check if backend version matches:
   ```bash
   git log -1 --oneline
   ```
   Should show: "Fix revenue balance calculation - use absolute value for closing"
