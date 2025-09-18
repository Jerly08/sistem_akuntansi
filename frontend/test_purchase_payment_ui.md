# Frontend Testing Guide - Purchase Payment Integration

## 🧪 Manual Testing Checklist

### Prerequisites
- Backend server running on http://localhost:8080
- Frontend development server running
- User logged in with appropriate role (Admin/Finance/Director)
- At least one approved credit purchase with outstanding amount in the system

### Test Scenarios

#### 1. **Dashboard Statistics Display**
**Expected Results:**
- ✅ "Total Paid" card displays sum of paid amounts
- ✅ "Outstanding Amount" card shows total outstanding
- ✅ Cards have appropriate colors (green for paid, orange for outstanding)

**Test Steps:**
1. Navigate to `/purchases`
2. Observe the dashboard statistics cards
3. Verify the values match the table data

#### 2. **Enhanced Purchase Table**
**Expected Columns:**
- ✅ Purchase #
- ✅ Vendor  
- ✅ Date
- ✅ Total
- ✅ **Paid** (new - green if > 0)
- ✅ **Outstanding** (new - orange if > 0) 
- ✅ **Payment** (new - shows method + "Can Pay" badge)
- ✅ Status
- ✅ Approval Status
- ✅ Actions

**Test Steps:**
1. Check table headers include new payment columns
2. Verify paid amounts show in green when > 0
3. Verify outstanding amounts show in orange when > 0
4. Check payment method badges appear correctly
5. Look for "Can Pay" badge on eligible purchases

#### 3. **Record Payment Button Visibility**
**Should Show "Record Payment" Button:**
- ✅ Purchase status: APPROVED
- ✅ Payment method: CREDIT
- ✅ Outstanding amount: > 0
- ✅ User role: Admin, Finance, or Director

**Should NOT Show Button:**
- ❌ Purchase status: DRAFT, PENDING, REJECTED
- ❌ Payment method: CASH
- ❌ Outstanding amount: 0
- ❌ User role: Employee, Inventory Manager

**Test Steps:**
1. Log in as Admin/Finance/Director
2. Find approved credit purchase with outstanding amount
3. Verify "Record Payment" button appears
4. Log in as Employee
5. Verify button is hidden

#### 4. **Payment Modal Form**
**Form Fields:**
- ✅ Purchase information display (read-only)
- ✅ Payment amount (required, ≤ outstanding)
- ✅ Payment date (required, default today)
- ✅ Payment method (required, dropdown)
- ✅ Bank account (required for non-cash, dropdown)
- ✅ Reference (optional)
- ✅ Notes (optional, with default text)

**Test Steps:**
1. Click "Record Payment" on eligible purchase
2. Verify modal opens with purchase information
3. Check all form fields are present
4. Verify amount field is pre-filled with outstanding amount
5. Test validation by entering amount > outstanding
6. Test bank account dropdown loads correctly

#### 5. **Payment Form Validation**
**Validation Rules:**
- ✅ Amount > 0
- ✅ Amount ≤ outstanding amount
- ✅ Date is required
- ✅ Payment method is required
- ✅ Bank account required for non-cash methods

**Test Steps:**
1. Try to submit form with amount = 0
2. Try to submit with amount > outstanding
3. Try to submit without selecting payment method
4. Select "Bank Transfer" and try to submit without bank account
5. Verify error messages appear correctly

#### 6. **Payment Submission Process**
**Expected Flow:**
1. ✅ Form validation passes
2. ✅ Loading state shows during submission
3. ✅ Success notification appears
4. ✅ Modal closes automatically
5. ✅ Purchase list refreshes with updated amounts
6. ✅ Statistics cards update

**Test Steps:**
1. Fill out valid payment form
2. Click "Record Payment"
3. Verify loading spinner appears on button
4. Wait for success notification
5. Check modal closes
6. Verify purchase table shows updated paid/outstanding amounts
7. Check statistics cards reflect the new payment

#### 7. **Error Handling**
**Error Scenarios:**
- ✅ Network error during submission
- ✅ Backend validation errors
- ✅ Cash banks loading failure
- ✅ Purchase data loading failure

**Test Steps:**
1. Disconnect network and try to submit payment
2. Enter invalid data and submit
3. Verify error messages are user-friendly
4. Check form remains open on error

#### 8. **Integration with Payment Management**
**Expected Results:**
- ✅ Payment appears in Payment Management system
- ✅ Purchase amounts update immediately
- ✅ Payment code follows standard format
- ✅ Journal entries created automatically

**Test Steps:**
1. Record a payment through Purchase Management
2. Navigate to Payment Management system
3. Verify payment appears in payment list
4. Check payment details match submitted form
5. Verify purchase outstanding amount decreased

#### 9. **Role-Based Access Control**
**Admin Role:**
- ✅ Can see "Record Payment" button
- ✅ Can record payments
- ✅ Can see all payment information

**Finance Role:**
- ✅ Can see "Record Payment" button
- ✅ Can record payments
- ✅ Can see all payment information

**Director Role:**
- ✅ Can see "Record Payment" button
- ✅ Can record payments
- ✅ Can see all payment information

**Employee Role:**
- ❌ Cannot see "Record Payment" button
- ❌ Cannot record payments
- ✅ Can see payment information (read-only)

**Inventory Manager Role:**
- ❌ Cannot see "Record Payment" button
- ❌ Cannot record payments
- ✅ Can see payment information (read-only)

#### 10. **Responsive Design**
**Test Different Screen Sizes:**
- ✅ Desktop: All features work properly
- ✅ Tablet: Table scrolls horizontally if needed
- ✅ Mobile: Modal adapts to screen size

## 🐛 Common Issues to Check

### UI Issues
- [ ] Payment columns are too wide and break table layout
- [ ] Modal doesn't open on mobile devices
- [ ] Statistics cards don't update after payment
- [ ] Button spacing issues in action column

### Functional Issues
- [ ] "Record Payment" button shows for ineligible purchases
- [ ] Form validation allows invalid amounts
- [ ] Success notification doesn't appear
- [ ] Purchase list doesn't refresh after payment

### Data Issues  
- [ ] Outstanding amounts show incorrect values
- [ ] Payment methods display incorrectly
- [ ] Statistics calculations are wrong
- [ ] Bank accounts don't load in dropdown

### Integration Issues
- [ ] Payments don't appear in Payment Management
- [ ] Purchase amounts don't update after payment
- [ ] Error messages are not user-friendly
- [ ] Role-based access control not working

## ✅ Success Criteria

All tests pass when:
- [ ] Dashboard statistics display correctly
- [ ] Enhanced table shows payment information
- [ ] "Record Payment" button appears for eligible purchases only
- [ ] Payment modal form works with proper validation
- [ ] Payments are successfully recorded and integrated
- [ ] Purchase amounts update in real-time
- [ ] Role-based access control functions properly
- [ ] Error handling provides clear user feedback
- [ ] Integration with Payment Management is seamless

## 📝 Test Results Template

```
Date: _________________
Tester: _______________
Browser: ______________

Dashboard Statistics: ✅ / ❌
Enhanced Table: ✅ / ❌
Button Visibility: ✅ / ❌  
Payment Modal: ✅ / ❌
Form Validation: ✅ / ❌
Payment Submission: ✅ / ❌
Error Handling: ✅ / ❌
Payment Integration: ✅ / ❌
Access Control: ✅ / ❌
Responsive Design: ✅ / ❌

Issues Found:
_________________________________
_________________________________
_________________________________

Overall Status: ✅ PASS / ❌ FAIL
```