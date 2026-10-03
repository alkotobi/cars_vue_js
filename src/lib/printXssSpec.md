Print-document XSS: what is and is not covered by tests
========================================================

Every print view builds a full HTML document as a template literal and injects it
with `printWindow.document.write(...)`. That window is same-origin with the app, so
script running in it can read the API token from localStorage and call the API as
the signed-in user. Template literals interpolate verbatim, so any stored field
placed in one must be escaped at the point it lands.

What is fixed
-------------
All of the following now route stored content through `escapeHtml()` from
src/utils/escapeHtml.js:

  components/car-stock/CarStockTable.vue
  components/car-stock/ShowSelectionsModal.vue
  components/car-stock/CarStockPrintReport.vue   (a v-html sink)
  components/cashier/MoneyMovements.vue
  components/cashier/Expenses.vue
  components/cashier/UserTransactionsTable.vue
  components/cashier/TransfersInterTable.vue
  components/sells/SellBillsTable.vue
  lib/loadingPrintContent.js
  composables/useInvoiceCompanyInfo.js           (buildLetterheadHtml)

`rawHtml()` marks the fragments that are deliberately markup the app assembles
itself - pre-built `<td>` runs, the letterhead, the header/footer blocks. It is an
identity function and exists purely so those sites are greppable and reviewable:

    grep -rn 'rawHtml(' src/

What is covered by tests
------------------------
  utils/escapeHtml.spec.js          the escaper, including the stored-payload cases
  lib/loadingPrintContent.spec.js   behavioural: a payload in the car name, the
                                    note, the client name/mobile/NIN/ID, the VIN,
                                    the colour and the container name, plus the
                                    src-attribute break-out

That is deliberate and real coverage for `generatePrintContent`, which is a pure
function and therefore easy to drive with hostile input.

What is NOT covered, and why
---------------------------
The eight .vue builders above have no equivalent test. Two attempts were made and
both were dropped:

  1. "Scan every `${...}` in the file and require it to be wrapped."
     These builders nest template literals several deep - a `${cond ? `<td>
     ${escapeHtml(x)}</td>` : ''}` inside a .map() inside a report. A regex cannot
     tell where one interpolation ends, so it reported fragments of already-escaped
     calls as unescaped. Worse, the first version of it had no capture group, so it
     destructured `undefined` and passed with zero assertions - a green test that
     tested nothing.

  2. "Flag `${` + <known stored field> unless escapeHtml( appears before it."
     Fewer false positives, but still noisy: the common shape
     `${coreContent ? `<div>${escapeHtml(coreContent)}</div>` : ''}` puts the bare
     field name first and the wrapper after it, so every correctly-escaped site was
     reported.

A test that cannot parse its input cannot be trusted to catch anything, and one
that reports correctly-escaped code as broken trains people to ignore it. So
neither is kept.

What actually protects this
---------------------------
The review convention, which the `rawHtml()` marker exists to support: in these
files, an interpolation is either `escapeHtml(value)` for data or `rawHtml(fragment)`
for markup the app built. A new field added to a report should be one or the other -
not a bare `${...}`. The behavioural tests above pin the rule for the pure function,
and the marker makes the rest auditable by grep.

If this ever needs automated enforcement, the route that would work is component
tests that mount each builder, feed a hostile car/client/record, and assert on the
generated HTML - not static analysis of template literals.
