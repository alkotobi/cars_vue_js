<?php

declare(strict_types=1);

/**
 * Render the multi-tenant nginx configuration from the dbs registry.
 *
 *   php deploy/render-nginx-multitenant.php [--config /etc/cars-deploy.json] [--out FILE] [--check]
 *
 * Why this is a script rather than a hand-written file: the tenant list IS the
 * registry. A configuration that has to be edited by hand for every new client
 * is a configuration that will be wrong — either the client gets no server block
 * (their site 404s with no indication why) or a block is left behind after a
 * client is deleted (their data stays reachable at a URL nobody remembers).
 * Generating it from the same table the app reads makes those two states
 * impossible to reach.
 *
 * This file only produces text. Installing it, testing it and reloading nginx
 * are done by deploy/cars-nginx-render, which runs as root; keeping the two
 * apart is what makes the rendering testable without root and lets the failure
 * path (a bad config) be exercised deliberately.
 *
 * --check prints the configuration and exits 0 without writing anything, which
 * is how the operator reviews a change before it is installed.
 */

// Installed to /usr/local/lib/cars/, deliberately NOT inside the web root: this
// file is executed as root (through cars-nginx-render), so if the web server's
// user can write it, a file-write bug in the app becomes root code execution.
// cars-nginx-render checks the ownership before it runs it.
//
// The two candidates are the production location and the repository, so the same
// file runs from a checkout without a second copy drifting out of sync.
// Declared before the library block below because --print-includes reports it, and
// that has to happen before any of it is loaded.
const CARS_NGINX_TEMPLATE_PATH = __DIR__ . '/nginx-multitenant.conf.template';

if (!function_exists('tenant_assert_valid_db_name')) {
    $carsNginxCandidates = [
        __DIR__ . '/../api/lib/tenant-provision.php',
        '/var/www/api/lib/tenant-provision.php',
    ];

    // A server that keeps api/ inside a tenant folder matches neither path above, so the
    // configured api_dir is tried as well. cars-deploy.json is root-owned and not writable
    // by the web server's user, and cars-nginx-render still requires_root_owned whatever is
    // resolved here, so this adds a path to try rather than a way to choose one.
    $carsNginxConfigPath = '/etc/cars-deploy.json';
    if (is_readable($carsNginxConfigPath)) {
        $carsNginxDecoded = json_decode((string) file_get_contents($carsNginxConfigPath), true);
        if (is_array($carsNginxDecoded) && !empty($carsNginxDecoded['api_dir'])) {
            $carsNginxCandidates[] = rtrim((string) $carsNginxDecoded['api_dir'], '/')
                . '/lib/tenant-provision.php';
        }
    }

    $carsNginxLib = null;
    foreach ($carsNginxCandidates as $candidate) {
        if (is_file($candidate)) {
            $carsNginxLib = $candidate;
            break;
        }
    }

    if ($carsNginxLib === null) {
        fwrite(STDERR, "render-nginx: cannot find api/lib/tenant-provision.php, looked in:\n  "
            . implode("\n  ", $carsNginxCandidates) . "\n");
        exit(1);
    }

    // Checked before the require, not after: --print-includes exists so that
    // cars-nginx-render can verify this path's ownership BEFORE root loads code
    // from it. Printing it has to be possible without having already trusted it.
    if (PHP_SAPI === 'cli' && in_array('--print-includes', $argv ?? [], true)) {
        echo $carsNginxLib, "\n";
        echo CARS_NGINX_TEMPLATE_PATH, "\n";
        exit(0);
    }

    require_once $carsNginxLib;
}

const CARS_NGINX_TEMPLATE = CARS_NGINX_TEMPLATE_PATH;

// ---------------------------------------------------------------------------
// Reading the tenant list
// ---------------------------------------------------------------------------

/**
 * The tenants that should have a server block.
 *
 * Only rows that have been provisioned. A row an operator has added but not
 * provisioned has no folder and no database, and a block for it would hand out a
 * URL that 500s on the first request.
 *
 * @return list<array{db_name:string,app_folder:string,files_folder:string}>
 */
