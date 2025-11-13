# Bugfix: Dropdown Account Tidak Bisa Di-scroll di Create Purchase

## Masalah
Saat membuat purchase baru dan memilih account di kolom "Account" pada Purchase Items, dropdown menampilkan beberapa account tapi **tidak bisa di-scroll** untuk melihat account yang ada di bawah. Padahal masih ada banyak account lainnya yang tidak terlihat.

## Screenshot Masalah
Dropdown terlihat mentok di "5202 - BEBAN LISTRIK", "5203 - BEBAN TELEPON", "5204 - BEBAN TRANSPORTASI..." (terpotong) dan tidak bisa di-scroll lagi.

## Root Cause
1. **maxHeight terlalu kecil**: Component `SearchableSelect` memiliki `maxHeight="200px"` yang terlalu kecil untuk menampilkan banyak account
2. **zIndex terlalu rendah**: zIndex 1000 bisa tertutup oleh elemen modal lainnya (modal biasanya punya zIndex 1400+)
3. **Tidak ada visual scrollbar**: User tidak tahu bahwa dropdown sebenarnya bisa di-scroll
4. **overflowX tidak diatur**: Bisa menyebabkan horizontal scroll yang tidak diinginkan
5. **Parent overflow constraints**: Dropdown menggunakan `position: absolute` sehingga dibatasi oleh parent container (Table > Td > Box maxW="240px")

## Solusi yang Diterapkan

### File: `frontend/src/components/common/SearchableSelect.tsx`

#### 1. Meningkatkan maxHeight dropdown
```typescript
// BEFORE:
maxHeight="200px"

// AFTER:
maxHeight="300px"  // Lebih besar untuk menampilkan lebih banyak opsi
```

#### 2. Meningkatkan zIndex
```typescript
// BEFORE:
zIndex={1000}

// AFTER:
zIndex={1500}  // Lebih tinggi dari modal (1400) untuk memastikan dropdown tampil di atas
```

#### 3. Menambahkan overflowX hidden
```typescript
// AFTER:
overflowX="hidden"  // Mencegah horizontal scroll yang tidak perlu
```

#### 4. Menambahkan Custom Scrollbar Styling
```typescript
sx={{ 
  scrollBehavior: 'smooth', 
  overscrollBehavior: 'contain',
  '&::-webkit-scrollbar': {
    width: '8px',  // Scrollbar lebih visible
  },
  '&::-webkit-scrollbar-track': {
    background: '#f1f1f1',
    borderRadius: '4px',
  },
  '&::-webkit-scrollbar-thumb': {
    background: '#888',
    borderRadius: '4px',
  },
  '&::-webkit-scrollbar-thumb:hover': {
    background: '#555',
  },
}}
```

#### 5. Menambahkan size prop support
```typescript
interface SearchableSelectProps {
  // ... existing props
  size?: 'sm' | 'md' | 'lg';  // NEW: Support untuk ukuran berbeda
}

// Default size = 'md'
// Bisa digunakan dengan: <SearchableSelect size="sm" ... />
```

#### 6. Menggunakan Portal dengan Fixed Positioning (SOLUSI UTAMA)
```typescript
// Import Portal
import { Portal } from '@chakra-ui/react';

// State untuk posisi dropdown
const [dropdownPosition, setDropdownPosition] = useState({ top: 0, left: 0, width: 0 });

// Update posisi saat dropdown dibuka
useEffect(() => {
  const updatePosition = () => {
    if (containerRef.current && isOpen) {
      const rect = containerRef.current.getBoundingClientRect();
      setDropdownPosition({
        top: rect.bottom + window.scrollY,
        left: rect.left + window.scrollX,
        width: rect.width,
      });
    }
  };
  // ... listeners
}, [isOpen]);

// Render dengan Portal dan fixed positioning
{isOpen && !isDisabled && (
  <Portal>
    <Box
      position="fixed"  // CHANGED: absolute → fixed
      top={`${dropdownPosition.top}px`}
      left={`${dropdownPosition.left}px`}
      width={`${dropdownPosition.width}px`}
      zIndex={2000}  // INCREASED: 1500 → 2000
      // ... other props
    >
      {/* dropdown content */}
    </Box>
  </Portal>
)}
```

**Mengapa Portal + Fixed Positioning?**
- Portal merender dropdown di luar hierarki DOM parent
- Fixed positioning membuat dropdown tidak dibatasi oleh parent overflow
- Dropdown bisa muncul di atas semua elemen tanpa batasan dari table, modal, atau container lainnya
- Position dihitung dinamis berdasarkan posisi input field

#### 7. Menghapus maxW constraint di parent
```typescript
// File: frontend/app/purchases/page.tsx

// BEFORE:
<Td minW="240px">
  <Box maxW="240px">  // ❌ Membatasi dropdown
    <SearchableSelect ... />
  </Box>
</Td>

// AFTER:
<Td minW="240px" position="relative">
  <Box minW="240px" w="100%">  // ✅ Tidak membatasi dropdown
    <SearchableSelect ... />
  </Box>
</Td>
```

## Testing
1. Buka halaman Purchases
2. Klik "Create New Purchase"
3. Tambah item dengan klik "Add Item"
4. Pilih product
5. Klik dropdown "Account"
6. ✅ Dropdown sekarang bisa di-scroll dengan mouse wheel
7. ✅ Scrollbar terlihat jelas di sisi kanan dropdown
8. ✅ Dropdown tidak terpotong dan bisa menampilkan semua account
9. ✅ Dropdown muncul di atas elemen modal lainnya

## Impact
- ✅ User sekarang bisa scroll dan memilih semua account yang tersedia
- ✅ Visual scrollbar membuat jelas bahwa ada lebih banyak opsi di bawah
- ✅ Smooth scrolling memberikan UX yang lebih baik
- ✅ Tidak ada perubahan breaking pada API atau data structure
- ✅ Fix ini juga berlaku untuk semua penggunaan `SearchableSelect` component di aplikasi

## Related Files
- `frontend/src/components/common/SearchableSelect.tsx` - Component yang diperbaiki
- `frontend/app/purchases/page.tsx` - Halaman yang menggunakan component

## Related Components
Component `SearchableSelect` digunakan di beberapa tempat:
- Purchase form (Account selection) ✅
- Sales form (jika ada)
- Other forms yang menggunakan searchable dropdown

## Notes
- Wheel scroll handler sudah ada sebelumnya dan berfungsi dengan baik
- Hanya perlu meningkatkan visibility (maxHeight) dan prioritas rendering (zIndex)
- Custom scrollbar styling hanya untuk Webkit browsers (Chrome, Edge, Safari)
- Firefox akan menggunakan native scrollbar

## Author
AI Assistant - Warp Agent Mode

## Date
2025-01-13
