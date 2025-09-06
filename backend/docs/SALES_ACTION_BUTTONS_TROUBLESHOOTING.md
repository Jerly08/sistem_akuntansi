# Sales Action Buttons Troubleshooting Guide

## 🔍 **Problem Analysis**

User reports: "tombol action nya tidak berfungsi" di Sales Management table.

## 🛠️ **Implemented Fixes**

### ✅ **1. Fixed Layout Component Issues**

#### **Problem in `/sales/[id]/page.tsx`:**
```typescript
// ❌ BEFORE - Wrong component import
import UnifiedLayout from '@/components/layout/UnifiedLayout';
// But using wrong component name
<Layout allowedRoles={[...]}>  // Component tidak match
```

#### **Fixed:**
```typescript
// ✅ AFTER - Consistent component usage
import UnifiedLayout from '@/components/layout/UnifiedLayout';
<UnifiedLayout>  // Menggunakan component yang benar
```

### ✅ **2. Added Comprehensive Debug Logging**

#### **Enhanced Action Menu Debug:**
```typescript
// Debug logging untuk props
console.log('EnhancedSalesTable props:', {
  salesCount: sales?.length,
  canEdit,
  canDelete,
  hasEditHandler: !!onEdit,
  hasConfirmHandler: !!onConfirm,
  hasDeleteHandler: !!onDelete,
  hasDownloadHandler: !!onDownloadInvoice
});

// Debug logging untuk setiap action click
onClick={() => {
  console.log('View Details clicked for sale:', sale.id);
  onViewDetails(sale);
}}
```

### ✅ **3. Enhanced Action Button UX**

#### **Improved Menu Button:**
```typescript
<MenuButton
  as={IconButton}
  icon={<FiMoreVertical />}
  variant="ghost"
  size="sm"
  aria-label="Actions for sale"
  _hover={{ bg: hoverBg }}
  data-testid={`actions-${sale.id}`}  // Better testability
/>
```

#### **Development Debug Info:**
```typescript
{process.env.NODE_ENV === 'development' && (
  <MenuItem isDisabled fontSize="xs" color="gray.500">
    Debug: canEdit={String(canEdit)}, canDelete={String(canDelete)}
  </MenuItem>
)}
```

#### **Action Fallback:**
```typescript
{!canEdit && !canDelete && !onDownloadInvoice && (
  <>
    <MenuDivider />
    <MenuItem isDisabled fontSize="xs" color="gray.500">
      No actions available
    </MenuItem>
  </>
)}
```

## 🧪 **Debug Steps to Test**

### **Step 1: Check Console Logs**
Open browser console dan lihat output:
```
EnhancedSalesTable props: {
  salesCount: 1,
  canEdit: true,
  canDelete: true,
  hasEditHandler: true,
  hasConfirmHandler: true,
  hasDeleteHandler: true,
  hasDownloadHandler: true
}
```

### **Step 2: Test Each Action**
1. **View Details** - Harus selalu muncul
2. **Edit** - Hanya muncul jika `canEdit=true` AND `status='DRAFT'`
3. **Confirm & Invoice** - Hanya muncul jika `canEdit=true` AND `status='DRAFT'`
4. **Record Payment** - Hanya muncul jika `canEdit=true` AND `outstanding_amount > 0`
5. **Cancel Sale** - Hanya muncul jika `canEdit=true` AND status bukan 'PAID'/'CANCELLED'
6. **Download Invoice** - Harus selalu muncul jika handler ada
7. **Delete** - Hanya muncul jika `canDelete=true`

### **Step 3: Check Permissions**
```typescript
// Di browser console, check:
console.log('User permissions:', {
  canCreate: canCreate,
  canEdit: canEdit, 
  canDelete: canDelete,
  canExport: canExport
});
```

## 🔧 **Common Issues & Solutions**

### **Issue 1: Actions Not Showing**
**Symptoms:** Menu kosong atau tidak ada menu items
**Cause:** Permission `canEdit`/`canDelete` false
**Solution:** Check `useModulePermissions('sales')` result

### **Issue 2: Actions Not Clickable**  
**Symptoms:** Menu items ada tapi tidak respond saat diklik
**Cause:** Handler functions tidak terdefinisi
**Solution:** Check console logs untuk error messages

### **Issue 3: View Details Opens Blank Page**
**Symptoms:** `/sales/[id]` page tidak load dengan benar
**Cause:** Layout component issues (sudah fixed)
**Solution:** Pastikan menggunakan `UnifiedLayout` yang benar

### **Issue 4: Permissions Always False**
**Symptoms:** `canEdit` dan `canDelete` selalu false
**Cause:** User role tidak memiliki akses sales module
**Solution:** 
```typescript
// Check user role di console
console.log('Current user role:', user?.role);
console.log('Module permissions:', useModulePermissions('sales'));
```

## 📋 **Action Menu Structure**

### **Complete Action Menu Flow:**
```
┌─────────────────────────┐
│ ⋮ Actions               │
├─────────────────────────┤
│ 👁️ View Details        │  ← Always available
├─────────────────────────┤
│ Debug: canEdit=true     │  ← Development only
├─────────────────────────┤
│ ✏️ Edit                 │  ← canEdit && DRAFT
│ ✅ Confirm & Invoice    │  ← canEdit && DRAFT  
│ 💰 Record Payment      │  ← canEdit && outstanding > 0
│ ❌ Cancel Sale         │  ← canEdit && not PAID/CANCELLED
├─────────────────────────┤
│ 📥 Download Invoice    │  ← Always if handler exists
├─────────────────────────┤
│ 🗑️ Delete              │  ← canDelete only
└─────────────────────────┘
```

## 🚀 **Testing Checklist**

### **For Users:**
- [ ] Action menu appears when clicking ⋮ button
- [ ] View Details opens sale detail page in new tab
- [ ] Edit opens sales form modal with existing data
- [ ] Confirm shows success message and updates status
- [ ] Download triggers PDF download
- [ ] Delete shows confirmation dialog

### **For Developers:**
- [ ] Console shows debug logs without errors
- [ ] Permissions correctly reflect user role
- [ ] All handlers are properly defined
- [ ] Error boundaries catch and display issues
- [ ] Layout components render without issues

## 🔄 **Next Steps If Issues Persist**

1. **Check Network Tab** - untuk API call errors
2. **Check User Permissions** - role dan module access
3. **Check Component Props** - pastikan handlers passed correctly
4. **Check Backend API** - endpoint `/sales/:id` responds correctly
5. **Check Route Configuration** - `/sales/[id]` route exists

---

**Implementation Date:** September 5, 2025  
**Status:** ✅ Enhanced with debugging  
**Files Modified:** `EnhancedSalesTable.tsx`, `sales/[id]/page.tsx`  
**Impact:** Better debugging and UX for action buttons