function cars_nginx_tenants(PDO $registry): array
{
    $stmt = $registry->query(
        'SELECT db_name FROM dbs WHERE is_created = 1 AND db_name IS NOT NULL AND db_name <> \'\' ORDER BY db_name'
    );

    $tenants = [];
    foreach ($stmt->fetchAll(PDO::FETCH_COLUMN) as $dbName) {
        $dbName = (string) $dbName;

        // Re-validated here even though the API validates on write. The registry is
        // a table an operator can edit with a SQL client, and its db_name goes into
        // an nginx config verbatim: a name containing a space or a brace would end
        // the token early and quietly remove every deny rule below it.
        try {
            tenant_assert_valid_db_name($dbName);
        } catch (TenantProvisionError $e) {
            throw new TenantProvisionError(
                'Refusing to render: the registry contains a database name that cannot be used in an nginx '
                . 'configuration (' . $e->getMessage() . '). Fix or remove that row first - rendering anyway would '
                . 'produce a config file that does not parse, or worse, one that parses with the wrong rules.'
            );
        }

        $tenants[] = [
            'db_name' => $dbName,
            // The tenant folder, named after the database. The build is NOT in it:
            // that is the shared __WEBROOT__/dist/, and nginx aliases it in.
            'app_folder' => $dbName,
            // uploads, inside that folder. Relative, because it is a path under the
            // tenant folder rather than a sibling of it.
            'files_folder' => tenant_files_dir_name($dbName),
        ];
    }

    return $tenants;
}

/**
 * Refuse a set of tenants that would collide in the URL space.
 *
 * ## What can still collide
 *
 * Only the tenant folders themselves. Uploads moved inside the tenant folder, so
 * there is no longer a `<client>_files` sibling for another client to be named
 * after, and the whole collision class disappeared with it.
 *
 * What remains is a client named `api` or `dist`: those are the shared code and the
 * shared build, at the same level, so a client called `dist` would have its folder
 * aliased over by the build it is served from. And two names differing only in case
 * collapse into one folder on a case-insensitive filesystem.
 *
 * @param list<array{db_name:string,app_folder:string,files_folder:string}> $tenants
 */
function cars_nginx_assert_no_collisions(array $tenants): void
{
    $seen = [];

    foreach ($tenants as $tenant) {
        $folder = (string) $tenant['app_folder'];

        foreach (['api', 'dist'] as $reserved) {
            if (strtolower($folder) === $reserved) {
                throw new TenantProvisionError(sprintf(
                    'Refusing to render: "%s" is a client, and %s is the shared %s/ directory at the same level. Its '
                    . 'location block would be unreachable, and its uploads folder would sit next to shared code. '
                    . 'Rename it in the registry.',
                    $folder,
                    $reserved,
                    $reserved === 'api' ? 'code' : 'build'
                ));
            }
        }

        // Case-insensitively, because a case-insensitive client or a future move to
        // one would collapse the two paths into the same directory - and both
        // tenants' uploads would end up in one folder.
        $key = strtolower($folder);
        if (isset($seen[$key])) {
            throw new TenantProvisionError(sprintf(
                'Refusing to render: %s and %s are the same folder on a case-insensitive filesystem, so they would '
                . 'share one uploads directory. Rename one of them in the registry.',
                $folder,
                $seen[$key]
            ));
        }

        $seen[$key] = $folder;
    }
}

// ---------------------------------------------------------------------------
// Rendering
// ---------------------------------------------------------------------------

/**
 * @param array<string,mixed> $config
 * @param list<array{db_name:string,app_folder:string,files_folder:string}> $tenants
 */
