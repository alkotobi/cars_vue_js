import { createReadStream, existsSync, readFileSync, statSync } from 'node:fs'
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

// There is deliberately no plugin stripping db_code.json from the build any more:
// the file is gone. It named the database a deployment talked to, and now the tenant
// in the request does - so the build contains nothing per-server to leak, and
// public/db_code.json was removed rather than filtered.

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

// Serve a tenant as its own mount on the dev server, e.g.
// http://localhost:5173/merhab_cars/ next to the root app on /cars/.
//
// A tenant folder holds its uploads and nothing else, so the three things a mount
// has to serve come from two places:
//
//   /<tenant>/            the SHARED build in dist/
//   /<tenant>/files/...   the tenant's own files/ folder
//   /<tenant>/api/...     the SHARED api/, executed - see server.proxy
//
// Vite cannot serve any of them on its own. Its html pipeline rewrote the build's
// relative ./index.<hash>.js to /index.<hash>.js - the root app's folder - and its
// SPA fallback answered any path with no file behind it using the ROOT index.html,
// so /merhab_cars/cars loaded the root app. So this middleware answers the mount
// itself and calls next() only for <mount>/api, which the proxy owns because PHP has
// to execute those.
//
// The <base> tag is the same thing deploy/nginx-multitenant.conf.template injects via
// sub_filter, and it is load-bearing: the built index.html references its chunk
// relatively, which resolves to the right file at depth 1 only - without it every
// deep link such as /merhab_cars/cars/sell-bills/5 asks for the chunk one level too
// high and comes back as this same HTML file.
const folderMountFor = (name) => {
  const deploymentRoot = fileURLToPath(new URL('./dist/', import.meta.url)).replace(/[\\/]+$/, '')
  const tenantRoot = fileURLToPath(new URL(`./${name}/`, import.meta.url)).replace(/[\\/]+$/, '')
  return {
    name,
    mount: `/${name}`,
    folderRoot: deploymentRoot,
    tenantRoot,
    indexFile: join(deploymentRoot, 'index.html'),
    isInside: (candidate) =>
      candidate === deploymentRoot || candidate.startsWith(deploymentRoot + sep),
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
  // files/ is what makes a folder a tenant: it is where that tenant's uploads live,
  // and nothing else in this repository has one. It used to be a copy of api/ that
  // decided which database the tenant talked to, so the test asked for a copy of the
  // application per client - which is what the shared api/ removed.
  if (isDirectory(join(root, name, 'api')) || isDirectory(join(root, name, 'dist'))) return null
  if (!isDirectory(join(root, name, 'files'))) return null

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

      const { mount, folderRoot, tenantRoot, indexFile, isInside } = tenant
      // The one route under the mount that is NOT the shared build: the tenant's own
      // uploads. /<tenant>/files/cars/x.png has to come from <tenant>/files/, so it is
      // answered here rather than looked for in dist/ - where it does not exist, and
      // where it must not, since a file in the build folder is the same file for every
      // tenant.
      const filesPrefix = `${mount}/files`

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

      // The tenant's own API, executed by the shared api/. Handled by server.proxy,
      // which runs after this and keeps the /<tenant> prefix, because that prefix is
      // what tells app_db_name() which tenant the request is for.
      if (url === `${mount}/api` || url.startsWith(`${mount}/api/`)) return next()

       if (url === filesPrefix || url.startsWith(`${filesPrefix}/`)) {
         let filesRelative
         try {
           filesRelative = decodeURIComponent(url.slice(filesPrefix.length)).replace(/^\/+|\/+$/g, '')
         } catch {
           res.statusCode = 400
           res.end('Bad request')
           return
         }
         if (filesRelative === '' || filesRelative.includes('\0')) {
           res.statusCode = 404
           res.setHeader('Content-Type', 'text/plain; charset=utf-8')
           res.end('Not found')
           return
         }

         // Inside the tenant's OWN files/ folder. Containment is checked against
         // tenantRoot, not folderRoot: a request path that escapes files/ must not
         // land in the shared build or anywhere else, and normalise() alone is not
         // what proves that - it is the prefix comparison.
         const filesRoot = join(tenantRoot, 'files')
         const fileTarget = resolve(filesRoot, normalize(filesRelative))
         if (fileTarget !== filesRoot && !fileTarget.startsWith(filesRoot + sep)) {
           res.statusCode = 403
           res.end('Forbidden')
           return
         }

         // No SPA fallback here, deliberately. A missing upload is a missing upload,
         // and answering it with index.html would hand the browser an HTML file it
         // would try to render as a logo.
         if (existsSync(fileTarget) && statSync(fileTarget).isFile()) {
           const fileType = FOLDER_MOUNT_MIME[extname(fileTarget).toLowerCase()]
           if (!fileType) {
             res.statusCode = 404
             res.setHeader('Content-Type', 'text/plain; charset=utf-8')
             res.end('Not found')
             return
           }
           res.setHeader('Content-Type', fileType)
           res.setHeader('Content-Length', statSync(fileTarget).size)
           createReadStream(fileTarget).pipe(res)
           return
         }

         res.statusCode = 404
         res.setHeader('Content-Type', 'text/plain; charset=utf-8')
         res.end('Not found')
         return
       }

       // Branding assets: prefer tenant files/ first, then shared dist/
       const brandRe = /^(logo|logo_default|letter_head|letter_head_default|gml2)\.png$/
       let relBrand
       try {
         relBrand = decodeURIComponent(url.slice(mount.length + 1)).replace(/\/+$/, '')
       } catch {
         relBrand = ''
       }
       if (relBrand && brandRe.test(relBrand)) {
         const brandFile = relBrand
         const tenantBrand = join(tenantRoot, 'files', brandFile)
         if (existsSync(tenantBrand) && statSync(tenantBrand).isFile()) {
           const type = FOLDER_MOUNT_MIME['.png']
           res.setHeader('Content-Type', type)
           res.setHeader('Content-Length', statSync(tenantBrand).size)
           createReadStream(tenantBrand).pipe(res)
           return
         }
         const sharedBrand = join(folderRoot, brandFile)
         if (existsSync(sharedBrand) && statSync(sharedBrand).isFile()) {
           const type = FOLDER_MOUNT_MIME['.png']
           res.setHeader('Content-Type', type)
           res.setHeader('Content-Length', statSync(sharedBrand).size)
           createReadStream(sharedBrand).pipe(res)
           return
         }
       }

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
// 302 to /cars/ and booted the root SPA. That SPA has no tenant - a request at the
// root resolves to no tenant, so every request it makes is refused with
// db_unavailable.
// Its login form is a dead end that is indistinguishable from a working one right up
// until a password is typed into it, which is worse than not offering the form.
//
// Production never had this bug, and the reason is worth keeping in mind before
// "fixing" it by pointing the dev server at the root app again: the rendered nginx
// config has no `location = /` at all, so / falls through to `root __WEBROOT__;` and
// index index.html, i.e. __WEBROOT__/index.html. deploy.sh rsyncs the build into
// $WEBROOT/dist/ and the code into $WEBROOT/api/, so no file is ever written at the
// webroot root and nginx answers 404. The dev server was the only place where /
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

// Name the tenant on every /<tenant>/api/ request, before Vite's proxy middleware runs.
//
// The shared api/ decides which database and which files folder a request belongs to
// from the tenant in the path (see the proxy entry in server.proxy). In production
// nginx passes that along as CARDS_TENANT, per location. The dev server has to do the
// same, and it used to do it in a proxyReq listener that never fired - rewrite()
// reassigns req.url to /api/<rest> before 'proxyReq' is emitted, so the listener read
// a path whose tenant prefix was already gone.
//
// The failure was silent in a way that made it look like a data problem rather than a
// routing one. With no tenant, app_db_files_dir() returns null, so upload.php resolved
// the file against the webroot instead of the tenant folder: the upload succeeded and
// the file really was on disk in <tenant>/files/, but every subsequent read 404'd
// because it was looking in a folder where it had never been written. An <img src>
// cannot send headers, so display broke for exactly the files the app had just stored.
//
// Middleware, not a proxy event, because this is the only hook that runs while
// req.url still has the tenant in it. Plugin middleware registered here also runs
// before Vite's internal ones, including the proxy - the same ordering note as
// serveNothingAtRoot below.
const nameTenantApiRequests = {
  name: 'name-tenant-api-requests',
  configureServer(server) {
    server.middlewares.use((req, res, next) => {
      const pathname = (req.url || '').split('?')[0]
      const m = pathname.match(/^\/([^/]+)\/api\//)
      if (m) {
        // Overwrite rather than defer to an incoming header: the path is the dev
        // server's own answer to which tenant this is, matching what nginx does per
        // location in production. app_request_tenant() treats a set-but-invalid
        // CARDS_TENANT as final instead of falling back to a guess, so letting a
        // client's header override this would make the path meaningless.
        req.headers['x-cards-tenant'] = m[1]
        req.headers['cards-tenant'] = m[1]
      }
      next()
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
    serveFolderMounts,
    serveNothingAtRoot,
    nameTenantApiRequests,
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
      // A tenant's API: /<tenant>/api/... is executed by the ONE shared api/, reached
      // by the same path - no rewrite, because the /<tenant> prefix is what tells
      // app_db_name() which database to open. Rewriting it away would send every
      // tenant's request to the webroot with no tenant, which resolves to nothing and
      // is refused.
      //
      // A pattern rather than one entry per client, for the same reason the middleware
      // above resolves per request: a proxy table is built when the config loads, so a
      // client provisioned afterwards had no entry, and its api.php fell through to
      // Vite's static handling. The cost is that '^/[^/]+/api' also matches a path that
      // is not a tenant mount - /src/api/..., say - and that request is forwarded to the
      // PHP server, which has no such file and answers 404. The alternative is a list
      // that is wrong by construction after the next provisioning run.
      '^/([^/]+)/api/(.*)$': {
        target: apiUrl,
        changeOrigin: true,
        secure: isProduction,
        // The tenant is already on the request by the time this runs - see
        // nameTenantApiRequests above. This only strips the prefix so the shared api/
        // answers, which is why app_db_name() reads the tenant from a header and not
        // from the path.
        //
        // The tenant used to be set in a proxyReq listener on this entry, reading
        // req.url. That cannot work: rewrite() below runs first and is what reassigns
        // req.url to /api/<rest>, so by the time 'proxyReq' is emitted the /<tenant>/
        // prefix the match depended on is already gone. The listener therefore never
        // fired, and every request reached PHP with no tenant.
        rewrite: (path) => {
          const m = path.match(/^\/[^/]+\/api\/(.*)$/)
          return '/api/' + (m ? m[1] : '')
        },
      },
      // Deliberately no '^/[^/]+/files' entry. A tenant's uploads are served from the
      // filesystem by the middleware above, and proxying them to PHP instead would be
      // a second path to the same bytes with different rules - one that only works
      // locally and would not exist in production.
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
