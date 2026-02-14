# Excel Export — Reusability & Maintainability Design

## Goals

- **Reusable**: One core fills different templates (sell bill, car list, reports, etc.)
- **Maintainable**: Change template layout or add new exports without touching core logic
- **Testable**: Unit-test row operations; integration-test with example templates

---

## 1. Layered Architecture

```
┌─────────────────────────────────────────────────────────────┐
│  Export Handlers (sell_bill, car_list, …)                   │  ← Domain-specific
│  - Fetch data, map to row values, call ExcelExporter         │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  ExcelExporter (or TemplateExporter)                        │  ← Reusable core
│  - Open template, apply layout, fill, save                   │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  ExcelSheetOperations                                       │  ← Low-level helpers
│  - shiftRowsDown, copyRowFormat, fillDetailRows              │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  OpenXLSX                                                    │
└─────────────────────────────────────────────────────────────┘
```

---

## 2. Layout Configuration (data-driven)

Define template layout in a struct instead of magic numbers:

```cpp
// include/excel/excel_layout.h
struct ExcelDetailBlock {
    uint32_t headerRow;        // e.g. 6
    uint32_t detailStartRow;   // e.g. 7 (has formatting)
    uint32_t detailNumCols;    // e.g. 7 (A..G)
    uint32_t shiftNumCols;     // e.g. 16 (A..P) for content below
};

// Preset layouts
namespace ExcelLayouts {
    inline ExcelDetailBlock sellBill() {
        return {6, 7, 7, 16};
    }
    // Future: carList(), inventoryReport(), etc.
}
```

**Benefit**: Change template row numbers in one place; all exports using that layout update.

---

## 3. Core Classes

### 3.1 `ExcelSheetOperations` (static helpers)

Pure functions for row operations — easy to unit-test.

```cpp
// include/excel/excel_sheet_ops.h
namespace excel {

void shiftRowsDown(OpenXLSX::XLWorksheet& wks,
                  uint32_t firstRow, uint32_t lastRow,
                  int rowCount, int numCols);

void copyRowFormatToRow(OpenXLSX::XLWorksheet& wks,
                        uint32_t srcRow, uint32_t destRow,
                        int numCols);

void fillDetailRows(OpenXLSX::XLWorksheet& wks,
                   const ExcelDetailBlock& layout,
                   const std::vector<std::vector<std::string>>& rows);

}
```

### 3.2 `ExcelExporter` (reusable facade)

```cpp
// include/excel/excel_exporter.h
class ExcelExporter {
public:
    explicit ExcelExporter(const std::string& templatePath);
    ~ExcelExporter();

    OpenXLSX::XLWorksheet worksheet(uint32_t index = 1);

    // Fill headers (single row)
    void setHeaderRow(uint32_t row, const std::vector<std::string>& values);

    // Fill detail block (handles shift + format copy + values)
    void fillDetailBlock(const ExcelDetailBlock& layout,
                        const std::vector<std::vector<std::string>>& detailRows);

    // Fill single cell (for totals, labels, etc.)
    void setCell(uint32_t row, uint16_t col, const std::string& value);

    bool saveAs(const std::string& path);

private:
    OpenXLSX::XLDocument doc_;
};
```

### 3.3 Domain-specific handler (sell bill example)

```cpp
// src/excel/sell_bill_exporter.cpp (or similar)
void exportSellBillToExcel(const SellBillData& bill, const std::string& outputPath) {
    ExcelExporter exporter("assets/template_sell_1.xlsx");
    auto layout = ExcelLayouts::sellBill();

    exporter.setHeaderRow(layout.headerRow, {"ColA", "ColB", /* ... */});

    std::vector<std::vector<std::string>> details;
    for (const auto& car : bill.cars) {
        details.push_back({car.name, car.price, /* ... */});
    }
    exporter.fillDetailBlock(layout, details);

    exporter.saveAs(outputPath);
}
```

---

## 4. File Structure (implemented)

```
cross_dev/
├── include/excel/
│   ├── excel_layout.h      # ExcelDetailBlock, ExcelLayouts::sellBill()
│   ├── excel_paths.h       # resolveTemplatePath()
│   ├── excel_sheet_ops.h   # shiftRowsDown, copyRowFormatToRow, fillDetailBlock
│   └── excel_exporter.h    # ExcelExporter class
├── src/excel/
│   ├── excel_sheet_ops.cpp
│   ├── excel_paths.cpp
│   └── excel_exporter.cpp
├── assets/
│   └── template_sell_1.xlsx
└── examples/
    └── excel_template_test.cpp     # uses ExcelExporter + layout
```

---

## 5. Template Discovery

Centralize template path resolution:

```cpp
// include/excel/excel_paths.h
namespace excel {
    std::string resolveTemplatePath(const std::string& filename);
    // Tries: ./assets/, ../assets/, etc.
}
```

---

## 6. Error Handling

- Use `std::optional` or return `bool` + `lastError()` for non-exceptions
- Or use exceptions for open/save and document them
- Log failures with clear messages (e.g. "Template not found: assets/template_sell_1.xlsx")

---

## 7. Extensibility Checklist

| Need | Approach |
|------|----------|
| New export type | Add `exportFooToExcel()` using `ExcelExporter` + new layout |
| New template layout | Add `ExcelDetailBlock` preset in `excel_layout.h` |
| Different sheet | `exporter.worksheet(2)` or add `worksheetByName()` |
| Multiple detail blocks | Extend `fillDetailBlock` or add `fillDetailBlockAt(block, startRow)` |
| Formulas | `exporter.setCellFormula(row, col, "=SUM(A7:A10)")` — add to `ExcelExporter` |
| Merge cells | Use OpenXLSX `mergeCells()` in sheet ops if needed |

---

## 8. Testing

- **Unit**: `ExcelSheetOperations` with in-memory doc (create minimal xlsx, call ops, assert cells)
- **Integration**: `excel_template_test` with real template — regression test
- **Manual**: Run export from app, open in Excel, verify layout

---

## 9. Implementation Order

1. Extract `ExcelSheetOperations` from `excel_template_test.cpp`
2. Add `ExcelDetailBlock` and `ExcelLayouts::sellBill()`
3. Implement `ExcelExporter` that uses them
4. Refactor `excel_template_test` to use `ExcelExporter`
5. Add `resolveTemplatePath()` and wire into CMake assets copy
6. When ready: add `SellBillExporter` and invoke from JS/Native bridge