function cars_nginx_render(array $config, array $tenants): string
{
    $serverName = trim((string) ($config['server_name'] ?? ''));
    $webroot = rtrim(trim((string) ($config['webroot'] ?? '')), '/');
    $phpSocket = trim((string) ($config['php_socket'] ?? ''));
    $certDir = rtrim(trim((string) ($config['cert_dir'] ?? '')), '/');
    $apiDir = rtrim(trim((string) ($config['api_dir'] ?? '')), '/');
    // The shared build. Defaults to <webroot>/dist rather than being required,
    // because that is where the layout puts it and a config that spells it out
    // adds one more thing to keep in step with the directory that has to exist.
    $distDir = rtrim(trim((string) ($config['dist_dir'] ?? '')), '/');
    if ($distDir === '') {
        $distDir = $webroot === '' ? '/dist' : $webroot . '/dist';
    }
    $defaultServer = !empty($config['default_server']);

    // Subdomains are opt-in, and off unless base_domain is a real domain.
    //
    // Off by default because the certificate is the blocker, not the routing: a
    // wildcard cert covers *.example.com but not example.com, so enabling this
    // without one produces a vhost that fails TLS and a client who cannot log in
    // at all - worse than the path form, which works. base_domain is what makes
    // it a real option rather than a wildcard against any Host that arrives.
    $baseDomain = strtolower(rtrim(trim((string) ($config['base_domain'] ?? '')), '.'));
    $subdomains = !empty($config['subdomains']) && $baseDomain !== '';
    if ($subdomains && preg_match('/^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)+$/', $baseDomain) !== 1) {
        throw new TenantProvisionError(sprintf(
            'Refusing to render: base_domain "%s" is not a domain. Set it to the bare domain clients are served on '
            . '(no scheme, no path), or turn subdomains off.',
            $baseDomain
        ));
    }
    // Whether to bind [::] as well.
    //
    // Separate because a kernel without IPv6 makes `listen [::]:80` a fatal bind
    // error, and this file is installed by an unattended root command: the failure
    // mode is nginx refusing to start, i.e. every site on the box goes down at
    // once, over a decision that has nothing to do with the client being
    // provisioned. Default on, matching stock Ubuntu.
    $ipv6 = !array_key_exists('ipv6', $config) || !empty($config['ipv6']);
    $default = $defaultServer ? ' default_server' : '';

    // Refused rather than defaulted. Emitting a server block with a certificate
    // path that does not exist produces a config that fails `nginx -t` for a
    // reason that has nothing to do with the change being made, and emitting one
    // without TLS produces a site that quietly downgrades.
    foreach ([
        'server_name' => $serverName,
        'webroot' => $webroot,
        'php_socket' => $phpSocket,
        'cert_dir' => $certDir,
        'api_dir' => $apiDir,
    ] as $key => $value) {
        if ($value === '') {
            throw new TenantProvisionError(sprintf(
                'Refusing to render: "%s" is not set in %s. Every one of these has to be answered before a server '
                . 'block can be written - see deploy/cars-deploy.example.json.',
                $key,
                // config_path, not a literal /etc/cars-deploy.json: --config can point
                // somewhere else, and telling an operator to edit a file that is not
                // the one being read is the same dead end as naming no file at all.
                $config['config_path'] !== '' ? $config['config_path'] : 'the server configuration'
            ));
        }
    }

    // A server_name is pasted into the config verbatim, same as a folder name.
    foreach ([$serverName, $webroot, $certDir, $apiDir, $distDir, $phpSocket] as $value) {
        if (preg_match('/[\s;{}#\'"\\\\]/', $value) === 1) {
            throw new TenantProvisionError(
                'Refusing to render: one of server_name, webroot, cert_dir, api_dir or php_socket contains a '
                . "character that would end the nginx token early ('{$value}')."
            );
        }
    }

    if (!is_file(CARS_NGINX_TEMPLATE)) {
        throw new TenantProvisionError('template not found: ' . CARS_NGINX_TEMPLATE);
    }

    // Sorted here, not left to the caller's query. cars_nginx_tenants() does order by
    // db_name, but "the output does not depend on the input order" is a property
    // worth having at the layer that produces the text: otherwise a caller that
    // forgets the ORDER BY installs a byte-different file for an unchanged registry,
    // and "did anything actually change" stops being answerable by looking at the
    // file. Cheap, and it makes the render reproducible from the tenant list alone.
    usort($tenants, static fn (array $a, array $b): int => strcmp((string) $a['db_name'], (string) $b['db_name']));

    $template = (string) file_get_contents(CARS_NGINX_TEMPLATE);
    $tenantBlocks = cars_nginx_tenant_blocks($tenants, $apiDir, $distDir, $webroot, $phpSocket);

    $subdomainVhosts = '';
    $subdomainDoc = '';
    if ($subdomains) {
        $subdomainVhosts = cars_nginx_subdomain_vhosts($tenants, $apiDir, $distDir, $webroot, $phpSocket, $baseDomain, $certDir);
        $subdomainDoc = <<<DOC
#
#   https://<tenant>.{$baseDomain}/                the same app, the same tenant
#   https://<tenant>.{$baseDomain}/files/...       the same uploads
#   https://<tenant>.{$baseDomain}/api/*.php       the same API
#
# Both forms serve one tenant from one folder. The subdomain vhosts set
# CARDS_TENANT, which api/lib/appdb.php reads first - so a request on
# <tenant>.{$baseDomain} cannot be pointed at a different tenant by its URL, and
# the Host header is never used to decide.
DOC;
    }

    $rendered = strtr($template, [
        '__RENDERED_AT__' => gmdate('Y-m-d\TH:i:s\Z'),
        '__SERVER_NAME__' => $serverName,
        '__WEBROOT__' => $webroot,
        '__CERT_DIR__' => $certDir,
        '__API_DIR__' => $apiDir,
        '__DIST_DIR__' => $distDir,
        '__DEFAULT_SERVER__' => $default,
        // No trailing newline: the placeholder sits on its own line in the
        // template, which already supplies one.
        '__IPV6_LISTEN_80__' => $ipv6 ? '    listen [::]:80' . $default . ';' : '',
        '__IPV6_LISTEN_443__' => $ipv6 ? '    listen [::]:443 ssl http2' . $default . ';' : '',
        '__TENANT_COUNT__' => (string) count($tenants),
        '__TENANTS__' => $tenantBlocks,
        '__SUBDOMAIN_VHOSTS__' => $subdomainVhosts,
        '__SUBDOMAIN_DOC__' => $subdomainDoc,
        '__SUBDOMAIN_STATE__' => $subdomains
            ? 'on, ' . $baseDomain
            : 'off (set base_domain and "subdomains": true in the server config to enable)',
    ]);

    // A leftover placeholder means the template and this renderer have drifted
    // apart, and the result would contain a literal __TENANTS__ - which nginx
    // accepts as a bare token inside a location block and serves as a 404 for
    // every client at once.
    if (preg_match('/__[A-Z_]+__/', $rendered, $leftover) === 1) {
        throw new TenantProvisionError(
            'Refusing to render: unsubstituted placeholder ' . $leftover[0] . ' left in the output. '
            . 'deploy/nginx-multitenant.conf.template and this renderer have drifted apart.'
        );
    }

    return $rendered;
}

