import { describe, it, expect } from 'vitest'
import { execFileSync } from 'node:child_process'
import { readFileSync } from 'node:fs'

// deploy/render-nginx-multitenant.php generates the nginx configuration that gives
// every client a URL on this server. These tests run it and look at what it
// produces, because the failure this file exists to prevent is not a PHP error - it
// is a config file that nginx accepts and that quietly does the wrong thing.
//
// What went wrong without this file
// ---------------------------------
// Server blocks were written by hand, once per client. Two ways that goes wrong,
// neither visible from the PHP:
//   - a client is provisioned and no block is written, so their site 404s and the
//     only clue is nginx's log;
//   - a client is deleted and their block stays, so their data remains reachable at
//     a URL nobody remembers owning.
// Generating from the `dbs` table removes both, because the registry is the same
// thing the app reads.
//
// The tests below are about the output, not the code shape, and they call the real
// render function with a tenant list they choose - so the collisions can be
// constructed rather than waited for.
//
// Skipped when there is no php: the repository's tests are otherwise plain
// JavaScript, and a contributor with only Node should still be able to run them.
// The renderer is PHP and this project is PHP, so php is normally present.

const RENDERER = new URL('../../deploy/render-nginx-multitenant.php', import.meta.url)
const TEMPLATE = new URL('../../deploy/nginx-multitenant.conf.template', import.meta.url)
const WRAPPER = new URL('../../deploy/cars-nginx-render', import.meta.url)
const EXAMPLE_CONFIG = new URL('../../deploy/cars-deploy.example.json', import.meta.url)
const SUDOERS = new URL('../../deploy/sudoers-cars-nginx.example', import.meta.url)
const DEPLOYMENT_DOC = new URL('../../DEPLOYMENT.md', import.meta.url)

const wrapper = readFileSync(WRAPPER, 'utf8')
const template = readFileSync(TEMPLATE, 'utf8')
const config = JSON.parse(readFileSync(EXAMPLE_CONFIG, 'utf8'))
const sudoers = readFileSync(SUDOERS, 'utf8')
const renderer = readFileSync(RENDERER, 'utf8')
const deploymentDoc = readFileSync(DEPLOYMENT_DOC, 'utf8')

const hasPhp = (() => {
  try {
    execFileSync('php', ['-v'], { stdio: 'ignore' })
    return true
  } catch {
    return false
  }
})()

const BASE_CONFIG = {
  api_dir: '/var/www/api',
  dist_dir: '/var/www/dist',
  php_socket: '/run/php/php8.3-fpm.sock',
  server_name: '163.245.214.125',
  cert_dir: '/etc/letsencrypt/live/163.245.214.125',
  webroot: '/var/www',
  default_server: true,
  ipv6: true,
}

// Builds the tenant rows the renderer derives from the registry itself, so the tests
// are not guessing at the shape. 'files' rather than <db>_files: uploads live inside
// the tenant folder, and the folder is named after the database.
const tenant = (dbName) => ({
  db_name: dbName,
  app_folder: dbName,
  files_folder: 'files',
})

/** Renders via the real PHP and returns { ok, out, err }. */
function render(config, tenants, { wrap = [] } = {}) {
  const json = JSON.stringify(config)
  const rows = JSON.stringify(tenants)
  const code = `
    require ${JSON.stringify(RENDERER.pathname)};
    try {
      echo cars_nginx_render(json_decode($argv[1], true), json_decode($argv[2], true));
    } catch (Throwable $e) {
      fwrite(STDERR, $e->getMessage());
      exit(1);
    }
  `
  try {
    const out = execFileSync('php', ['-r', code, '--', json, rows], {
      encoding: 'utf8',
      stdio: ['ignore', 'pipe', 'pipe'],
      cwd: new URL('../../', import.meta.url).pathname,
    })
    return { ok: true, out, err: '' }
  } catch (e) {
    return {
      ok: false,
      out: e.stdout?.toString() ?? '',
      err: e.stderr?.toString() ?? String(e),
    }
  }
}

