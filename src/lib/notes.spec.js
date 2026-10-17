import { describe, it, expect } from 'vitest'
import { parseNotes, extractNoteTexts, noteText, formatNotes } from './notes'

describe('parseNotes', () => {
  it('returns [] for empty values', () => {
    expect(parseNotes(null)).toEqual([])
    expect(parseNotes(undefined)).toEqual([])
    expect(parseNotes('')).toEqual([])
  })

  it('parses a JSON array of entries', () => {
    const raw = '[{"id_user":1,"note":"hello","timestamp":"t"}]'
    expect(parseNotes(raw)).toEqual([{ id_user: 1, note: 'hello', timestamp: 't' }])
  })

  it('accepts an already-parsed array', () => {
    const arr = [{ note: 'hello' }]
    expect(parseNotes(arr)).toBe(arr)
  })

  it('wraps a root object', () => {
    expect(parseNotes('{"note":"hello"}')).toEqual([{ note: 'hello' }])
  })

  it('wraps plain text', () => {
    expect(parseNotes('plain text')).toEqual([
      { id_user: null, note: 'plain text', timestamp: null },
    ])
  })

  it('does not throw on malformed JSON', () => {
    expect(parseNotes('[broken')).toEqual([{ id_user: null, note: '[broken', timestamp: null }])
  })
})

describe('extractNoteTexts', () => {
  it('extracts the note text from entries', () => {
    const raw = '[{"id_user":1,"note":"first"},{"id_user":2,"note":"second"}]'
    expect(extractNoteTexts(raw)).toEqual(['first', 'second'])
  })

  it('handles plain text', () => {
    expect(extractNoteTexts('just text')).toEqual(['just text'])
  })

  it('walks nested arrays', () => {
    const raw = '[{"id_user":1,"note":[{"note":"nested"}]}]'
    expect(extractNoteTexts(raw)).toEqual(['nested'])
  })

  it('walks nested objects', () => {
    const raw = '[{"note":{"note":"deep"}}]'
    expect(extractNoteTexts(raw)).toEqual(['deep'])
  })

  it('unwraps double-encoded notes (note is a JSON string)', () => {
    const inner = JSON.stringify([{ id_user: null, note: 'FOB -MER', timestamp: 't' }])
    const raw = JSON.stringify([{ id_user: 1, note: inner, timestamp: 't2' }])
    expect(extractNoteTexts(raw)).toEqual(['FOB -MER'])
  })

  it('walks nested object with an array inside a stringified note', () => {
    const inner = JSON.stringify([
      { id_user: null, note: 'with spare tyre' },
      { id_user: 5, note: 'FOB' },
    ])
    const raw = JSON.stringify([{ id_user: 1, note: inner, timestamp: 't' }])
    expect(extractNoteTexts(raw)).toEqual(['with spare tyre', 'FOB'])
  })

  it('ignores empty and whitespace-only notes', () => {
    const raw = '[{"note":"  "},{"note":"real"}]'
    expect(extractNoteTexts(raw)).toEqual(['real'])
  })

  it('ignores entries with no note field', () => {
    expect(extractNoteTexts('[{"id_user":1,"timestamp":"t"}]')).toEqual([])
  })

  it('coerces non-string primitives', () => {
    expect(extractNoteTexts('[{"note":123}]')).toEqual(['123'])
  })

  it('keeps text that merely looks like JSON but is not', () => {
    expect(extractNoteTexts('[{"note":"[not json}"}]')).toEqual(['[not json}'])
  })
})

describe('noteText', () => {
  it('returns the text of a single entry', () => {
    expect(noteText({ id_user: 1, note: 'hello' })).toBe('hello')
  })

  it('flattens a double-encoded single entry', () => {
    expect(noteText({ note: '[{"note":"inner"}]' })).toBe('inner')
  })

  it('handles a plain string', () => {
    expect(noteText('plain')).toBe('plain')
  })
})

describe('formatNotes', () => {
  it('joins all note texts with newlines', () => {
    const raw = '[{"note":"a"},{"note":"b"}]'
    expect(formatNotes(raw)).toBe('a\nb')
  })

  it('merges a second notes cell (sell bill notes)', () => {
    const carNotes = '[{"note":"car note"}]'
    const sellNotes = '[{"note":"sell note"}]'
    expect(formatNotes(carNotes, sellNotes)).toBe('car note\nsell note')
  })

  it('returns an empty string when there is nothing', () => {
    expect(formatNotes(null)).toBe('')
    expect(formatNotes('[{"id_user":1}]')).toBe('')
  })

  it('handles the mig_26 double-encoded sell bill notes', () => {
    const inner = JSON.stringify([
      { id_user: null, note: 'FOB -MER20151229002', timestamp: '2026-01-18 16:53:53' },
    ])
    const raw = JSON.stringify([{ id_user: 1, note: inner, timestamp: '2025-12-29 10:12:54' }])
    expect(formatNotes(raw)).toBe('FOB -MER20151229002')
  })
})