/**
 * The per-client section: app, uploads, and the tenant's API.
 *
 * @param list<array{db_name:string,app_folder:string,files_folder:string}> $tenants
 */
function cars_nginx_tenant_blocks(
    array $tenants,
    string $apiDir,
    string $distDir,
    string $webroot,
    string $phpSocket
): string {
    if ($tenants === []) {
        return "    # No clients are provisioned yet. Provisioning one adds its block here.\n";
    }

    $blocks = [];
    $captures = [];
    foreach ($tenants as $tenant) {
        $name = (string) $tenant['db_name'];
        $app = (string) $tenant['app_folder'];
        $files = (string) $tenant['files_folder'];

        // A named capture, because the same pattern is repeated once per client in
        // one file, and two locations sharing an unnamed group is a duplicate
        // variable - nginx refuses to start on it.
        //
        // The sanitised name alone is not enough: nginx variable names allow only
        // [A-Za-z0-9_], so any name carrying a character outside that set collapses
        // onto another name's variable - and nginx refuses to start on a duplicate
        // one, taking the whole machine down. Tenant names are now restricted to
        // [a-z0-9_] upstream, so this cannot happen through the app; a registry
        // edited by hand can still get here, and a hash suffix costs nothing. sha1
        // keeps it stable, so an unchanged registry still renders byte-identical text.
        $capture = 'cars_api_'
            . preg_replace('/[^A-Za-z0-9_]/', '_', $name)
            . '_' . substr(sha1($name), 0, 8);

        if (isset($captures[$capture])) {
            // Belt and braces: the hash makes this unreachable in practice, but
            // "nginx will not start" is a worse failure than "refused to render",
            // and this is the one check that turns it into the latter.
            throw new TenantProvisionError(sprintf(
                'Refusing to render: %s and %s would produce the same nginx variable name (%s).',
                $captures[$capture],
                $name,
                $capture
            ));
        }
        $captures[$capture] = $name;

        $blocks[] = <<<NGINX
    # ---- {$name} ----
    location = /{$app} {
        return 301 /{$app}/;
    }

    # The SPA shell, with <base> injected so relative asset URLs resolve at this
    # depth. Exact match, so it is chosen ahead of the prefix block below and of
    # every regex on the server.
    location = /{$app}/index.html {
        alias {$distDir}/index.html;
        sub_filter_once on;
        sub_filter '<meta charset="UTF-8">' '<base href="/{$app}/"><meta charset="UTF-8">';
    }

    # The tenant's uploads. Declared BEFORE the app prefix because it is a longer,
    # more specific prefix, and nginx picks the longest matching one - so the two
    # cannot shadow each other. It resolves inside the tenant's own folder, which
    # is where the files are; nothing here reaches the shared build.
    location /{$app}/{$files}/ {
        location ~ \\.php\$ { deny all; }
        location ~ /\\.     { deny all; }
        add_header X-Content-Type-Options nosniff;
        # =404, not the SPA: a missing upload is a missing upload, and answering it
        # with index.html hands the browser HTML where it expects an image.
        try_files \$uri =404;
    }

    # Branding assets: prefer tenant's own files/ first, then fall back to shared dist/
    location ~ ^/{$app}/(logo|logo_default|letter_head|letter_head_default|gml2)\.png$ {
        alias {$webroot}/{$name}/{$files}/\$1.png;
        error_page 404 = @{$capture}_brand_shared;
    }
    location @{$capture}_brand_shared {
        alias {$distDir}/\$1.png;
    }

    # The app. alias, not try_files against root: the build is the ONE dist/ shared
    # by every tenant, so it is not under /{$app}/ on disk and the request path
    # would miss every file. try_files is kept, so a deep link falls back to the
    # shell.
    #
    # Plain prefix, NOT ^~.
    #
    # ^~ stops nginx from ever considering this client's API location, which is a
    # regex - so /{$app}/api/*.php would fall through to the SPA fallback and hand
    # the browser HTML for a POST. That fails as a MIME-type error in the console
    # with nothing in the access log to explain it.
    #
    # A directory is deliberately NOT in the try_files list: a route can collide with
    # a real directory, and matching the directory produced a 403 where the app
    # should have rendered.
    location /{$app}/ {
        alias {$distDir}/;
        try_files \$uri \$uri/ /{$app}/index.html;
    }

    # This client's API. The file comes from the one shared api/, while SCRIPT_NAME
    # keeps the /{$app}/api/ prefix, which is the only thing telling the app which
    # client the request is for.
    #
    # fastcgi.conf rather than snippets/fastcgi-php.conf: the snippet carries
    # `try_files \$fastcgi_script_name =404`, which resolves against the tenant
    # folder and finds nothing, because the file only exists under the shared
    # api/. That 404s every request.
    #
    # There is deliberately no generic `location ~ \.php$` anywhere on this server.
    # This block is the only route to PHP, so a .php file that ends up in a tenant
    # folder or an uploads folder is unreachable as code, and one nested inside the
    # prefix above would be evaluated ahead of this location and deny the API.
    location ~ ^/{$app}/api/(?<{$capture}>[A-Za-z0-9_.-]+(?:/[A-Za-z0-9_.-]+)*\\.php)\$ {
        include fastcgi.conf;
        fastcgi_param SCRIPT_FILENAME {$apiDir}/\${$capture};
        fastcgi_param SCRIPT_NAME      /{$app}/api/\${$capture};
        fastcgi_param DOCUMENT_ROOT    {$apiDir};
        # The tenant, named by the location block rather than parsed out of the URL.
        # api/lib/appdb.php reads this first, before Host or SCRIPT_NAME, so a
        # request cannot be aimed at a different tenant than the one this block was
        # generated for.
        fastcgi_param CARDS_TENANT     {$app};
        fastcgi_pass unix:{$phpSocket};
        fastcgi_send_timeout 600;
        fastcgi_read_timeout 600;
    }

NGINX;
    }

    return implode('', $blocks);
}

