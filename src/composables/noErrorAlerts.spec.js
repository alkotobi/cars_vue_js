import { describe, it, expect } from 'vitest'
import { readdirSync, readFileSync, statSync } from 'node:fs'
import { join } from 'node:path'

// A guard on how failures are shown, not on anything the app does at runtime.
//
// ~63 call sites render a caught API failure through the global alert(). Each was
// written on its own, and none of them knew about the dead-session case, so a session
// that expired mid-action produced the action's own error message - often blaming
// the action, never telling the reader they had to sign in again. useSessionLost.js
// now owns that event and SessionExpiredModal owns the message, and the alert gate in
// that module stops the native dialog from stacking on top of it.
//
// The gate is the fix. This file exists so the shape does not quietly grow back: a
// new `alert()` in a `catch` is not automatically wrong, but it is automatically
// invisible to the gate's authors and needs a decision, so it has to be either
// counted here or removed from the list. The list shrinks; nothing ever grows it.
//
// It is a source scan rather than a runtime assertion because the thing being
// prevented is a line of code, and no runtime behaviour distinguishes the 63 sites
// that already look like this from a 64th.

/**
 * Files with `alert()` inside a `catch`, and how many each has.
 *
 * Keyed by file with a count rather than by line number: line numbers shift on every
 * unrelated edit above them, which would make this test fail constantly for no
 * reason. A count only moves when someone adds or removes one of these calls, which
 * is exactly the change worth noticing.
 *
 * To allow a new one, add the file with its count - after checking the call really is
 * a caught-failure message and that the alert gate covers it, which it does
 * automatically. Prefer deleting the alert: inline `error.value` plus
 * sessionErrorMessage(err, ...) is what this codebase already does in 80 files, and
 * it does not interrupt the reader at all.
 */
const KNOWN_ERROR_ALERTS = {
  'src/components/buy/BuyDetailsTable.vue': 1,
  'src/components/car-stock/CarColorBulkEditForm.vue': 2,
  'src/components/car-stock/CarStockTable.vue': 7,
  'src/components/car-stock/SelectionChatButton.vue': 1,
  'src/components/car-stock/ShowSelectionsModal.vue': 4,
  'src/components/car-stock/TaskForm.vue': 2,
  'src/components/car-stock/VinAssignmentModal.vue': 2,
  'src/components/chat/AddClientToGroup.vue': 1,
  'src/components/chat/AddGroup.vue': 1,
  'src/components/chat/AddUserToGroup.vue': 1,
  'src/components/chat/ChatMessages.vue': 5,
  'src/components/chat/ChatSidebar.vue': 1,
  'src/components/chat/ChatUsersPopup.vue': 2,
  'src/components/loading/LoadingTable.vue': 5,
  'src/components/sells/SellBillsTable.vue': 3,
  'src/components/shared/NotesManagementModal.vue': 3,
  'src/components/teams/AddEditTeamForm.vue': 1,
  'src/components/teams/TeamMembersModal.vue': 2,
  'src/components/teams/TeamMembersTable.vue': 2,
  'src/views/BuyBillPaymentsView.vue': 2,
  'src/views/BuyView.vue': 5,
  'src/views/ClientsView.vue': 1,
  'src/views/SellBillsView.vue': 1,
  'src/views/TasksView.vue': 5,
  'src/views/TeamsView.vue': 1,
}

/**
 * Every .vue and .js file under a directory, skipping node_modules and specs.
 *
 * Spec files have to be excluded: the cases below contain the very pattern being
 * scanned for as string fixtures, and a scanner that counted its own test data would
 * both fail on itself and be one edit away from counting a fixture as real code.
 */
function* sourceFiles(dir) {
  for (const entry of readdirSync(dir)) {
    const path = join(dir, entry)
    if (statSync(path).isDirectory()) {
      if (entry !== 'node_modules') yield* sourceFiles(path)
    } else if (/\.(vue|js)$/.test(entry) && !/\.spec\.js$/.test(entry)) {
      yield path
    }
  }
}

/**
 * Blank out comments and string bodies, keeping every character offset intact.
 *
 * Needed because the brace walk below infers structure from `{` and `}`; a brace
 * inside a string or a comment would send it off the rails and every later alert in
 * the file would be misattributed.
 */
