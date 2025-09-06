# Manual Entry - Comprehensive Guide

## 🎯 **What is Manual Entry?**

**Manual Entry** adalah fitur yang memungkinkan user untuk menambahkan item ke sales invoice **tanpa harus memilih dari master product**. User dapat mengetik description dan harga secara manual.

## 🔄 **How Manual Entry Works**

### **1. Product Selection Logic**

#### **Dropdown Options:**
```
┌─────────────────────────┐
│ Select product          │ ← Placeholder
├─────────────────────────┤
│ Manual entry           │ ← value="" (Manual mode)
├─────────────────────────┤  
│ PRN-THR-0300 - Kertas  │ ← value="1" (Product ID)
│ BHG-05-19T - Logam     │ ← value="2" (Product ID)  
│ KEPTAL-002 - Kemas     │ ← value="3" (Product ID)
│ AQUA001 - Aqua         │ ← value="4" (Product ID)
└─────────────────────────┘
```

### **2. Behavior Matrix:**

| User Action | Product ID | Description | Unit Price | Behavior |
|-------------|------------|-------------|------------|----------|
| **Select "Manual entry"** | `0` | Empty (user input) | `0` (user input) | ✅ Manual Mode |
| **Select actual product** | `product.id` | Auto-filled | Auto-filled | ✅ Product Mode |
| **Switch back to Manual** | `0` | Cleared | Cleared | ✅ Reset to Manual |

### **3. Code Implementation:**

```typescript
// Product selection handler
const handleProductChange = (index: number, productId: number) => {
  const product = products.find(p => p.id === parseInt(productId.toString()));
  if (product) {
    // Product mode: Auto-fill from master data
    setValue(`items.${index}.product_id`, product.id);
    setValue(`items.${index}.description`, product.name);
    setValue(`items.${index}.unit_price`, product.price);
  } else {
    // Manual mode: Clear fields for manual input
    setValue(`items.${index}.product_id`, 0);
    setValue(`items.${index}.description`, '');
    setValue(`items.${index}.unit_price`, 0);
  }
};

// Helper to check if item is manual entry
const isManualEntry = (index: number) => {
  return !watchItems[index]?.product_id || watchItems[index]?.product_id === 0;
};
```

## 🎨 **Visual Indicators (Enhanced)**

### **1. Manual Entry Badge**
- **Orange "Manual" badge** muncul di bawah dropdown ketika manual entry selected
- Memberikan visual confirmation bahwa item ini manual entry

### **2. Field Styling**
```typescript
// Description field with manual entry styling
<Input
  placeholder={isManualEntry(index) 
    ? "Enter item description manually"  // Clear instruction
    : "Item description"                 // Default placeholder
  }
  bg={isManualEntry(index) ? 'orange.50' : inputBg}        // Light orange background
  borderColor={isManualEntry(index) ? 'orange.200' : 'gray.200'}  // Orange border
/>
```

### **3. Visual Feedback System**
- ✅ **Orange background** pada description field untuk manual items
- ✅ **Orange badge** "Manual" di product column
- ✅ **Dynamic placeholder** text yang lebih descriptive
- ✅ **Color coding** untuk easy identification

## 📋 **User Workflow**

### **Scenario 1: Creating Manual Entry Item**

```
Step 1: User click "Add Item" atau use existing row
Step 2: User click product dropdown
Step 3: User select "Manual entry" 
       ↓
       🎯 Visual changes happen:
       - Orange "Manual" badge appears
       - Description field gets orange background
       - Placeholder changes to "Enter item description manually"
       
Step 4: User type description: "Custom Service - Website Design"
Step 5: User set quantity: 1
Step 6: User input unit price: "Rp 2.500.000"
Step 7: User configure discount if needed: 10%
Step 8: User set taxable status: ON/OFF
Step 9: System calculates line total: Rp 2.250.000 (after discount)
```

### **Scenario 2: Switching from Product to Manual**