/**
 * One vhost per tenant on <tenant>.<baseDomain>.
 *
 * The same tenant as the path form, served from the same folder, with the prefix
 * gone from the URL. Two consequences worth stating, because both are easy to get
 * backwards:
 *
 *   - CARDS_TENANT is SET here rather than left to the Host header. It is what makes
 *     a subdomain request resolve the same tenant as its path form, and it means a
 *     tenant cannot be reached through another tenant's vhost.
 *   - No <base> is injected. The app is at the root of this host, so its built-in
 *     relative base is already correct, and injecting /<tenant>/ here would send
 *     every asset request to a path that does not exist on this host.
 *
 * @param list<array{db_name:string,app_folder:string,files_folder:string}> $tenants
 */
function cars_nginx_subdomain_vhosts(
    array $tenants,
    string $apiDir,
    string $distDir,
    string $webroot,
    string $phpSocket,
    string $baseDomain,
    string $certDir
): string {
    if ($tenants === []) {
        return '';
    }

    $blocks = [];
    foreach ($tenants as $tenant) {
        $name = (string) $tenant['db_name'];
        $files = (string) $tenant['files_folder'];
        $host = $name . '.' . $baseDomain;
        $capture = 'cars_sd_api_' . $name;

        $blocks[] = <<<NGINX
# ---- {$host} ----
server {
    listen 80;
    server_name {$host};
    return 301 https://\$host\$request_uri;
}

server {
    listen 443 ssl http2;
    server_name {$host};

    # The same certificate directory as the parent domain. A wildcard certificate
    # for {$baseDomain} covers this host, which is why subdomains are opt-in: without
    # one, TLS fails here and the client cannot log in at all.
    ssl_certificate     {$certDir}/fullchain.pem;
    ssl_certificate_key {$certDir}/privkey.pem;

    root {$webroot};
    client_max_body_size 110M;

    location ^~ /.well-known/acme-challenge/ {
        root {$webroot};
        default_type "text/plain";
        allow all;
    }

    # The tenant's uploads, at /files/ on its own host.
    #
    # alias, not try_files against root: the uploads are in {$webroot}/{$name}/{$files}/,
    # and on this host they are not at /files/ - so without the alias, every one of
    # them 404s while the app keeps returning the URL it just wrote.
    location /{$files}/ {
        alias {$webroot}/{$name}/{$files}/;
        location ~ \\.php\$ { deny all; }
        location ~ /\\.     { deny all; }
        add_header X-Content-Type-Options nosniff;
        try_files \$uri =404;
    }

    # The API, from the one shared api/. SCRIPT_NAME has no prefix, so the tenant
    # comes from CARDS_TENANT - which this vhost sets, not from the request.
    location ~ ^/api/(?<{$capture}>[A-Za-z0-9_.-]+(?:/[A-Za-z0-9_.-]+)*\\.php)\$ {
        include fastcgi.conf;
        fastcgi_param SCRIPT_FILENAME {$apiDir}/\${$capture};
        fastcgi_param SCRIPT_NAME      /api/\${$capture};
        fastcgi_param DOCUMENT_ROOT    {$apiDir};
        fastcgi_param CARDS_TENANT     {$name};
        fastcgi_pass unix:{$phpSocket};
        fastcgi_send_timeout 600;
        fastcgi_read_timeout 600;
    }

    # Branding assets: prefer tenant's own files/ first, then fall back to shared dist/
    location ~ ^/(logo|logo_default|letter_head|letter_head_default|gml2)\.png$ {
        alias {$webroot}/{$name}/{$files}/\$1.png;
        error_page 404 = @{$capture}_brand_shared_sd;
    }
    location @{$capture}_brand_shared_sd {
        alias {$distDir}/\$1.png;
    }

    # The build. No <base> injection: the app is at the root of this host.
    location / {
        alias {$distDir}/;
        try_files \$uri \$uri/ /index.html;
    }

    # No <base> is injected, so index.html is served straight from the shared
    # build. Declared so the SPA fallback above has a real target to name.
    location = /index.html {
        alias {$distDir}/index.html;
    }

    location ~* \\.sql\$ { deny all; }
    location ~ /\\.      { deny all; }
}

NGINX;
    }

    return implode('', $blocks);
}