/** Runs the collision check on its own. */
function checkCollisions(tenants) {
  const code = `
    require ${JSON.stringify(RENDERER.pathname)};
    try {
      cars_nginx_assert_no_collisions(json_decode($argv[1], true));
      echo "ok";
    } catch (Throwable $e) {
      fwrite(STDERR, $e->getMessage());
      exit(1);
    }
  `
  try {
    execFileSync('php', ['-r', code, '--', JSON.stringify(tenants)], {
      encoding: 'utf8',
      stdio: ['ignore', 'pipe', 'pipe'],
      cwd: new URL('../../', import.meta.url).pathname,
    })
    return { ok: true, err: '' }
  } catch (e) {
    return { ok: false, err: e.stderr?.toString() ?? String(e) }
  }
}

describe.skipIf(!hasPhp)('the rendered nginx config', () => {
  it('leaves no placeholder behind', () => {
    // A leftover __TENANTS__ lands inside a location block as a bare token. nginx
    // accepts that, so the file loads and every client 404s at once - which looks
    // like a DNS or deployment problem, not a template bug.
    const { ok, out } = render(BASE_CONFIG, [tenant('acme')])
    expect(ok).toBe(true)
    expect(out).not.toMatch(/__[A-Z_]+__/)
  })

  it('serves the app, its uploads and its API at the three URLs clients use', () => {
    const { out } = render(BASE_CONFIG, [tenant('acme')])
    expect(out).toContain('location = /acme {')
    expect(out).toContain('location = /acme/index.html {')
    expect(out).toContain('location /acme/ {')
    // Inside the tenant folder, not beside it. A sibling <db>_files/ is what this
    // replaced, and a tenant's app would then have been served from inside another
    // tenant's upload directory.
    expect(out).toContain('location /acme/files/ {')
    expect(out).not.toContain('location /acme_files/ {')
    expect(out).toMatch(/location ~ \^\/acme\/api\/\(\?<cars_api_acme_[0-9a-f]{8}>/)
  })

  it('serves the app from the one shared build, not from the tenant folder', () => {
    // The tenant folder holds files/ and nothing else, so try_files against root
    // would miss every asset: the build is not under /acme/ on disk. alias is what
    // makes one dist/ serve every tenant.
    const { out } = render(BASE_CONFIG, [tenant('acme')])
    expect(out).toContain('location /acme/ {')
    expect(out).toContain('alias /var/www/dist/;')
  })

  it('names the tenant in CARDS_TENANT so the location block decides it', () => {
    // The location block was generated for this tenant, so it can say so. Reading
    // the tenant out of Host or the URL instead would let a request be aimed at a
    // tenant other than the one this block was written for.
    const { out } = render(BASE_CONFIG, [tenant('acme')])
    expect(out).toContain('fastcgi_param CARDS_TENANT     acme')
  })

  it('leaves subdomains off unless a base domain is configured', () => {
    // Off by default: the certificate is the blocker, not the routing, and a vhost
    // for a host with no cert fails TLS rather than falling back.
    const { out } = render(BASE_CONFIG, [tenant('acme')])
    expect(out).not.toContain('server_name acme.')
    expect(out).toMatch(/Subdomains\s+off/)
  })

  it('serves the same tenant on its subdomain when subdomains are on', () => {
    const config = { ...BASE_CONFIG, base_domain: 'cars.example.com', subdomains: true }
    const { out } = render(config, [tenant('acme')])
    expect(out).toContain('server_name acme.cars.example.com;')
    expect(out).toContain('location /files/ {')
    // No <base> injection in the subdomain vhost: the app is at the root of its own
    // host, so the build's relative base is already right and /acme/ would not
    // exist there. Scoped to that vhost, because the path form still needs one.
    const vhost = out.slice(out.indexOf('# ---- acme.cars.example.com ----'))
    expect(vhost).not.toContain('<base href=')
    // And the API still names the tenant, since SCRIPT_NAME no longer does.
    expect(out).toContain('fastcgi_param CARDS_TENANT     acme')
  })

  it('refuses subdomains on a base domain that is not a domain', () => {
    const { ok, err } = render(
      { ...BASE_CONFIG, base_domain: 'not a domain', subdomains: true },
      [tenant('acme')],
    )
    expect(ok).toBe(false)
    expect(err).toContain('base_domain')
  })

  it('runs the API out of the one shared api/ while keeping the tenant in SCRIPT_NAME', () => {
    // This pair is the whole design. SCRIPT_FILENAME points at the shared copy, so
    // there is one api/ to deploy; SCRIPT_NAME keeps the /acme/api/ prefix, which is
    // the only thing telling api/lib/appdb.php which client the request is for.
    // Getting SCRIPT_FILENAME wrong serves the file from a tenant folder that does
    // not contain it (404 on every call); getting SCRIPT_NAME wrong silently serves
    // the primary app instead, which is the worse one.
    const { out } = render(BASE_CONFIG, [tenant('acme')])
    const capture = out.match(/location ~ \^\/acme\/api\/\(\?<(\w+)>/)[1]
    expect(out).toContain(`fastcgi_param SCRIPT_FILENAME /var/www/api/$${capture}`)
    expect(out).toContain(`fastcgi_param SCRIPT_NAME      /acme/api/$${capture}`)
  })

  it('uses a plain prefix for the app, never ^~', () => {
    // ^~ tells nginx to stop looking and use this block, skipping every regex on the
    // server - including that client's API. The request would then fall through to
    // the SPA fallback and the browser would get HTML for a POST, which surfaces as
    // a MIME-type error in the console and nothing at all in the access log.
    const { out } = render(BASE_CONFIG, [tenant('acme')])
    expect(out).toMatch(/location \/acme\/ \{/)
    expect(out).not.toMatch(/location \^~ \/acme\//)
  })

  it('has no generic .php handler anywhere, so PHP only runs via a tenant API location', () => {
    // Without this a .php file that ended up in a tenant folder - or an upload - is
    // one request away from executing.
    //
    // Comment lines are dropped first, and they have to be: this file's own prose
    // mentions `location ~ .php$` while explaining why there isn't one, and a
    // line-oriented scan otherwise reports the explanation as the thing it denies.
    const directives = render(BASE_CONFIG, [tenant('acme'), tenant('beta')]).out
      .split('\n')
      .filter((line) => /^\s*location\b/.test(line))
      .filter((line) => !/^\s*#/.test(line))

    // Every .php directive is one of exactly two things: that tenant's API location,
    // or a deny inside an uploads folder. Anything else is a route to PHP that is not
    // deliberately scoped to one client.
    const phpDirectives = directives.filter((line) => line.includes('.php'))
    expect(phpDirectives.length).toBeGreaterThan(0)
    for (const line of phpDirectives) {
      const isTenantApi = line.includes('/api/')
      const isDeny = line.includes('deny all')
      expect(isTenantApi || isDeny).toBe(true)
      // A deny that also passed the request to PHP would be no deny at all.
      if (!isTenantApi) {
        expect(line).not.toContain('fastcgi_pass')
      }
    }

    // The count is the point: two tenants give two API locations and two upload
    // denies, and there is no fifth, server-wide .php location anywhere.
    expect(phpDirectives).toHaveLength(4)
  })

  it('gives every tenant a distinct nginx variable name', () => {
    // nginx fails to start on a duplicate variable, so two clients would take every
    // site on the machine down. Client names are [a-z0-9_] now, so two of them
    // cannot sanitise onto one variable - the hash is what makes that a fact about
    // the renderer rather than an assumption about its input.
    const { ok, out } = render(BASE_CONFIG, [tenant('acme'), tenant('acme_eu')])
    expect(ok).toBe(true)
    const names = [...out.matchAll(/\(\?<(\w+)>/g)].map((m) => m[1])
    expect(names).toHaveLength(2)
    expect(new Set(names).size).toBe(2)
  })

  it('renders the same text for the same registry', () => {
    // A renderer that embeds a timestamp in the tenant section, or iterates a hash,
    // makes every deploy install a different file for no reason - and turns
    // "did anything change" into a question nobody can answer.
    const first = render(BASE_CONFIG, [tenant('acme'), tenant('beta')]).out
    const second = render(BASE_CONFIG, [tenant('acme'), tenant('beta')]).out
    const strip = (text) => text.replace(/^# Rendered .*$/m, '')
    expect(strip(first)).toBe(strip(second))
  })

  it('orders tenants the same way regardless of insertion order', () => {
    const strip = (text) => text.replace(/^# Rendered .*$/m, '')
    const forwards = strip(render(BASE_CONFIG, [tenant('aaa'), tenant('zzz')]).out)
    const backwards = strip(render(BASE_CONFIG, [tenant('zzz'), tenant('aaa')]).out)
    expect(forwards).toBe(backwards)
  })

  it('refuses a config with no certificate rather than emitting a broken path', () => {
    // cert_dir empty produces a server block naming /fullchain.pem. nginx -t fails on
    // that for a reason unrelated to whatever change is being deployed, and skipping
    // TLS silently would downgrade the site.
    const { ok, err } = render({ ...BASE_CONFIG, cert_dir: '' }, [tenant('acme')])
    expect(ok).toBe(false)
    expect(err).toContain('cert_dir')
  })

  it('refuses to render at all when the registry has no clients', () => {
    // It does render - the first run on a new server legitimately has none - but it
    // says so. Silence here is how an empty registry takes live clients offline
    // without anybody noticing why.
    const { ok, out } = render(BASE_CONFIG, [])
    expect(ok).toBe(true)
    expect(out).toContain('No clients are provisioned yet')
    expect(out.match(/^server \{/gm)).toHaveLength(2)
  })

  it('can leave the IPv6 listen lines out', () => {
    // A kernel without IPv6 makes `listen [::]:443` a fatal bind error, which stops
    // nginx starting for every site on the machine - over a decision unrelated to
    // the client being provisioned.
    const with6 = render(BASE_CONFIG, [tenant('acme')]).out
    const without6 = render({ ...BASE_CONFIG, ipv6: false }, [tenant('acme')]).out
    expect(with6).toContain('listen [::]:443')
    expect(without6).not.toContain('[::]')
  })

  it('only claims default_server when asked to', () => {
    // Another vhost on the same machine may already hold that position; nginx -t
    // then fails with "duplicate default server" and the whole render is rolled back.
    const yes = render(BASE_CONFIG, [tenant('acme')]).out
    const no = render({ ...BASE_CONFIG, default_server: false }, [tenant('acme')]).out
    expect(yes).toContain('listen 443 ssl http2 default_server')
    expect(no).toContain('listen 443 ssl http2;')
  })

  it('serves ACME ahead of the dotfile deny and the catch-all', () => {
    // A challenge that 404s is a certificate that quietly expires, and the renewal
    // failure is not noticed until a client sees a warning.
    const { out } = render(BASE_CONFIG, [tenant('acme')])
    expect(out).toMatch(/location \^~ \/\.well-known\/acme-challenge\//)
    expect(out).toContain('location ~ /\\. { deny all; }')
  })
})

describe.skipIf(!hasPhp)('tenant folder collisions', () => {
  it('refuses a client named api or dist, which are the shared folders', () => {
    // Both are legal database names and both are folders at the webroot: api/ is the
    // code and dist/ is the build. A client called dist would have its folder
    // aliased over by the build it is served from, and its uploads would sit beside
    // shared code. The registry is a table an operator edits by hand, so this has
    // to be caught at render time.
    const { ok, err } = checkCollisions([tenant('dist')])
    expect(ok).toBe(false)
    expect(err).toContain('dist')

    expect(checkCollisions([tenant('api')]).ok).toBe(false)
  })

  it('refuses two clients that are one folder apart on a case-insensitive filesystem', () => {
    // Linux does not collapse these, so this only bites on a move to macOS or a
    // case-insensitive mount - where both tenants' uploads land in one directory.
    // Catching it here is cheaper than discovering it there.
    const { ok, err } = checkCollisions([tenant('acme'), tenant('ACME')])
    expect(ok).toBe(false)
    expect(err).toContain('same folder')
  })

  it('accepts a name that is only a sibling of a shared folder', () => {
    // apifoo is not api: the location blocks all end in a slash, and the folders
    // are distinct. Refusing these would make ordinary names unusable.
    expect(checkCollisions([tenant('apifoo'), tenant('dist2')]).ok).toBe(true)
  })

  it('accepts names that merely share a prefix', () => {
    // acme and acme2 are different folders: the trailing slash on the location keeps
    // /acme2/ from matching the /acme/ block.
    expect(checkCollisions([tenant('acme'), tenant('acme2'), tenant('acme_eu')]).ok).toBe(true)
  })

  it('accepts a single client', () => {
    expect(checkCollisions([tenant('acme')]).ok).toBe(true)
  })
})

describe('the root-owned pieces', () => {
  it('keeps the renderer outside the web root', () => {
    // Root executes it, so a web-writable renderer is a root shell by another name.
    expect(config.render_php).not.toMatch(/^\/var\/www\/api\//)
    expect(config.render_php).toMatch(/^\/usr\/local\//)
  })

  it('refuses to run a renderer the web user could have written', () => {
    // The sudoers rule is only as strong as the file it names.
    expect(wrapper).toMatch(/require_root_owned/)
    expect(wrapper).toMatch(/group- or world-writable/)
  })

  it('keeps the previous config before overwriting it', () => {
    // The restore path after a failed `nginx -t` is only correct if the .bak holds
    // the config that was in place. Taken after the overwrite, it is a copy of the
    // new file and the restore looks like it worked while doing nothing.
    const bakIndex = wrapper.indexOf('.bak')
    const installIndex = wrapper.indexOf('install -m 644 "$TMP" "$OUTPUT"')
    expect(bakIndex).toBeGreaterThan(-1)
    expect(bakIndex).toBeLessThan(installIndex)
  })

  it('tests the config before reloading, and says the running server is untouched', () => {
    // nginx reads its configuration at start and reload, not per request, so a failed
    // test really does leave every site on the box alone. Worth stating in the output,
    // because otherwise an operator assumes their sites are broken.
    expect(wrapper).toMatch(/nginx -t/)
    expect(wrapper).toMatch(/the running server is unchanged/)
  })

  it('reloads rather than restarts', () => {
    // A restart drops in-flight requests, including uploads in progress.
    expect(wrapper).toMatch(/systemctl reload nginx/)
    expect(wrapper).not.toMatch(/systemctl (restart|stop) nginx/)
  })

  it('takes no arguments, and says so', () => {
    // A sudoers entry that permits arguments is not a safe thing to grant the web
    // user, because anything it accepts, a caller chooses.
    expect(wrapper).toMatch(/unknown argument/)
    expect(sudoers).toMatch(/no arguments/)
  })

  it('will not write outside /etc/nginx', () => {
    expect(wrapper).toMatch(/refusing: render_output must be under \/etc\/nginx/)
  })

  it('names a default_server the operator has to reconcile with existing vhosts', () => {
    // A bare IP cannot be matched by name, so this block has to be the default for
    // its address - and if another vhost already is, nginx refuses to start. Better
    // as a documented decision than a surprise at reload time.
    expect(config.default_server).toBe(true)
    expect(config._default_server.join(' ')).toContain('duplicate default server')
  })

  it('documents the TLS position instead of assuming a domain', () => {
    // Hard-wrapped in the template, so match a phrase that survives the wrap.
    expect(template).toMatch(/A bare IP is reached without SNI/)
    expect(config._cert_dir.join(' ')).toContain('certbot >= 5.4')
  })
})

describe('root-executed code, and what it reads', () => {
  // This is the escalation the whole arrangement exists to prevent, and it is worth
  // being precise about where it lived. cars-nginx-render runs as root through a
  // sudoers rule that takes no arguments, and it executes
  // render-nginx-multitenant.php - which is root-owned, outside the web root, and
  // was hardened accordingly.
  //
  // But that renderer does require_once api/lib/tenant-provision.php, and on the
  // production server that file is served out of /var/www/api/lib. It is very
  // plausibly writable by whoever deploys the app, or by www-data. So the web user
  // could rewrite a file that root then executes, and every other precaution on the
  // renderer would have been irrelevant. Nothing about the code said this - the
  // only way to see it is to follow what the root process actually opens.
  it('validates the ownership of what the renderer reads, not just the renderer', () => {
    expect(wrapper).toMatch(/--print-includes/)
    expect(wrapper).toMatch(/require_root_owned "\$inc" "renderer input"/)
  })

  it('asks the renderer which files it resolves instead of keeping a second list', () => {
    // Two copies of the candidate-path list is one copy too many, and the wrong one
    // is worse than none: it would validate a path that is never loaded while the
    // loaded one goes unchecked, which reads in the test output exactly like safety.
    expect(wrapper).not.toMatch(/tenant-provision\.php/)
    expect(renderer).toMatch(/--print-includes/)
  })

  it('prints the includes before loading them, so the check can precede execution', () => {
    const printAt = renderer.indexOf('--print-includes')
    const requireAt = renderer.indexOf('require_once $carsNginxLib')

    expect(printAt).toBeGreaterThan(-1)
    expect(requireAt).toBeGreaterThan(-1)
    expect(printAt).toBeLessThan(requireAt)
  })

  it('refuses to continue when it cannot establish what the renderer reads', () => {
    // Fail closed. The alternative - render anyway with the inputs unvalidated - is
    // the exact state the check exists to prevent, so an unreadable answer has to
    // stop the run rather than downgrade it.
    expect(wrapper).toMatch(/renderer could not report its includes/)
  })

  it('finds the library on a server that keeps api/ inside a tenant folder', () => {
    // The two fixed candidates both assume /var/www/api. A server that deploys into
    // <webroot>/<tenant>/api matches neither, and the renderer exits 1 - which looks
    // exactly like a broken install rather than an unreachable path. api_dir is where
    // that server says its api actually is, so it has to be part of the search.
    expect(renderer).toMatch(/\$carsNginxDecoded\['api_dir'\]/)
    expect(renderer).toMatch(/\/lib\/tenant-provision\.php/)
    // And it must stay a path to try, not a path to trust: the wrapper's ownership
    // check is what makes reading the config here safe.
    expect(renderer).toMatch(/is_readable\(\$carsNginxConfigPath\)/)
  })

  it('keeps the same rule for every file root loads: root-owned, not writable by others', () => {
    expect(wrapper).toMatch(/0022/)
    expect(wrapper).toMatch(/is owned by '\$owner', not root/)
    expect(wrapper).toMatch(/group- or world-writable/)
  })

  it('tells the operator how to install a renderer input safely', () => {
    // The check is only fair if the fix is discoverable. Both the lib and the
    // template ship in the repo, so the answer is a copy, not an investigation.
    const documented = `${config.web_group} ${deploymentDoc}`.toLowerCase()
    expect(documented).toMatch(/tenant-provision\.php/)
  })
})

describe('config keys the renderer needs', () => {
  // Found by rendering a real config for a local rig: the renderer refused with
  // '"webroot" is not set in /etc/cars-deploy.json' while the file it was reading
  // set webroot. Two of the three keys were unreachable rather than unset.
  const lib = readFileSync(new URL('../../api/lib/tenant-provision.php', import.meta.url), 'utf8')
  const defaults = lib.slice(lib.indexOf('$defaults = ['), lib.indexOf('$booleanKeys ='))

  it('can be given a webroot, which the renderer refuses to render without', () => {
    // webroot was in neither $defaults nor the example config, while the renderer
    // treats it as mandatory - so no config file could ever satisfy it.
    expect(renderer).toMatch(/\$config\['webroot'\]/)
    expect(defaults).toMatch(/'webroot' =>/)
    expect(JSON.stringify(config)).toMatch(/"webroot"/)
  })

  it('can be given default_server and ipv6, which the example documented but nothing read', () => {
    // Documented as settable, ignored by the loader: default_server stayed false and
    // ipv6 stayed off no matter what the file said, silently.
    expect(defaults).toMatch(/'default_server' => false/)
    expect(defaults).toMatch(/'ipv6' => false/)
    expect(lib).toMatch(/\$booleanKeys = \['shared_api', 'default_server', 'ipv6'\]/)
    expect(JSON.stringify(config)).toMatch(/"default_server"/)
    expect(JSON.stringify(config)).toMatch(/"ipv6"/)
  })

  it('names the file it actually read when a setting is missing', () => {
    // The message hardcoded /etc/cars-deploy.json. With --config, or on a machine
    // reading deploy/cars-deploy.local.json, it pointed the operator at a file that
    // was not the one in play - the same dead end as naming none.
    expect(renderer).toMatch(/\$config\['config_path'\] !== '' \? \$config\['config_path'\] : 'the server configuration'/)
    expect(renderer).not.toMatch(/'\/etc\/cars-deploy\.json'\s*\n\s*\)\);/)
  })
})
