<?php
// Copy this file to config.local.php and fill in the real values.
// config.local.php is git-ignored and must never be committed.
//
// Database settings are required. The ai_* settings are not: leave them empty
// and the supplier credibility button reports "AI is not configured on this
// server" instead of failing. The same keys can come from the AI_API_KEY,
// AI_BASE_URL and AI_MODEL environment variables.

return [
    'db_host' => '127.0.0.1',
    'db_user' => '',
    'db_pass' => '',
    // Which database to create. Required only by api/install.php - every other
    // request resolves its database from the tenant in the request, never from
    // here. Leave empty otherwise: naming a live database here is how an unresolved
    // request used to end up serving a real tenant's data.
    'db_name' => '',

    // The bare domain tenants are served on, for the subdomain form
    // (<tenant>.example.com). Empty means subdomain requests are not recognised at
    // all and only the path form works.
    //
    // This has to be set, and cannot be guessed. The request's Host header is the
    // only signal otherwise, and it is attacker-controlled: guessing the first label
    // as the tenant means a request to example.com resolves to tenant "example" and
    // connects to a database called example. Naming the domain this installation
    // owns turns "is this host mine?" into a comparison against a configured value,
    // and everything else is not a tenant. Empty is the safe answer: it fails the
    // comparison for every host, so subdomain routing is simply off.
    //
    // May also come from the CARS_BASE_DOMAIN environment variable.
    'base_domain' => '',

    // Required only to run api/install.php. There is deliberately no default: an
    // empty value makes the installer refuse, so forgetting to set it fails
    // closed. Pass it as the X-Install-Key header, then delete install.php.
    'install_key' => '',

    // Any OpenAI-compatible endpoint: OpenAI, DeepSeek, OpenRouter, Ollama, ...
    // The value is the API root, without /chat/completions, which is appended.
    'ai_base_url' => '',
    'ai_api_key' => '',
    // Any model the endpoint offers. Must be able to answer with JSON.
    'ai_model' => '',
    // Seconds to wait for the whole request, and for the connection alone.
    'ai_timeout' => 60,
    'ai_connect_timeout' => 10,

    // Server-side off switch for the supplier credibility feature. '0' makes every
    // credibility endpoint answer with credibility_disabled, so nothing can spend
    // money on the model above - not a cached browser bundle, not a direct call.
    // Defaults to '1' when unset or unrecognised, so upgrading a working server
    // never switches the feature off by surprise. This is independent of
    // CREDIBILITY_ENABLED in src/lib/featureFlags.js, which only hides the UI.
    'credibility_enabled' => '1',
];
