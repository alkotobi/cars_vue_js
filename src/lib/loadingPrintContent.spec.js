import { describe, it, expect } from 'vitest'
import { generatePrintContent } from './loadingPrintContent.js'

/** Minimal loading record for tests */
const minimalLoadingRecord = {
  id: 1,
  date_loading: '2025-01-15',
  shipping_line_name: 'Test Line',
  freight: 100,
  loading_port_name: 'Port A',
  discharge_port_name: 'Port B',
  EDD: '2025-02-01',
  date_loaded: '2025-01-14',
  note: 'Test note',
}

/** One container with one car */
const minimalContainersData = {
  10: {
    id: 10,
    name: 'CONT-1',
    ref_container: 'REF1',
    so: 'SO123',
    is_released: false,
    date_on_board: '2025-01-16',
    cars: [
      {
        id: 100,
        car_name: 'Toyota',
        color: 'Red',
        vin: 'VIN123',
        payment_status: 'Paid',
        client_name: 'John',
        client_mobiles: '555-1234',
        client_id_no: 'ID1',
        client_nin: 'NIN1',
        id_copy_path: '/path/to/id.jpg',
      },
    ],
  },
}

const stubGetFileUrl = (path) => (path ? `https://example.com/file?path=${encodeURIComponent(path)}` : '')

describe('generatePrintContent', () => {
  it('returns a full HTML document with DOCTYPE and html/head/body', () => {
    const html = generatePrintContent(
      minimalLoadingRecord,
      minimalContainersData,
      '',
      'Payment Status',
      undefined,
      stubGetFileUrl
    )
    expect(html).toMatch(/<!DOCTYPE html>/i)
    expect(html).toContain('<html>')
    expect(html).toContain('</head>')
    expect(html).toContain('</body>')
    expect(html).toContain('</html>')
  })

  it('includes loading info block when printOptions.loadingInfo is not false', () => {
    const html = generatePrintContent(
      minimalLoadingRecord,
      minimalContainersData,
      '',
      'Payment Status',
      {},
      stubGetFileUrl
    )
    expect(html).toContain('Operation Date:')
    expect(html).toContain('Shipping Line:')
    expect(html).toContain('Test Line')
    expect(html).toContain('loading-info')
  })

  it('omits loading info block when printOptions.loadingInfo is false', () => {
    const html = generatePrintContent(
      minimalLoadingRecord,
      minimalContainersData,
      '',
      'Payment Status',
      { loadingInfo: false },
      stubGetFileUrl
    )
    expect(html).not.toContain('class="loading-info"')
    expect(html).not.toContain('Operation Date:')
  })

  it('includes all car table columns when all options are true', () => {
    const html = generatePrintContent(
      minimalLoadingRecord,
      minimalContainersData,
      '',
      'Payment Status',
      {},
      stubGetFileUrl
    )
    expect(html).toContain('<th>Car ID</th>')
    expect(html).toContain('<th>Car Name</th>')
    expect(html).toContain('<th>Color</th>')
    expect(html).toContain('<th>VIN</th>')
    expect(html).toContain('Payment Status')
    expect(html).toContain('<th>Client</th>')
    expect(html).toContain('#100')
    expect(html).toContain('Toyota')
    expect(html).toContain('Red')
    expect(html).toContain('VIN123')
    expect(html).toContain('Paid')
    expect(html).toContain('John')
  })

  it('omits car columns when corresponding options are false', () => {
    const html = generatePrintContent(
      minimalLoadingRecord,
      minimalContainersData,
      '',
      'Payment Status',
      { carId: true, carName: false, color: false, vin: false, paymentStatus: false, client: false },
      stubGetFileUrl
    )
    expect(html).toContain('<th>Car ID</th>')
    expect(html).not.toContain('<th>Car Name</th>')
    expect(html).not.toContain('<th>Color</th>')
    expect(html).not.toContain('<th>Client</th>')
    expect(html).toContain('#100')
    expect(html).not.toMatch(/Toyota/)
  })

  it('when no car columns selected, shows message instead of empty table', () => {
    const html = generatePrintContent(
      minimalLoadingRecord,
      minimalContainersData,
      '',
      'Payment Status',
      {
        carId: false,
        carName: false,
        color: false,
        vin: false,
        paymentStatus: false,
        client: false,
      },
      stubGetFileUrl
    )
    expect(html).toContain('No columns selected for print')
    expect(html).not.toContain('<th>Car ID</th>')
  })

  it('includes summary block when printOptions.summary is not false', () => {
    const html = generatePrintContent(
      minimalLoadingRecord,
      minimalContainersData,
      '',
      'Payment Status',
      {},
      stubGetFileUrl
    )
    expect(html).toContain('class="summary"')
    expect(html).toContain('Summary')
    expect(html).toContain('Containers')
    expect(html).toContain('Total Cars')
    expect(html).toContain('On Board')
    expect(html).toContain('Pending')
  })

  it('omits summary block when printOptions.summary is false', () => {
    const html = generatePrintContent(
      minimalLoadingRecord,
      minimalContainersData,
      '',
      'Payment Status',
      { summary: false },
      stubGetFileUrl
    )
    expect(html).not.toContain('class="summary"')
  })

  it('injects letterheadHtml in body', () => {
    const letter = '<div class="letterhead">Company</div>'
    const html = generatePrintContent(
      minimalLoadingRecord,
      minimalContainersData,
      letter,
      'Payment Status',
      {},
      stubGetFileUrl
    )
    expect(html).toContain(letter)
  })

  it('uses paymentStatusLabel in table header', () => {
    const html = generatePrintContent(
      minimalLoadingRecord,
      minimalContainersData,
      '',
      'Statut de paiement',
      {},
      stubGetFileUrl
    )
    expect(html).toContain('Statut de paiement')
  })

  it('uses getFileUrl for client id_copy_path when client column is included', () => {
    const html = generatePrintContent(
      minimalLoadingRecord,
      minimalContainersData,
      '',
      'Payment Status',
      {},
      stubGetFileUrl
    )
    expect(html).toContain('https://example.com/file?path=%2Fpath%2Fto%2Fid.jpg')
  })

  it('handles container with no cars', () => {
    const noCarsData = {
      10: {
        id: 10,
        name: 'Empty',
        cars: [],
      },
    }
    const html = generatePrintContent(
      minimalLoadingRecord,
      noCarsData,
      '',
      'Payment Status',
      {},
      stubGetFileUrl
    )
    expect(html).toContain('No cars assigned to this container')
  })
})
