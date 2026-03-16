/**
 * Generates HTML print content for a loading record.
 * Pure function: no side effects; getFileUrl is injected for testability and to avoid importing useApi in lib.
 */

import { getDefaultPrintOptions } from './loadingPrintOptions.js'

/**
 * Build the CSS block for the print document (shared by all outputs).
 * @returns {string} CSS string
 */
function getPrintStyles() {
  return `
        body {
          font-family: Arial, sans-serif;
          margin: 20px;
          line-height: 1.4;
        }
        .header {
          text-align: center;
          margin-bottom: 30px;
          border-bottom: 2px solid #333;
          padding-bottom: 20px;
        }
        .header h1 {
          margin: 0;
          color: #333;
          font-size: 24px;
        }
        .loading-info {
          display: grid;
          grid-template-columns: 1fr 1fr;
          gap: 20px;
          margin-bottom: 30px;
          background: #f8f9fa;
          padding: 20px;
          border-radius: 8px;
        }
        .info-group {
          margin-bottom: 15px;
        }
        .info-label {
          font-weight: bold;
          color: #555;
          margin-bottom: 5px;
        }
        .info-value {
          color: #333;
        }
        .container-section {
          margin-bottom: 40px;
          page-break-inside: avoid;
        }
        .container-header {
          background: #e3f2fd;
          padding: 15px;
          border-radius: 8px;
          margin-bottom: 15px;
          border-left: 4px solid #2196f3;
        }
        .container-title {
          font-size: 18px;
          font-weight: bold;
          margin: 0 0 10px 0;
          color: #1976d2;
        }
        .cars-table {
          width: 100%;
          border-collapse: collapse;
          margin-top: 10px;
        }
        .cars-table th {
          background: #f5f5f5;
          padding: 10px;
          text-align: left;
          border: 1px solid #ddd;
          font-weight: bold;
        }
        .cars-table td {
          padding: 8px 10px;
          border: 1px solid #ddd;
          vertical-align: top;
        }
        .cars-table tr:nth-child(even) {
          background: #f9f9f9;
        }
        .client-id-image {
          width: 100px;
          height: 70px;
          object-fit: cover;
          border-radius: 6px;
          border: 2px solid #ddd;
          cursor: pointer;
        }
        .client-info {
          display: flex;
          align-items: center;
          gap: 8px;
        }
        .client-details {
          display: flex;
          flex-direction: column;
          gap: 2px;
        }
        .client-name {
          font-weight: 500;
        }
        .client-id-no {
          font-size: 0.85rem;
          color: #666;
        }
        .client-mobile {
          font-size: 0.8rem;
          color: #1e40af;
          font-weight: 500;
          margin-top: 2px;
          display: flex;
          align-items: center;
          gap: 4px;
          background: #eff6ff;
          padding: 4px 8px;
          border-radius: 4px;
          border-left: 3px solid #3b82f6;
        }
        .client-mobile i {
          font-size: 0.7rem;
          color: #3b82f6;
        }
        .client-mobile strong {
          color: #1e40af;
          font-weight: 600;
        }
        .client-nin {
          font-size: 0.75rem;
          color: #1e40af;
          font-weight: 600;
          background: #dbeafe;
          border: 1px solid #93c5fd;
          border-radius: 4px;
          padding: 2px 4px;
          font-family: 'Courier New', monospace;
          text-align: center;
          margin-top: 2px;
          display: inline-block;
          width: fit-content;
        }
        .summary {
          margin-top: 30px;
          padding: 20px;
          background: #e8f5e8;
          border-radius: 8px;
          border-left: 4px solid #4caf50;
        }
        .summary h3 {
          margin: 0 0 15px 0;
          color: #2e7d32;
        }
        .summary-stats {
          display: grid;
          grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
          gap: 15px;
        }
        .stat-item {
          text-align: center;
        }
        .stat-number {
          font-size: 24px;
          font-weight: bold;
          color: #2e7d32;
        }
        .stat-label {
          font-size: 12px;
          color: #666;
          text-transform: uppercase;
        }
        @media print {
          body { margin: 0; }
          .container-section { page-break-inside: avoid; }
        }
  `
}

/**
 * Build the loading info block HTML (operation date, shipping line, ports, etc.).
 * @param {Record<string, unknown>} loadingRecord
 * @returns {string}
 */