```
Step 1: User initially selects "PRN-THR-0300 - Kertas Thermal"
       ↓
       Auto-filled:
       - Description: "Kertas Thermal" 
       - Unit Price: "Rp 10.000"
       
Step 2: User realizes need custom description
Step 3: User change dropdown back to "Manual entry"
       ↓  
       🎯 Fields get cleared:
       - Description: "" (empty for manual input)
       - Unit Price: "Rp 0" (user must input)
       - Orange styling applied
       
Step 4: User input custom description: "Kertas Thermal Roll 80mm x 80mm - Special Grade"
Step 5: User input custom price: "Rp 12.000"
```

### **Scenario 3: Mixed Items (Product + Manual)**

```
Row 1: Selected Product "PRN-THR-0300"
       - Description: Auto-filled
       - Price: Auto-filled
       - Visual: Normal styling
       
Row 2: Manual Entry
       - Description: "Installation Service"  
       - Price: "Rp 500.000"
       - Visual: Orange badge + orange styling
       
Row 3: Selected Product "AQUA001" 
       - Description: Auto-filled
       - Price: Auto-filled  
       - Visual: Normal styling
```

## 🔧 **Technical Details**

### **1. Data Structure**

#### **Manual Entry Item:**
```json
{
  "product_id": 0,                    // Always 0 for manual
  "description": "Custom Service",    // User input
  "quantity": 1,                      // User input
  "unit_price": 250000,               // User input (in currency)
  "discount_percent": 10,             // User input
  "taxable": true                     // User input
}
```

#### **Product-based Item:**
```json
{
  "product_id": 123,                  // Actual product ID
  "description": "Kertas Thermal",    // From master product
  "quantity": 5,                      // User input
  "unit_price": 10000,                // From master product (can be modified)
  "discount_percent": 5,              // User input
  "taxable": true                     // User input
}
```

### **2. Validation Rules**

#### **Manual Entry Validation:**
- ✅ **Description**: Required (tidak boleh kosong)
- ✅ **Unit Price**: Harus > 0 
- ✅ **Quantity**: Harus > 0
- ✅ **Discount**: 0-100%
- ✅ **Taxable**: Boolean

#### **Backend Processing:**
```typescript
// Backend akan menerima:
items: [
  {
    product_id: 0,           // Indicates manual entry
    description: "Custom Service",
    quantity: 1,
    unit_price: 250000,
    // ... other fields
  }
]
```

### **3. Use Cases**

#### **Perfect for:**
- ✅ **Custom services** (installation, consultation, design)
- ✅ **One-time items** yang tidak ada di master product
- ✅ **Special pricing** untuk specific customer
- ✅ **Bundling products** dengan description custom
- ✅ **Emergency items** yang belum diinput ke master

#### **Not Recommended for:**
- ❌ **Regular products** yang sudah ada di master
- ❌ **Inventory tracking** (manual entry tidak tracked)
- ❌ **Bulk recurring items** (better to add to master)

## 📊 **Calculation Impact**

### **Manual Entry items tetap mengikuti semua calculation rules:**

```
Manual Item: "Website Design Service"
- Quantity: 1
- Unit Price: Rp 2.000.000
- Discount: 10%
- Taxable: YES

Calculation:
- Line Total: Rp 1.800.000 (after 10% discount)
- Subject to Global Discount: YES
- Subject to PPN: YES (because taxable = true)
- Final contribution to total: Rp 1.800.000 + PPN
```

## 🚀 **Benefits**

### **Business Benefits:**
- ✅ **Flexibility** untuk handle custom items
- ✅ **Speed** dalam creating invoice tanpa setup master dulu
- ✅ **Professional** tetap maintain calculation accuracy
- ✅ **Scalability** mix manual dan product-based items

### **User Experience:**
- ✅ **Clear visual distinction** antara manual vs product items
- ✅ **Easy switching** between modes
- ✅ **Proper validation** mencegah errors
- ✅ **Intuitive workflow** dengan visual feedback

### **Technical Benefits:**
- ✅ **Consistent data structure** dengan product-based items
- ✅ **Same calculation logic** berlaku untuk semua items
- ✅ **Proper validation** di frontend dan backend
- ✅ **Audit trail** tetap terjaga dengan description

---

**Implementation Date:** September 5, 2025  
**Status:** ✅ Enhanced with visual indicators  
**Files Modified:** `SalesForm.tsx`  
**Impact:** Better UX for manual entry with clear visual feedback