// ---------------------------------------------------------------------------
// CLI
// ---------------------------------------------------------------------------

if (PHP_SAPI === 'cli' && isset($argv[0]) && realpath($argv[0]) === realpath(__FILE__)) {
    $args = array_slice($argv, 1);
    $configPath = null;
    $outPath = null;

    foreach ($args as $i => $arg) {
        if ($arg === '--config' && isset($args[$i + 1])) {
            $configPath = $args[$i + 1];
        }
        if ($arg === '--out' && isset($args[$i + 1])) {
            $outPath = $args[$i + 1];
        }
    }

    // An explicit --config is always read from that path. It used to be routed
    // through tenant_server_config() whenever the file happened to be named
    // cars-deploy.json, which meant --config was silently ignored for exactly the
    // file name it is used with - and the renderer then read the real
    // /etc/cars-deploy.json instead, so a test or a dry run against a copy
    // rendered the live server's configuration.
    $explicit = $configPath !== null;
    $configPath ??= '/etc/cars-deploy.json';

    try {
        if (is_file($configPath)) {
            $config = cars_nginx_config_from_file($configPath);
        } elseif ($explicit) {
            throw new TenantProvisionError('--config was given but there is no such file: ' . $configPath);
        } else {
            // No file and no flag: fall back to the built-in defaults so the renderer
            // can still be run locally to see what it would refuse to do.
            $config = tenant_server_config();
        }

        $registryCreds = tenant_registry_credentials_from($config);
        // The database name is a separate argument to tenant_pdo(), not read from
        // the creds array - passing only the array connects with no catalog
        // selected and MySQL answers "No database selected".
        $registry = tenant_pdo($registryCreds, $registryCreds['dbname']);
        $tenants = cars_nginx_tenants($registry);
        cars_nginx_assert_no_collisions($tenants);

        $rendered = cars_nginx_render($config, $tenants);
    } catch (Throwable $e) {
        fwrite(STDERR, 'render-nginx: ' . $e->getMessage() . "\n");
        exit(1);
    }

    if ($outPath === null) {
        echo $rendered;
        exit(0);
    }

    if (file_put_contents($outPath, $rendered) === false) {
        fwrite(STDERR, "render-nginx: could not write {$outPath}\n");
        exit(1);
    }

    fwrite(STDERR, sprintf("render-nginx: wrote %s (%d tenant(s))\n", $outPath, substr_count($rendered, ' ---- ')));
    exit(0);
}