function buildLoadingInfoBlock(loadingRecord) {
  return `
      <div class="loading-info">
        <div class="info-group">
          <div class="info-label">Operation Date:</div>
          <div class="info-value">${loadingRecord.date_loading ? new Date(loadingRecord.date_loading).toLocaleDateString() : 'Not set'}</div>
        </div>
        <div class="info-group">
          <div class="info-label">Shipping Line:</div>
          <div class="info-value">${loadingRecord.shipping_line_name || 'Not set'}</div>
        </div>
        <div class="info-group">
          <div class="info-label">Freight:</div>
          <div class="info-value">${loadingRecord.freight ? `$${loadingRecord.freight}` : 'Not set'}</div>
        </div>
        <div class="info-group">
          <div class="info-label">Loading Port:</div>
          <div class="info-value">${loadingRecord.loading_port_name || 'Not set'}</div>
        </div>
        <div class="info-group">
          <div class="info-label">Discharge Port:</div>
          <div class="info-value">${loadingRecord.discharge_port_name || 'Not set'}</div>
        </div>
        <div class="info-group">
          <div class="info-label">EDD:</div>
          <div class="info-value">${loadingRecord.EDD ? new Date(loadingRecord.EDD).toLocaleDateString() : 'Not set'}</div>
        </div>
        <div class="info-group">
          <div class="info-label">Loaded Date:</div>
          <div class="info-value">${loadingRecord.date_loaded ? new Date(loadingRecord.date_loaded).toLocaleDateString() : 'Not set'}</div>
        </div>
        <div class="info-group">
          <div class="info-label">Notes:</div>
          <div class="info-value">${loadingRecord.note || 'No notes'}</div>
        </div>
      </div>
  `
}

/**
 * Build table header row for cars table based on print options.
 * @param {Record<string, boolean>} opts
 * @param {string} paymentStatusLabel
 * @returns {{ headers: string[], keys: string[] }} keys are the option keys for building cells
 */
function getCarTableHeaderParts(opts, paymentStatusLabel) {
  const headers = []
  const keys = []
  if (opts.carId) {
    headers.push('Car ID')
    keys.push('carId')
  }
  if (opts.carName) {
    headers.push('Car Name')
    keys.push('carName')
  }
  if (opts.color) {
    headers.push('Color')
    keys.push('color')
  }
  if (opts.vin) {
    headers.push('VIN')
    keys.push('vin')
  }
  if (opts.paymentStatus) {
    headers.push(paymentStatusLabel)
    keys.push('paymentStatus')
  }
  if (opts.client) {
    headers.push('Client')
    keys.push('client')
  }
  return { headers, keys }
}

/**
 * Build a single car row HTML based on selected columns.
 * @param {Record<string, unknown>} car
 * @param {string[]} keys - from getCarTableHeaderParts
 * @param {(path: string) => string} getFileUrl
 * @returns {string}
 */
function buildCarRow(car, keys, getFileUrl) {
  const cells = []
  for (const k of keys) {
    if (k === 'carId') {
      cells.push(`<td>#${car.id}</td>`)
    } else if (k === 'carName') {
      cells.push(`<td>${car.car_name || 'N/A'}</td>`)
    } else if (k === 'color') {
      cells.push(`<td>${car.color || 'N/A'}</td>`)
    } else if (k === 'vin') {
      cells.push(`<td>${car.vin || 'N/A'}</td>`)
    } else if (k === 'paymentStatus') {
      cells.push(`<td>${car.payment_status || '-'}</td>`)
    } else if (k === 'client') {
      const imgHtml =
        car.id_copy_path && getFileUrl
          ? `<img src="${getFileUrl(car.id_copy_path)}" alt="Client ID" class="client-id-image" onerror="this.style.display='none'" />`
          : ''
      const mobileHtml =
        car.client_mobiles && car.client_mobiles !== 'please provide mobile'
          ? `<div class="client-mobile"><i class="fas fa-phone"></i> <strong>Mobile:</strong> ${car.client_mobiles}</div>`
          : ''
      const ninHtml = car.client_nin ? `<div class="client-nin">${car.client_nin}</div>` : ''
      cells.push(
        `<td>
          <div class="client-info">
            ${imgHtml}
            <div class="client-details">
              <div class="client-name">${car.client_name || 'N/A'}</div>
              ${mobileHtml}
              <div class="client-id-no">${car.client_id_no || 'No ID'}</div>
              ${ninHtml}
            </div>
          </div>
        </td>`
      )
    }
  }
  return `<tr>${cells.join('')}</tr>`
}

/**
 * Build the summary block HTML (containers count, total cars, on board, pending).
 * @param {Array<{ date_on_board?: unknown }>} containers
 * @param {number} totalCars
 * @returns {string}
 */
