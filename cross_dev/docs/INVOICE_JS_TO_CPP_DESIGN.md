# Invoice: JS → C++ Data Design

## Overview

JavaScript (Vue app) sends invoice data to C++ via `CrossDev.invoke('createInvoice', payload)`.
C++ parses the JSON payload, maps it to `InvoiceData`, and calls `excel::createInvoice()`.

---

## 1. JS Payload Schema

```typescript
// What JS sends via CrossDev.invoke('createInvoice', payload)
interface CreateInvoicePayload {
  // Invoice info (required)
  ref: string;           // Invoice reference/number, e.g. "INV-2025-0042"
  date: string;          // Invoice date, e.g. "2025-02-14" or "February 14, 2025"

  // Client info
  client: {
    name: string;
    mobile: string;
    id?: string;         // e.g. tax ID, NIF; can be shown in Bill To block
    address?: string;    // optional full address
    email?: string;
  };

  // Purchased items
  items: Array<{
    description: string;
    qty: string | number;   // "2", 2, or "—" for services
    price: string;          // "28,500.00"
    amount?: string;        // optional; if omitted, C++ can compute qty × price
  }>;

  // Totals (required)
  subtotal: string;
  taxLabel?: string;     // e.g. "VAT (19%)"
  taxAmount?: string;
  total: string;
  currency?: string;     // "USD", "DZD", etc.

  // Company (issuer) — optional; if omitted, use app default
  company?: {
    name: string;
    address?: string;
    phone?: string;      // Mobile from banks.mobile
    email?: string;
    website?: string;
  };

  // Logo — optional; base64-encoded PNG from banks.logo_path
  logoBase64?: string;

  // Output
  outputPath?: string;   // Optional; if omitted, C++ generates unique path
  openAfterCreate?: boolean;  // Default true — open file in Excel after creation
}
```

---

## 2. Field Mapping (JS → C++ InvoiceData)

| JS Field | C++ Field |
|----------|-----------|
| `ref` | `invoiceNumber` |
| `date` | `invoiceDate` |
| `client.name` | `client.name` |
| `client.mobile` | `client.phone` |
| `client.id` | Append to `client.address` or as separate line |
| `client.address` | `client.address` |
| `client.email` | `client.email` |
| `items[].description` | `items[].description` |
| `items[].qty` | `items[].quantity` (stringified) |
| `items[].price` | `items[].unitPrice` |
| `items[].amount` or qty×price | `items[].amount` |
| `subtotal` | `subtotal` |
| `taxLabel` | `taxLabel` |
| `taxAmount` | `taxAmount` |
| `total` | `total` |
| `company` | `company` (or default) |

---

## 3. C++ Handler Flow

```
CrossDev.invoke('createInvoice', payload)
    → MessageRouter routes to CreateInvoiceHandler
    → Handler parses JSON
    → Maps payload to excel::InvoiceData
    → Resolves logo path (resolveLogoPath)
    → Generates output path (payload.outputPath or makeUniquePath)
    → excel::createInvoice(path, data, logoPath)
    → Optionally: open file, print
    → Returns { success: true, path: "..." } or { success: false, error: "..." }
```

---

## 4. JS Usage Example

```javascript
// In Vue component or composable
async function exportInvoice(sellBill) {
  if (!window.CrossDev?.invoke) {
    console.error('CrossDev not available');
    return;
  }

  const payload = {
    ref: sellBill.ref,
    date: formatDate(sellBill.date),
    client: {
      name: sellBill.clientName,
      mobile: sellBill.clientMobile,
      id: sellBill.clientId,
      address: sellBill.clientAddress,
      email: sellBill.clientEmail,
    },
    items: sellBill.cars.map(car => ({
      description: `${car.year} ${car.make} ${car.model} — ${car.exterior}`,
      qty: 1,
      price: formatMoney(car.price),
      amount: formatMoney(car.price),
    })),
    subtotal: formatMoney(sellBill.subtotal),
    taxLabel: 'VAT (19%)',
    taxAmount: formatMoney(sellBill.taxAmount),
    total: formatMoney(sellBill.total),
    currency: sellBill.currency || 'USD',
    openAfterCreate: true,
  };

  const result = await window.CrossDev.invoke('createInvoice', payload);
  if (result.success) {
    console.log('Invoice saved:', result.path);
  } else {
    console.error('Invoice failed:', result.error);
  }
}
```

---

## 5. Implementation Checklist

- [ ] Create `CreateInvoiceHandler` (handlers/create_invoice_handler.cpp)
- [ ] Add `createInvoiceHandler.h` and register in `app_runner.cpp`
- [ ] Add JSON→InvoiceData mapping with null checks
- [ ] Add default company (from options or hardcoded)
- [ ] Generate unique path when `outputPath` omitted
- [ ] Optional: open file after create (reuse openInDefaultApp logic)
- [ ] Add `useInvoiceExport` composable in src/composables/ for Vue

---

## 6. Optional Extensions

- **Company from options**: Store company info in options.json, merge with payload
- **Template path**: Allow `templatePath` in payload for custom templates
- **Print after open**: Add `printAfterOpen: true` to trigger Excel print (macOS)
