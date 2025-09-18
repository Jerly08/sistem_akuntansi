# 🧪 Test Manual Asset Creation dari Receipt

## **Step-by-Step Testing:**

### **1. Buka Browser Console**
1. Buka Chrome/Edge Dev Tools (F12)
2. Pergi ke tab **Console**
3. Ready untuk melihat debug logs

### **2. Create Test Purchase**
1. Pergi ke **Purchases** page
2. Klik **"New Purchase"**
3. Isi form:
   ```
   Vendor: PT Epson Indonesia
   Date: Today
   Product: Mesin Printer (contoh)
   Quantity: 1
   Unit Price: Rp 3.885.000
   ```
4. **Save Purchase** → Status: DRAFT

### **3. Approval Process**
1. **Employee** submit for approval
2. **Finance** approve
3. **Status** berubah jadi: APPROVED

### **4. Create Receipt dengan Asset** ✨
1. Klik **"Create Receipt"** pada purchase yang APPROVED
2. **PENTING** - Di tabel Receipt Items:
   - ✅ **Centang "Create Asset"** checkbox
   - **Select Category**: Equipment
   - **Useful Life**: 3 years (untuk printer)
   - **Serial Number**: HP2024001 (optional)
3. Klik **🔍 Debug Info** button (jika muncul)
4. **Check Console** untuk debug logs
5. Klik **"Create Receipt"**

### **5. Verifikasi Hasil**
1. **Check Console Logs** - harus muncul:
   ```
   🔍 Debug - Receipt Items: [...] 
   🔍 Debug - Assets to Create: [...]
   🚀 Creating assets from receipt...
   📝 Starting asset creation process...
   🔧 Processing asset item: {...}
   ✅ Found purchase item: {...}
   📋 Asset data prepared: {...}
   🚀 Calling assetService.createAsset with data: {...}
   ✅ Asset created successfully via assetService: {...}
   📊 Asset creation summary: { created: 1, errors: 0 }
   ```

2. **Check Toast Notifications**:
   - Success toast: "Receipt Created Successfully! 🎉"
   - Asset toast: "Assets Created Successfully! 🎉"

3. **Navigate to Asset Master**:
   - Pergi ke **Assets** page
   - **Refresh** halaman
   - Asset baru harus muncul dengan nama: "Mesin Printer (PO/2025/09/001)"

### **6. Troubleshooting**

#### **Jika Asset Tidak Muncul:**
1. **Check Console Logs** untuk error messages
2. **Verifikasi checkbox** "Create Asset" dicentang
3. **Check Network Tab** di Dev Tools untuk API calls yang failed
4. **Manual refresh** Assets page

#### **Common Issues:**
- ❌ **Checkbox tidak dicentang** → No assets created
- ❌ **API error** → Check backend server
- ❌ **Permission error** → Check user token
- ❌ **Data validation error** → Check required fields

---

## **Expected Debug Flow:**

```
📋 User creates receipt with ✅ "Create Asset" checked
    ↓
🔍 Console logs: "Debug - Assets to Create: 1"
    ↓
🚀 Console logs: "Creating assets from receipt..."
    ↓
📝 Console logs: "Starting asset creation process..."
    ↓
✅ Console logs: "Asset created successfully"
    ↓
🎉 Toast: "Assets Created Successfully! 1 asset(s) created"
    ↓
🔄 Navigate to Assets page → Asset appears!
```

---

## **Success Indicators:**
- ✅ Console shows asset creation logs
- ✅ Success toast appears
- ✅ Asset appears in Asset Master
- ✅ Asset has correct data (name, price, vendor, etc.)

## **Next Steps If Working:**
- Test dengan multiple items
- Test dengan different categories
- Test error handling