function buildSummaryBlock(containers, totalCars) {
  const onBoard = containers.filter((c) => c.date_on_board).length
  const pending = containers.filter((c) => !c.date_on_board).length
  return `
      <div class="summary">
        <h3>Summary</h3>
        <div class="summary-stats">
          <div class="stat-item">
            <div class="stat-number">${containers.length}</div>
            <div class="stat-label">Containers</div>
          </div>
          <div class="stat-item">
            <div class="stat-number">${totalCars}</div>
            <div class="stat-label">Total Cars</div>
          </div>
          <div class="stat-item">
            <div class="stat-number">${onBoard}</div>
            <div class="stat-label">On Board</div>
          </div>
          <div class="stat-item">
            <div class="stat-number">${pending}</div>
            <div class="stat-label">Pending</div>
          </div>
        </div>
      </div>
  `
}

/**
 * Generates full HTML document string for printing a loading record.
 * Respects printOptions to include or omit loading info block, car table columns, and summary.
 * If no car columns are selected, the cars table is replaced by a short message.
 *
 * @param {Record<string, unknown>} loadingRecord - The loading record (id, date_loading, shipping_line_name, etc.).
 * @param {Record<string, { id: number; name?: string; ref_container?: string; so?: string; is_released?: boolean; date_on_board?: unknown; cars: Array<Record<string, unknown>> }>} containersData - Map of container id to container object with cars array.
 * @param {string} [letterheadHtml=''] - Pre-built letterhead HTML fragment.
 * @param {string} [paymentStatusLabel='Payment Status'] - Label for the payment status column.
 * @param {Record<string, boolean>} [printOptions] - Which columns/sections to include; defaults to all true.
 * @param {(path: string) => string} [getFileUrl] - Function to resolve file path to URL (e.g. for client ID image). Optional; if omitted, client images are skipped.
 * @returns {string} Full HTML document string.
 */
export function generatePrintContent(
  loadingRecord,
  containersData,
  letterheadHtml = '',
  paymentStatusLabel = 'Payment Status',
  printOptions = undefined,
  getFileUrl = undefined
) {
  const opts = printOptions && typeof printOptions === 'object' ? { ...getDefaultPrintOptions(), ...printOptions } : getDefaultPrintOptions()
  const containers = Object.values(containersData)
  const totalCars = containers.reduce((sum, container) => sum + container.cars.length, 0)

  const loadingInfoHtml =
    opts.loadingInfo !== false ? buildLoadingInfoBlock(loadingRecord) : ''

  const { headers: carHeaders, keys: carKeys } = getCarTableHeaderParts(opts, paymentStatusLabel)
  const hasCarColumns = carHeaders.length > 0

  const containersHtml = containers
    .map((container) => {
      const containerTitle = `Container: ${container.name || 'Unnamed'} ${container.ref_container ? `(${container.ref_container})` : ''}${container.so ? ` - SO: ${container.so}` : ''}${container.is_released ? ' - RELEASED' : ''} - Loading #${loadingRecord.id} - Container #${container.id}`
      let tableHtml
      if (container.cars.length === 0) {
        tableHtml =
          '<p style="text-align: center; color: #666; font-style: italic;">No cars assigned to this container</p>'
      } else if (!hasCarColumns) {
        tableHtml =
          '<p style="text-align: center; color: #666; font-style: italic;">No columns selected for print.</p>'
      } else {
        const headerRow = `<tr>${carHeaders.map((h) => `<th>${h}</th>`).join('')}</tr>`
        const bodyRows = container.cars
          .map((car) => buildCarRow(car, carKeys, getFileUrl))
          .join('')
        tableHtml = `
            <table class="cars-table">
              <thead>${headerRow}</thead>
              <tbody>${bodyRows}</tbody>
            </table>
          `
      }
      return `
        <div class="container-section">
          <div class="container-header">
            <div class="container-title">${containerTitle}</div>
          </div>
          ${tableHtml}
        </div>
      `
    })
    .join('')

  const summaryHtml =
    opts.summary !== false ? buildSummaryBlock(containers, totalCars) : ''

  return `
    <!DOCTYPE html>
    <html>
    <head>
      <title>Loading Record #${loadingRecord.id} - Print</title>
      <style>${getPrintStyles()}</style>
    </head>
    <body>
      ${letterheadHtml}
      <div class="header">
        <h1>Loading Record #${loadingRecord.id}</h1>
        <p>Generated on ${new Date().toLocaleDateString()} at ${new Date().toLocaleTimeString()}</p>
      </div>

      ${loadingInfoHtml}

      ${containersHtml}

      ${summaryHtml}
    </body>
    </html>
  `
}