function scrub(source) {
  const blanks = (match) => ' '.repeat(match.length)
  return source
    .replace(/\/\*[\s\S]*?\*\//g, blanks)
    .replace(/\/\/[^\n]*/g, blanks)
    .replace(/'(?:\\.|[^'\\])*'/g, (m) => "'" + blanks(m.slice(1, -1)) + "'")
    .replace(/"(?:\\.|[^"\\])*"/g, (m) => '"' + blanks(m.slice(1, -1)) + '"')
    .replace(/`(?:\\.|[^`\\])*`/g, (m) => '`' + blanks(m.slice(1, -1)) + '`')
}

/**
 * Count `alert()` calls lexically inside a `catch` block.
 *
 * The `<script>` block only: a `.vue` template is full of `{` in `:class` bindings
 * that would wreck the depth count, and a template cannot open a catch block anyway.
 *
 * Structural rather than a "within N lines of a catch" window, because a window
 * counts alerts that merely follow a catch - `if (errors.length) alert(...)` in a
 * later block is not the caught error being reported, and including those inflated
 * the real figure from 63 to 99.
 */
function countAlertsInCatch(source) {
  const script = source.match(/<script[^>]*>([\s\S]*?)<\/script>/)
  const code = scrub(script ? script[1] : source)

  // Each entry is the block we are directly inside, so `catch` is a marker the walk
  // can compare against.
  const stack = []
  let count = 0

  for (const match of code.matchAll(/\{|\}|\balert\b|\.alert\b/g)) {
    const token = match[0]
    if (token === '{') {
      const opensCatch = /catch\s*\([^)]*\)\s*$/.test(code.slice(0, match.index))
      stack.push(opensCatch ? 'catch' : 'block')
    } else if (token === '}') {
      stack.pop()
    } else if (token === 'alert' && stack[stack.length - 1] === 'catch') {
      count += 1
    }
  }

  return count
}

function errorAlertsInRepo() {
  const found = {}
  for (const file of sourceFiles('src')) {
    const count = countAlertsInCatch(readFileSync(file, 'utf8'))
    if (count > 0) found[file] = count
  }
  return found
}

describe('no new caught failure is reported through a native dialog', () => {
  const actual = errorAlertsInRepo()

  it('finds the calls it is guarding', () => {
    // A scanner that silently matches nothing would pass every case below for the
    // wrong reason, so it has to be shown finding something first.
    expect(Object.keys(actual).length).toBeGreaterThan(10)
    expect(Object.values(actual).reduce((a, b) => a + b, 0)).toBeGreaterThan(30)
  })

  it('adds no file to the list', () => {
    const added = Object.keys(actual).filter((file) => !(file in KNOWN_ERROR_ALERTS))
    expect(
      added,
      `alert() inside a catch was added to:\n${added.join('\n')}\n` +
        'Either render it inline (error.value = sessionErrorMessage(err, ...) - ' +
        'see CarFilesManagement.vue) or add the file to KNOWN_ERROR_ALERTS above ' +
        'after checking the alert gate covers it.',
    ).toEqual([])
  })

  it('adds no call to a file already on the list', () => {
    const grew = Object.entries(actual)
      .filter(([file, count]) => count > KNOWN_ERROR_ALERTS[file])
      .map(([file, count]) => `${file}: ${KNOWN_ERROR_ALERTS[file]} -> ${count}`)
    expect(
      grew,
      `alert() count grew inside:\n${grew.join('\n')}\n` +
        'Bump the count deliberately, or better, remove the alert.',
    ).toEqual([])
  })

  it('lists no file that has been cleaned up', () => {
    // Keeps the list honest in the other direction too: a file whose alerts were all
    // deleted should leave the list, otherwise it decays into a permanent exemption.
    const stale = Object.keys(KNOWN_ERROR_ALERTS).filter(
      (file) => !(file in actual) || actual[file] < KNOWN_ERROR_ALERTS[file],
    )
    expect(
      stale,
      `These files no longer have that many caught alerts - lower or delete them:\n${stale.join('\n')}`,
    ).toEqual([])
  })

  it('does not confuse the rest of alert() use for an error report', () => {
    // The scan has to stay narrow. ~340 alerts exist across 47 files; almost all are
    // confirmations, success messages and input validation, which are legitimate and
    // must not be dragged into this.
    const confirmInCatch = countAlertsInCatch("try { x() } catch (e) { confirm('really?') }")
    expect(confirmInCatch).toBe(0)
    // An alert in a success path is equally not an error report.
    expect(countAlertsInCatch("try { x() } catch (e) {} if (ok) { alert('Saved') }")).toBe(0)
    // Nor is one in a try block reporting a failed envelope - that is a different
    // shape, handled by apiErrorText() rather than the alert gate.
    expect(countAlertsInCatch('try { if (!r.success) { alert(r.error) } } catch (e) {}')).toBe(0)
  })

  it('is not fooled by braces in strings, comments or templates', () => {
    // Each of these would desynchronise a naive brace count and silently zero out the
    // rest of the file, which is how a guard like this quietly stops guarding.
    expect(countAlertsInCatch("const s = '{ }'\ntry { x() } catch (e) { alert('boom') }")).toBe(1)
    expect(countAlertsInCatch("// } catch (e) {\ntry { x() } catch (e) { alert('boom') }")).toBe(1)
    expect(countAlertsInCatch('/* { */\ntry { x() } catch (e) { alert("boom") }')).toBe(1)
    expect(countAlertsInCatch("try { x() } catch (e) { alert('}') }")).toBe(1)
    const vue = [
      '<template>',
      '  <div :class="{ active: on, off: !on }">{{ a ? 1 : 2 }}</div>',
      '</template>',
      '<script setup>',
      "try { x() } catch (e) { alert('boom') }",
      '</script>',
    ].join('\n')
    expect(countAlertsInCatch(vue)).toBe(1)
  })

  it('leaves the alert gate as the thing that makes the list survivable', () => {
    const lost = readFileSync(new URL('./useSessionLost.js', import.meta.url), 'utf8')
    expect(lost).toContain('export function armSessionAlertGate')
    expect(lost).toContain('export function releaseSessionAlertGate')
    // confirm() must stay reachable: these dialogs ask the reader something.
    expect(lost).not.toMatch(/window\.confirm\s*=/)
  })
})