/**
 * Read a config file without going through tenant_server_config()'s static cache,
 * so a non-default path is actually honoured.
 *
 * @return array<string,mixed>
 */
function cars_nginx_config_from_file(string $path): array
{
    if (!is_file($path) || !is_readable($path)) {
        throw new TenantProvisionError('config not readable: ' . $path);
    }

    $decoded = json_decode((string) file_get_contents($path), true);
    if (!is_array($decoded)) {
        throw new TenantProvisionError('config is not valid JSON: ' . $path);
    }

    return $decoded;
}

/**
 * @param array<string,mixed> $config
 * @return array{host:string,user:string,pass:string,dbname:string}
 */
function tenant_registry_credentials_from(array $config): array
{
    $creds = $config['registry'] ?? [];

    if (!is_array($creds) || trim((string) ($creds['user'] ?? '')) === '') {
        throw new TenantProvisionError(
            'Refusing to render: no registry credentials in the config. The renderer runs as root through sudo, '
            . 'where the environment api/config.php reads from has been scrubbed, so it cannot fall back to it.'
        );
    }

    return [
        'host' => (string) ($creds['host'] ?? 'localhost'),
        'dbname' => (string) ($creds['dbname'] ?? 'merhab_databases'),
        'user' => (string) $creds['user'],
        'pass' => (string) ($creds['pass'] ?? ''),
    ];
}
