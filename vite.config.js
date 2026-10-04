import { createReadStream, existsSync, readFileSync, rmSync, statSync } from 'node:fs'
import { extname, join, normalize, resolve, sep } from 'node:path'
import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// Build ID so every build produces different filenames (avoids stale cache after deploy)
const buildId = Date.now().toString(36)

// Plugin to remove vendor chunk preload links
const removeVendorPreload = () => {
  return {
    name: 'remove-vendor-preload',
    transformIndexHtml(html) {
      // Remove preload/modulepreload links for vendor chunks
      // This prevents browser warnings about unused preloaded resources
      return html.replace(
        /<link[^>]*rel=["'](modulepreload|preload)["'][^>]*vendor[^>]*>/gi,
        ''
      )
    },
  }
}

// db_code.json is PER SERVER, not per build: it names the database a deployment
// talks to. public/db_code.json exists so `npm run dev` works locally, but it
// must never reach dist/ — deploying it would point every client at this
// machine's database. deploy/deploy.sh writes the real one on the server and
// fails the deploy if this file leaks into a build.
const removeDbCode = () => {
  return {
    name: 'remove-db-code-json',
    apply: 'build',
    closeBundle() {
      const target = fileURLToPath(new URL('./dist/db_code.json', import.meta.url))
      if (existsSync(target)) {
        rmSync(target)
        this.warn('removed dist/db_code.json (per-server file, written by deploy/deploy.sh)')
      }
    },
  }
}

// Content types for a prebuilt tenant folder. Deliberately explicit: this
// middleware answers the request itself, so there is no transform pipeline to
// fall back on, and a wrong type here is what produces the classic "Failed to
// load module script: expected JavaScript MIME type" blank page.
const FOLDER_MOUNT_MIME = {
  '.html': 'text/html; charset=utf-8',
  '.js': 'text/javascript; charset=utf-8',
  '.mjs': 'text/javascript; charset=utf-8',
  '.css': 'text/css; charset=utf-8',
  '.json': 'application/json; charset=utf-8',
  '.map': 'application/json; charset=utf-8',
  '.txt': 'text/plain; charset=utf-8',
  '.svg': 'image/svg+xml',
  '.png': 'image/png',
  '.jpg': 'image/jpeg',
  '.jpeg': 'image/jpeg',
  '.gif': 'image/gif',
  '.webp': 'image/webp',
  '.ico': 'image/x-icon',
  '.pdf': 'application/pdf',
  '.woff': 'font/woff',
  '.woff2': 'font/woff2',
  '.ttf': 'font/ttf',
  '.eot': 'application/vnd.ms-fontobject',
}

// Serve a prebuilt app folder (a second tenant) as its own mount on the dev
// server, e.g. http://localhost:5173/mig_27/ next to the root app on /.
//
// Each tenant folder holds its own build, its own db_code.json and its own copy
// of api/ - that is the layout deploy/deploy.sh rsyncs to a client server, and
// it is why the tenant's PHP resolves a different database (api/lib/appdb.php
// reads <appfolder>/db_code.json). The dev server could not serve one:
//
//   - the folder's index.html went through vite's html pipeline, which rewrote
//     the build's relative `./index.<hash>.js` to `/index.<hash>.js` - the root
//     app's folder - and served its <link rel=stylesheet> as a JS module;
//   - any path under the mount with no file behind it hit the SPA fallback and
//     was answered with the ROOT index.html, so /mig_27/cars loaded the root app.
//
// So this middleware answers everything under the mount itself and calls next()
// only for <mount>/api, which the proxy owns (PHP has to execute those). The
// <base> tag is the same thing deploy/nginx-app.conf.template injects via
// sub_filter, and it is load-bearing: the built index.html references its chunk
// relatively, which resolves to the right file at depth 1 only - without it every
// deep link such as /mig_27/cars/sell-bills/5 asks for the chunk one level too
// high and comes back as this same HTML file.
// One tenant folder, resolved. The name is validated before it reaches new URL(),
// so a request path can never become part of a filesystem path unchecked.
const folderMountFor = (name) => {
  const folderRoot = fileURLToPath(new URL(`./${name}/`, import.meta.url)).replace(/[\\/]+$/, '')
  return {
    name,
    mount: `/${name}`,
    folderRoot,
    indexFile: join(folderRoot, 'index.html'),
    isInside: (candidate) => candidate === folderRoot || candidate.startsWith(folderRoot + sep),
  }
}

const isDirectory = (path) => existsSync(path) && statSync(path).isDirectory()

// Resolved per request, not once at config load. A list built when the process starts
// is a snapshot: a client provisioned while the dev server was already running is on
// disk but not in the list, so its API had no proxy entry and its folder was not a
// mount. Vite answered the api.php request itself, and because PHP is denied as a
// document (server.fs.deny below) that surfaced as
//
//   403 The request url "/Users/.../nn/api/api.php" is outside of Vite serving allow list
//
// which reads like a permissions problem and is not one. Two stat() calls per request
// is cheaper than a restart, and it means provisioning a client is immediately enough.
const resolveFolderMount = (name) => {
  if (!/^[a-z][a-z0-9_]*$/.test(name)) return null

  const root = fileURLToPath(new URL('.', import.meta.url))
  // db_code.json is api/lib/appdb.php's test for a provisioned deployment; api/ is
  // what the API proxy below needs to reach. Both, because db_code.json alone also
  // matches public/ (the root app's own copy, which is in git) and mig1/, a folder
  // with neither a build nor an api/.
  if (!existsSync(join(root, name, 'db_code.json')) || !isDirectory(join(root, name, 'api'))) return null

  return folderMountFor(name)
}

const serveFolderMounts = {
  name: 'serve-folder-mounts',

  configureServer(server) {
    // Registered here rather than in a post hook: plugin middlewares added directly
    // run BEFORE vite's transform, static and html-fallback middlewares, which is
    // the whole point - those are what used to rewrite and replace this app.
    server.middlewares.use((req, res, next) => {
      if (req.method !== 'GET' && req.method !== 'HEAD') return next()

      const url = (req.url || '').split('?')[0]
      const slash = url.indexOf('/', 1)
      const firstSegment = slash === -1 ? url.slice(1) : url.slice(1, slash)

      let tenant
      try {
        tenant = resolveFolderMount(decodeURIComponent(firstSegment))
      } catch {
        tenant = null
      }
      if (!tenant) return next()

      const { mount, folderRoot, indexFile, isInside } = tenant

      // Send the built index.html, with the mount injected as <base>.
      const sendIndex = () => {
        let html
        try {
          html = readFileSync(indexFile, 'utf8')
        } catch {
          res.statusCode = 404
          res.setHeader('Content-Type', 'text/plain; charset=utf-8')
          res.end(`${mount} is not built yet - run: npm run build\n`)
          return
        }

        if (!/<base\s/i.test(html)) {
          const baseTag = `<base href="${mount}/">`
          html = /<head[^>]*>/i.test(html)
            ? html.replace(/<head[^>]*>/i, (tag) => `${tag}\n    ${baseTag}`)
            : `${baseTag}\n${html}`
        }

        // no-store: this is a build snapshot, and a cached index.html would pin the
        // browser to the previous set of hashed filenames after the next rebuild.
        res.setHeader('Content-Type', FOLDER_MOUNT_MIME['.html'])
        res.setHeader('Cache-Control', 'no-store')
        res.end(html)
      }

      // Without the trailing slash the browser would resolve the build's relative
      // asset URLs against the parent folder.
      if (url === mount) {
        res.statusCode = 302
        res.setHeader('Location', `${mount}/`)
        res.end()
        return
      }
      if (!url.startsWith(`${mount}/`)) return next()

      // The tenant's own API. Handled by server.proxy, which runs after this.
      if (url === `${mount}/api` || url.startsWith(`${mount}/api/`)) return next()

      let relative
      try {
        // Trailing slashes go: '/nono/' would otherwise resolve to '/' and be refused
        // as an escape, instead of the folder index.
        relative = decodeURIComponent(url.slice(mount.length + 1)).replace(/\/+$/, '')
      } catch {
        res.statusCode = 400
        res.end('Bad request')
        return
      }
      if (relative.includes('\0')) {
        res.statusCode = 400
        res.end('Bad request')
        return
      }

      // normalize() collapses "..", and the isInside() check is what actually proves
      // the result cannot leave the folder.
      const target = resolve(folderRoot, normalize(relative))
      if (!isInside(target)) {
        res.statusCode = 403
        res.end('Forbidden')
        return
      }

      // A real file wins; anything else that is not obviously an asset is an app
      // route and gets the shell (so /nono/cars works like /cars does).
      if (existsSync(target) && statSync(target).isFile()) {
        const type = FOLDER_MOUNT_MIME[extname(target).toLowerCase()]
        if (!type) {
          res.statusCode = 404
          res.setHeader('Content-Type', 'text/plain; charset=utf-8')
          res.end('Not found')
          return
        }
        res.setHeader('Content-Type', type)
        res.setHeader('Content-Length', statSync(target).size)
        createReadStream(target).pipe(res)
        return
      }
      if (extname(target) === '' || extname(target) === '.html') {
        sendIndex()
        return
      }

      res.statusCode = 404
      res.setHeader('Content-Type', 'text/plain; charset=utf-8')
      res.end('Not found')
    })
  },
}

// Get environment
const isProduction = process.env.NODE_ENV === 'production'

// Set API URL based on environment
// Note: This is only used for the dev server proxy during development
// In production, the app uses relative URLs (window.location.origin)
// so it works from any client domain without code changes
const apiUrl = 'http://localhost:8000'

// Which folders count as tenant mounts is decided per request, in
// resolveFolderMount() above. There is deliberately no list here: this was
// ['mig_27'], and every client provisioned after it was written had no mount and no
// API proxy entry, so the dev server answered its api.php itself. Two rounds of that
// looked like two different bugs - first the file's own PHP source came back with a
// 200, and after server.fs.deny below started refusing to serve PHP as a document,
// a 403 saying the request was "outside of Vite serving allow list". Neither pointed
// at the dev server, and both were fixed by restarting it.

// The dev server mounts the app where production mounts it: /cars/.
//
// The dev URL and the production URL were different shapes for the same page - dev
// was at http://localhost:5173/login while production is
// https://host/cars/login - so every deep link, every bookmark and every bug report
// had to be translated between the two before anyone could follow it. `base` is what
// the router and every asset URL derive from (src/utils/basePath.js), so setting it
// here makes the dev server produce production-shaped URLs rather than the two being
// kept in sync by hand.
//
// It applies to `vite serve` ONLY, which is why this is keyed on `command` and not on
// NODE_ENV: the build has to stay relative, because one build is deployed to every
// tenant and each is served from a differently named folder (mig_27, b, nn, ...).
// An absolute /cars/ base would emit asset URLs that only resolve on the one server
// whose folder happens to be called cars.
//
// The tenant mounts above are unaffected: they are served by serveFolderMounts from
// their own prebuilt bundles, which already carry a relative base and get <base>
// injected per mount.
const DEV_BASE = process.env.VITE_DEV_BASE ?? '/cars/'

// The root of the dev server is not an app, and must not become one.
//
// Setting `base` makes Vite redirect / to DEV_BASE, so http://127.0.0.1:5173/ answered
// 302 to /cars/ and booted the root SPA. That SPA has no tenant - the repository root
// carries no db_code.json - so every request it makes is refused with db_unavailable.
// Its login form is a dead end that is indistinguishable from a working one right up
// until a password is typed into it, which is worse than not offering the form.
//
// Production never had this bug, and the reason is worth keeping in mind before
// "fixing" it by pointing the dev server at the root app again: the rendered nginx
// config has no `location = /` at all, so / falls through to `root __WEBROOT__;` and
// index index.html, i.e. __WEBROOT__/index.html. deploy.sh only ever rsyncs dist/ into
// $WEBROOT/$FOLDER/ and api/ into $WEBROOT/$FOLDER/api/, so no file is ever written at
// the webroot root and nginx answers 404. The dev server was the only place where /
// was an app.
//
// So / is answered here, and answered the way production answers it. Registered in
// configureServer directly rather than returned from it, because plugin middleware
// added that way runs before Vite's internal ones - including the base redirect that
// would otherwise win and 302 to /cars/ before this ever got a say.
//
// Nothing is served in its place. The webroot root is where a welcome page belongs,
// and index.html cannot be that: it is the SPA shell that the tenant mounts are built
// from, so reusing it here would put the SPA back at / under a different name. A
// welcome page needs a name that is not taken, and choosing one is a decision to make
// when there is a page to point at - not a reason to serve the shell instead.
const serveNothingAtRoot = {
  name: 'serve-nothing-at-root',

  configureServer(server) {
    server.middlewares.use((req, res, next) => {
      // /index.html is the same request spelled out: production resolves it to
      // __WEBROOT__/index.html and 404s, so it must not redirect here either.
      const pathname = (req.url || '').split('?')[0]
      if (pathname !== '/' && pathname !== '/index.html') {
        next()
        return
      }
      res.statusCode = 404
      res.end()
    })
  },
}

export default defineConfig(({ command }) => ({
  // vue-plugin-vue-devtools is intentionally not loaded: importing it evaluates
  // @vue/devtools-kit at config load, which touches localStorage and throws
  // "localStorage.getItem is not a function" under the Node build, breaking
  // `vite build` and `vitest`. Re-add it once the plugin is fixed or pinned.
  plugins: [
    vue(),
    removeVendorPreload(),
    removeDbCode(),
    serveFolderMounts,
    serveNothingAtRoot,
  ],
  test: {
    globals: true,
    environment: 'node',
    include: ['src/**/*.spec.js', 'src/**/*.test.js'],
  },
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    allowedHosts: ['python-platinum-rainbow-printed.trycloudflare.com'],
    // PHP is reached through server.proxy below, which executes it. Nothing should
    // ever be *served* as a document, because a tenant folder holds an api/ copy and
    // that copy holds config.local.php and db_manager_config.local.php - the database
    // password, in a file the dev server will happily hand over as text. The same
    // rule the rendered nginx config keeps: no generic PHP location, so a .php file
    // is either executed or unreachable.
    fs: {
      deny: ['**/*.php'],
    },
    proxy: {
      '/api': {
        target: apiUrl,
        changeOrigin: true,
        secure: isProduction,
      },
      '/uploads': {
        target: apiUrl,
        changeOrigin: true,
        secure: isProduction,
      },
      // Any folder mount's API: the api/ copy inside that folder, reached by the same
      // path - no rewrite, because PHP resolves the tenant's database from its own
      // location (dirname(__DIR__, 2) . '/db_code.json'), so the request has to keep
      // the /<folder>/ prefix to hit that copy.
      //
      // A pattern rather than one entry per client, for the same reason the middleware
      // above resolves per request: a proxy table is built when the config loads, so a
      // client provisioned afterwards had no entry, and its api.php fell through to
      // Vite's static handling. The cost is that '^/[^/]+/api' also matches a path that
      // is not a tenant mount - /src/api/..., say - and that request is forwarded to
      // the PHP server, which has no such file and answers 404. The alternative is a
      // list that is wrong by construction after the next provisioning run.
      '^/[^/]+/api': {
        target: apiUrl,
        changeOrigin: true,
        secure: isProduction,
      },
    },
  },
  base: command === 'serve' ? DEV_BASE : './',
  build: {
    rollupOptions: {
      output: {
        entryFileNames: `[name].[hash].${buildId}.js`,
        chunkFileNames: `[name].[hash].${buildId}.js`,
        assetFileNames: (assetInfo) => {
          // Preserve exact filenames for logo.png, letter_head.png, and gml2.png
          const preservedAssets = ['logo.png', 'letter_head.png', 'gml2.png']
          const assetName = assetInfo.name || ''

          // Check if the asset name ends with any of the preserved asset names
          // This handles cases where the path might be included
          const matchesPreserved = preservedAssets.some((name) => {
            return (
              assetName.endsWith(name) ||
              assetName.includes(`/${name}`) ||
              assetName.includes(`\\${name}`)
            )
          })

          if (matchesPreserved) {
            // Extract just the filename from the path
            const fileName = assetName.split('/').pop() || assetName.split('\\').pop() || assetName
            return fileName
          }

          // For all other assets: hash + buildId so names change every build
          return `[name].[hash].${buildId}.[ext]`
        },
        manualChunks: {
          vendor: ['vue', 'vue-router', 'pinia'],
          ui: ['element-plus'],
          utils: [],
        },
      },
    },
    manifest: true,
    // Off unless asked for. These maps are the whole unminified application: every
    // source file, every API path, every route and every secret-shaped string,
    // reconstructed from a fetch of /assets/<chunk>.js.map. `true` publishes them as
    // a normal file next to each chunk and adds the sourceMappingURL comment;
    // 'inline' is worse, since it embeds all of it in the bundle itself.
    //
    // `npm run build` is the deploy build, so the default has to be the safe one -
    // a debug build stays available as `VITE_SOURCEMAP=true npm run build`.
    sourcemap: process.env.VITE_SOURCEMAP === 'true',
    chunkSizeWarningLimit: 1000, // Increase warning limit to 1MB
  },
}))
