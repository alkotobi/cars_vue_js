// Feature switches, in one place, so turning something off is a one-line change
// rather than a hunt through a view.
//
// Every flag here is a compile-time constant on purpose. A server-side flag would
// mean a round trip before the suppliers table can render, and the credibility
// column would appear for a frame and then vanish. Since the app is built once
// and the same dist/ is deployed to every server, a flag here applies everywhere
// - which is the right scope for "nobody has AI configured right now".
//
// Nothing is deleted when a feature is switched off. The components, composables,
// API actions, styles and translations all stay in place and tested, so turning
// the flag back on restores the feature without a second implementation.

/**
 * Supplier credibility checks.
 *
 * Off while no server has an AI key configured: the check needs a model to ask,
 * so with AI switched off the column, the row-menu entry and the modal would only
 * ever produce an "AI not configured" error.
 *
 * To re-enable: set this to true. Nothing else needs to change.
 */
export const CREDIBILITY_ENABLED = false
