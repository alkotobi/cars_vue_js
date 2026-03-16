/**
 * Generate HTML content for printing a loading record.
 * This function is pure: it only depends on its arguments.
 *
 * @param {Object} loadingRecord
 * @param {Object<string, {id:number,name?:string,ref_container?:string,so?:string,is_released?:number|boolean,date_on_board?:string|boolean,note?:string,cars:any[]}>} containersData
 * @param {string} letterheadHtml
 * @param {string} paymentStatusLabel
 * @param {Record<string, boolean>} options
 * @param {(path: string) => string} [getFileUrl] optional helper for client ID image URLs
 * @returns {string}
 */
export function generateLoadingPrintContent(
  loadingRecord,
  containersData,
  letterheadHtml = '',
  paymentStatusLabel = 'Payment Status',
  options = {},
  getFileUrl
) {
  const opts = {
    carId: options.carId !== false,
    carName: options.carName !== false,
    color: options.color !== false,
    vin: options.vin !== false,
    paymentStatus: options.paymentStatus !== false,
    client: options.client !== false,
    loadingInfo: options.loadingInfo !== false,
    summary: options.summary !== false,
  }

  const containers = Object.values(containersData || {})
  const totalCars = containers.reduce((sum, container) => sum + (container.cars ? container.cars.length : 0), 0)

  const styles = `
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

  const loadingInfoHtml = opts.loadingInfo
    ? `
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
    : ''

  const columns = []
  if (opts.carId) columns.push({ key: 'carId', label: 'Car ID' })
  if (opts.carName) columns.push({ key: 'carName', label: 'Car Name' })
  if (opts.color) columns.push({ key: 'color', label: 'Color' })
  if (opts.vin) columns.push({ key: 'vin', label: 'VIN' })
  if (opts.paymentStatus) columns.push({ key: 'paymentStatus', label: paymentStatusLabel })
  if (opts.client) columns.push({ key: 'client', label: 'Client' })

  const containersHtml = containers
    .map((container) => {
      const titleParts = []
      titleParts.push(`Container: ${container.name || 'Unnamed'}`)
      if (container.ref_container) titleParts.push(`(${container.ref_container})`)
      if (container.so) titleParts.push(` - SO: ${container.so}`)
      if (container.is_released) titleParts.push(' - RELEASED')
      titleParts.push(` - Loading #${loadingRecord.id} - Container #${container.id}`)
      const containerTitle = titleParts.join(' ')

      let tableHtml = ''
      if (!container.cars || container.cars.length === 0) {
        tableHtml =
          '<p style="text-align: center; color: #666; font-style: italic;">No cars assigned to this container</p>'
      } else if (columns.length === 0) {
        tableHtml =
          '<p style="text-align: center; color: #666; font-style: italic;">No columns selected for print.</p>'
      } else {
        const headerRow = `<tr>${columns.map((c) => `<th>${c.label}</th>`).join('')}</tr>`
        const bodyRows = container.cars
          .map((car) => {
            const cells = columns.map((c) => {
              switch (c.key) {
                case 'carId':
                  return `<td>#${car.id}</td>`
                case 'carName':
                  return `<td>${car.car_name || 'N/A'}</td>`
                case 'color':
                  return `<td>${car.color || 'N/A'}</td>`
                case 'vin':
                  return `<td>${car.vin || 'N/A'}</td>`
                case 'paymentStatus':
                  return `<td>${car.payment_status || '-'}</td>`
                case 'client': {
                  const imgHtml =
                    car.id_copy_path && getFileUrl
                      ? `
                          <img 
                            src="${getFileUrl(car.id_copy_path)}" 
                            alt="Client ID" 
                            class="client-id-image"
                            onerror="this.style.display='none'"
                          />
                        `
                      : ''
                  const mobileHtml =
                    car.client_mobiles && car.client_mobiles !== 'please provide mobile'
                      ? `<div class="client-mobile"><i class="fas fa-phone"></i> <strong>Mobile:</strong> ${car.client_mobiles}</div>`
                      : ''
                  const ninHtml = car.client_nin ? `<div class="client-nin">${car.client_nin}</div>` : ''
                  return `
                    <td>
                      <div class="client-info">
                        ${imgHtml}
                        <div class="client-details">
                          <div class="client-name">${car.client_name || 'N/A'}</div>
                          ${mobileHtml}
                          <div class="client-id-no">${car.client_id_no || 'No ID'}</div>
                          ${ninHtml}
                        </div>
                      </div>
                    </td>
                  `
                }
                default:
                  return '<td></td>'
              }
            })
            return `<tr>${cells.join('')}</tr>`
          })
          .join('')

        tableHtml = `
            <table class="cars-table">
              <thead>
                ${headerRow}
              </thead>
              <tbody>
                ${bodyRows}
              </tbody>
            </table>
          `
      }

      return `
        <div class="container-section">
          <div class="container-header">
            <div class="container-title">
              ${containerTitle}
            </div>
          </div>
          ${tableHtml}
        </div>
      `
    })
    .join('')

  const summaryHtml = opts.summary
    ? `
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
            <div class="stat-number">${containers.filter((c) => c.date_on_board).length}</div>
            <div class="stat-label">On Board</div>
          </div>
          <div class="stat-item">
            <div class="stat-number">${containers.filter((c) => !c.date_on_board).length}</div>
            <div class="stat-label">Pending</div>
          </div>
        </div>
      </div>
  `
    : ''

  return `
    <!DOCTYPE html>
    <html>
    <head>
      <title>Loading Record #${loadingRecord.id} - Print</title>
      <style>
      ${styles}
      </style>
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

