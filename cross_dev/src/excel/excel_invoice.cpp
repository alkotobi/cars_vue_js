/**
 * Reusable invoice generation: creates a styled Excel invoice from data.
 */
#include "excel/excel_invoice.h"
#include "excel/excel_image.h"
#include <OpenXLSX.hpp>
#include <stdexcept>

namespace excel {

namespace {

using namespace OpenXLSX;

// Layout constants (1-based rows) — tight spacing, logo 0.5x0.5
constexpr uint32_t ROW_COMPANY_NAME = 1;   // Same line as logo, right of logo
constexpr uint32_t ROW_COMPANY_DETAILS = 2;  // Address, phone, etc. below company name
constexpr uint32_t ROW_INVOICE_TITLE = 5;   // "INVOICE" on its own line
constexpr uint32_t ROW_BILL_TO = 6;         // Bill To and client data start here
constexpr uint32_t ROW_TABLE_HEADER = 11;   // Details table header
constexpr uint32_t COL_COMPANY = 2;         // Company name/details to the right of logo
constexpr uint32_t COL_NO = 1;
constexpr uint32_t COL_DESC = 2;
constexpr uint32_t COL_DESC_END = 3;        // Description spans B–C
constexpr uint32_t COL_CLIENT = 4;
constexpr uint32_t COL_PORT = 5;
constexpr uint32_t COL_QTY = 6;
constexpr uint32_t COL_UNIT = 7;
constexpr uint32_t COL_AMOUNT = 8;

void writeCompany(XLWorksheet& wks, OpenXLSX::XLStyles& styles,
                  const InvoiceCompany& co, XLStyleIndex companyFmt) {
    // Company name on same line as logo, to the right of logo
    wks.cell(ROW_COMPANY_NAME, COL_COMPANY).value() = co.name;
    wks.cell(ROW_COMPANY_NAME, COL_COMPANY).setCellFormat(companyFmt);
    uint32_t r = ROW_COMPANY_DETAILS;
    if (!co.address.empty()) { wks.cell(r, COL_COMPANY).value() = co.address; wks.cell(r, COL_COMPANY).setCellFormat(companyFmt); r++; }
    if (!co.phone.empty())  { wks.cell(r, COL_COMPANY).value() = co.phone;  wks.cell(r, COL_COMPANY).setCellFormat(companyFmt); r++; }
    if (!co.email.empty())  { wks.cell(r, COL_COMPANY).value() = co.email;  wks.cell(r, COL_COMPANY).setCellFormat(companyFmt); r++; }
    if (!co.website.empty()){ wks.cell(r, COL_COMPANY).value() = co.website; wks.cell(r, COL_COMPANY).setCellFormat(companyFmt); }
}

void writeInvoiceMeta(XLWorksheet& wks, const InvoiceData& data, XLStyleIndex metaFmt) {
    wks.cell(ROW_BILL_TO, 4).value() = "Invoice #: " + data.invoiceNumber;
    wks.cell(ROW_BILL_TO, 4).setCellFormat(metaFmt);
    wks.cell(ROW_BILL_TO + 1, 4).value() = "Date: " + data.invoiceDate;
    wks.cell(ROW_BILL_TO + 1, 4).setCellFormat(metaFmt);
}

void writeBillTo(XLWorksheet& wks, const InvoiceClient& client, XLStyleIndex clientFmt) {
    // Bill To and name on same line
    std::string billToLine = "Bill To: " + client.name;
    wks.cell(ROW_BILL_TO, 1).value() = billToLine;
    wks.cell(ROW_BILL_TO, 1).setCellFormat(clientFmt);
    uint32_t r = ROW_BILL_TO + 1;
    if (!client.address.empty()) { wks.cell(r, 1).value() = client.address; wks.cell(r, 1).setCellFormat(clientFmt); r++; }
    if (!client.phone.empty())  { wks.cell(r, 1).value() = client.phone;  wks.cell(r, 1).setCellFormat(clientFmt); r++; }
    if (!client.email.empty()) { wks.cell(r, 1).value() = client.email;  wks.cell(r, 1).setCellFormat(clientFmt); }
}

static std::string rangeStr(uint32_t row, uint16_t col1, uint16_t col2) {
    auto colLetter = [](uint16_t c) {
        std::string s;
        while (c > 0) { s = char('A' + (c - 1) % 26) + s; c = (c - 1) / 26; }
        return s;
    };
    return colLetter(col1) + std::to_string(row) + ":" + colLetter(col2) + std::to_string(row);
}

void writeLineItems(XLWorksheet& wks, OpenXLSX::XLStyles& styles,
                    const std::vector<InvoiceLineItem>& items,
                    XLStyleIndex headerFmt, XLStyleIndex rowFmt) {
    // Header: No | Description | Client | Port | Qty | Unit Price | Amount
    wks.cell(ROW_TABLE_HEADER, COL_NO).value() = "No";
    wks.cell(ROW_TABLE_HEADER, COL_DESC).value() = "Description";
    wks.cell(ROW_TABLE_HEADER, COL_CLIENT).value() = "Client";
    wks.cell(ROW_TABLE_HEADER, COL_PORT).value() = "Port";
    wks.cell(ROW_TABLE_HEADER, COL_QTY).value() = "Qty";
    wks.cell(ROW_TABLE_HEADER, COL_UNIT).value() = "Unit Price";
    wks.cell(ROW_TABLE_HEADER, COL_AMOUNT).value() = "Amount";
    wks.range(rangeStr(ROW_TABLE_HEADER, COL_NO, COL_AMOUNT)).setFormat(headerFmt);
    wks.mergeCells(rangeStr(ROW_TABLE_HEADER, COL_DESC, COL_DESC_END), true);

    for (size_t i = 0; i < items.size(); ++i) {
        uint32_t r = ROW_TABLE_HEADER + 1 + static_cast<uint32_t>(i);
        const auto& it = items[i];
        wks.cell(r, COL_NO).value() = static_cast<int>(i + 1);
        wks.cell(r, COL_DESC).value() = it.description;
        wks.cell(r, COL_CLIENT).value() = it.clientName;
        wks.cell(r, COL_PORT).value() = it.port;
        wks.cell(r, COL_QTY).value() = it.quantity;
        wks.cell(r, COL_UNIT).value() = it.unitPrice;
        wks.cell(r, COL_AMOUNT).value() = it.amount;
        wks.range(rangeStr(r, COL_NO, COL_AMOUNT)).setFormat(rowFmt);
        wks.mergeCells(rangeStr(r, COL_DESC, COL_DESC_END), true);
    }
}

void writeTotals(XLWorksheet& wks, const InvoiceData& data,
                 uint32_t startRow, XLStyleIndex labelFmt, XLStyleIndex totalFmt) {
    uint32_t r = startRow;
    wks.cell(r, COL_UNIT).value() = "Subtotal:";
    wks.cell(r, COL_UNIT).setCellFormat(labelFmt);
    wks.cell(r, COL_AMOUNT).value() = data.subtotal;
    wks.cell(r, COL_AMOUNT).setCellFormat(labelFmt);
    r++;
    if (!data.taxLabel.empty()) {
        wks.cell(r, COL_UNIT).value() = data.taxLabel + ":";
        wks.cell(r, COL_UNIT).setCellFormat(labelFmt);
        wks.cell(r, COL_AMOUNT).value() = data.taxAmount;
        wks.cell(r, COL_AMOUNT).setCellFormat(labelFmt);
        r++;
    }
    wks.cell(r, COL_UNIT).value() = "Total:";
    wks.cell(r, COL_UNIT).setCellFormat(totalFmt);
    wks.cell(r, COL_AMOUNT).value() = data.total;
    wks.cell(r, COL_AMOUNT).setCellFormat(totalFmt);
}

} // namespace

bool createInvoice(const std::string& path, const InvoiceData& data,
                   const std::string& logoPath) {
    try {
        OpenXLSX::XLDocument doc;
        doc.create(path, OpenXLSX::XLForceOverwrite);

        OpenXLSX::XLWorksheet wks = doc.workbook().worksheet(1);
        wks.setName("Invoice");

        auto& styles = doc.styles();
        auto& fonts = styles.fonts();
        auto& fills = styles.fills();
        auto& borders = styles.borders();
        auto& cellFormats = styles.cellFormats();

        // --- Create styles ---
        // Title: large bold dark blue
        OpenXLSX::XLStyleIndex fontTitleIdx = fonts.create();
        fonts[fontTitleIdx].setFontName("Arial");
        fonts[fontTitleIdx].setFontSize(18);
        fonts[fontTitleIdx].setBold();
        fonts[fontTitleIdx].setFontColor(OpenXLSX::XLColor("ff1a5276"));

        OpenXLSX::XLStyleIndex titleFmt = cellFormats.create();
        cellFormats[titleFmt].setFontIndex(fontTitleIdx);
        cellFormats[titleFmt].setApplyFont(true);

        // "INVOICE" on its own line (row 5)
        wks.cell(ROW_INVOICE_TITLE, 4).value() = "INVOICE";
        wks.cell(ROW_INVOICE_TITLE, 4).setCellFormat(titleFmt);

        // Company: bold 12pt
        OpenXLSX::XLStyleIndex fontCompanyIdx = fonts.create();
        fonts[fontCompanyIdx].setFontName("Arial");
        fonts[fontCompanyIdx].setFontSize(16);
        fonts[fontCompanyIdx].setBold();

        OpenXLSX::XLStyleIndex companyFmt = cellFormats.create();
        cellFormats[companyFmt].setFontIndex(fontCompanyIdx);
        cellFormats[companyFmt].setApplyFont(true);

        // Meta (invoice #, date): right-aligned
        OpenXLSX::XLStyleIndex fontMetaIdx = fonts.create();
        fonts[fontMetaIdx].setFontName("Arial");
        fonts[fontMetaIdx].setFontSize(10);

        OpenXLSX::XLStyleIndex metaFmt = cellFormats.create();
        cellFormats[metaFmt].setFontIndex(fontMetaIdx);
        cellFormats[metaFmt].setApplyFont(true);
        cellFormats[metaFmt].alignment().setHorizontal(OpenXLSX::XLAlignRight);

        // Bill To label
        OpenXLSX::XLStyleIndex clientFmt = cellFormats.create();
        cellFormats[clientFmt].setFontIndex(fontMetaIdx);
        cellFormats[clientFmt].setApplyFont(true);

        // Table header: gray fill, bold, border (9pt for printing)
        OpenXLSX::XLStyleIndex fontHeaderIdx = fonts.create();
        fonts[fontHeaderIdx].setFontName("Arial");
        fonts[fontHeaderIdx].setFontSize(9);
        fonts[fontHeaderIdx].setBold();
        fonts[fontHeaderIdx].setFontColor(OpenXLSX::XLColor("ffffffff"));

        OpenXLSX::XLStyleIndex fillHeaderIdx = fills.create();
        fills[fillHeaderIdx].setFillType(OpenXLSX::XLPatternFill);
        fills[fillHeaderIdx].setPatternType(OpenXLSX::XLPatternSolid);
        fills[fillHeaderIdx].setColor(OpenXLSX::XLColor("ff3498db"));

        OpenXLSX::XLStyleIndex borderHeaderIdx = borders.create();
        borders[borderHeaderIdx].setOutline(true);
        borders[borderHeaderIdx].setLeft(OpenXLSX::XLLineStyleThin, OpenXLSX::XLColor("ff2c3e50"));
        borders[borderHeaderIdx].setRight(OpenXLSX::XLLineStyleThin, OpenXLSX::XLColor("ff2c3e50"));
        borders[borderHeaderIdx].setTop(OpenXLSX::XLLineStyleThin, OpenXLSX::XLColor("ff2c3e50"));
        borders[borderHeaderIdx].setBottom(OpenXLSX::XLLineStyleThin, OpenXLSX::XLColor("ff2c3e50"));

        OpenXLSX::XLStyleIndex headerFmt = cellFormats.create();
        cellFormats[headerFmt].setFontIndex(fontHeaderIdx);
        cellFormats[headerFmt].setFillIndex(fillHeaderIdx);
        cellFormats[headerFmt].setBorderIndex(borderHeaderIdx);
        cellFormats[headerFmt].setApplyFont(true);
        cellFormats[headerFmt].setApplyFill(true);
        cellFormats[headerFmt].setApplyBorder(true);
        cellFormats[headerFmt].alignment().setHorizontal(OpenXLSX::XLAlignCenter);

        // Data rows: light border (9pt for printing)
        OpenXLSX::XLStyleIndex fontRowIdx = fonts.create();
        fonts[fontRowIdx].setFontName("Arial");
        fonts[fontRowIdx].setFontSize(9);

        OpenXLSX::XLStyleIndex borderRowIdx = borders.create();
        borders[borderRowIdx].setOutline(true);
        borders[borderRowIdx].setLeft(OpenXLSX::XLLineStyleHair, OpenXLSX::XLColor("ffbdc3c7"));
        borders[borderRowIdx].setRight(OpenXLSX::XLLineStyleHair, OpenXLSX::XLColor("ffbdc3c7"));
        borders[borderRowIdx].setTop(OpenXLSX::XLLineStyleHair, OpenXLSX::XLColor("ffbdc3c7"));
        borders[borderRowIdx].setBottom(OpenXLSX::XLLineStyleHair, OpenXLSX::XLColor("ffbdc3c7"));

        OpenXLSX::XLStyleIndex rowFmt = cellFormats.create();
        cellFormats[rowFmt].setFontIndex(fontRowIdx);
        cellFormats[rowFmt].setBorderIndex(borderRowIdx);
        cellFormats[rowFmt].setApplyFont(true);
        cellFormats[rowFmt].setApplyBorder(true);
        cellFormats[rowFmt].setApplyAlignment(true);
        cellFormats[rowFmt].alignment(OpenXLSX::XLCreateIfMissing).setWrapText(true);

        // Totals: bold (10pt, slightly larger than table)
        OpenXLSX::XLStyleIndex fontTotalIdx = fonts.create();
        fonts[fontTotalIdx].setFontName("Arial");
        fonts[fontTotalIdx].setFontSize(10);
        fonts[fontTotalIdx].setBold();

        OpenXLSX::XLStyleIndex totalFmt = cellFormats.create();
        cellFormats[totalFmt].setFontIndex(fontTotalIdx);
        cellFormats[totalFmt].setApplyFont(true);
        cellFormats[totalFmt].alignment().setHorizontal(OpenXLSX::XLAlignRight);

        OpenXLSX::XLStyleIndex labelFmt = cellFormats.create();
        cellFormats[labelFmt].setFontIndex(fontRowIdx);
        cellFormats[labelFmt].setApplyFont(true);
        cellFormats[labelFmt].alignment().setHorizontal(OpenXLSX::XLAlignRight);

        // --- Write content ---
        writeCompany(wks, styles, data.company, companyFmt);
        writeInvoiceMeta(wks, data, metaFmt);
        writeBillTo(wks, data.client, clientFmt);
        writeLineItems(wks, styles, data.items, headerFmt, rowFmt);

        uint32_t totalsStart = ROW_TABLE_HEADER + 2 + static_cast<uint32_t>(data.items.size());
        writeTotals(wks, data, totalsStart, labelFmt, totalFmt);

        // Column widths (compact but readable for A4 printing)
        wks.column(1).setWidth(6);    // No
        wks.column(2).setWidth(16);   // Description (spans B–C)
        wks.column(3).setWidth(10);
        wks.column(4).setWidth(12);   // Client
        wks.column(5).setWidth(10);   // Port
        wks.column(6).setWidth(5);    // Qty
        wks.column(7).setWidth(9);    // Unit Price
        wks.column(8).setWidth(10);   // Amount

        doc.saveAs(path, OpenXLSX::XLForceOverwrite);
        doc.close();

        if (!logoPath.empty()) {
            // Logo at top-left, stretched to 1" x 1"; company name at row 1 col B
            injectLetterhead(path, logoPath, 0.5, 0.5, true);  // 0.5" logo; includes pageSetup
        } else {
            injectPrintFitToPage(path);  // pageSetup only
        }
        return true;
    } catch (const std::exception&) {
        return false;
    }
}

} // namespace excel
