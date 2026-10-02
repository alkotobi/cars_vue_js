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
    'db_name' => 'merhab_cars',

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
