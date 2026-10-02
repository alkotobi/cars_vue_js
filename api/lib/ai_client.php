<?php
// Client for an OpenAI-compatible chat completions endpoint, JSON answers only.
//
// "OpenAI-compatible" covers OpenAI itself plus DeepSeek, OpenRouter, Ollama and
// most local gateways: POST {base_url}/chat/completions, Authorization: Bearer
// <key>, a messages array. A provider that needs a different wire format
// belongs in its own client, not as a branch in here.
//
// This file knows how to ask a model for JSON and how to read the answer back.
// It knows nothing about suppliers - prompts live with their feature.
//
// Requires curl and the JSON extension, both standard on the deployment target.
// Set the endpoint in api/config.local.php (ai_api_key, ai_base_url, ai_model)
// or the AI_API_KEY / AI_BASE_URL / AI_MODEL environment variables.

require_once __DIR__ . '/settings.php';

/** A failed AI call, carrying the code the client maps to a translated message. */
class AiException extends RuntimeException
{
    /**
     * @var string one of ai_not_configured, ai_timeout, ai_rate_limited,
     *      ai_auth_failed, ai_http_error, ai_empty_response, ai_bad_response
     */
    public string $errorCode;

    public function __construct(string $errorCode, string $message)
    {
        parent::__construct($message);
        $this->errorCode = $errorCode;
    }
}

/**
 * How many tokens to allow.
 *
 * Generous on purpose, and it is a ceiling rather than a target: reasoning
 * models (Qwen3 on OpenRouter, for one) spend hundreds of tokens thinking
 * before they write anything, and a budget that runs out mid-thought returns an
 * empty message that looks like a broken endpoint. Cost follows what the model
 * actually writes, not this number.
 */
function ai_default_max_tokens(): int
{
    return app_setting_int('ai_max_tokens', 'AI_MAX_TOKENS', 3000);
}

/**
 * @return array{api_key:string,base_url:string,model:string,timeout:int,connect_timeout:int}
 */
function ai_settings(): array
{
    return [
        'api_key' => app_setting('ai_api_key', 'AI_API_KEY'),
        'base_url' => rtrim(app_setting('ai_base_url', 'AI_BASE_URL', 'https://api.openai.com/v1'), '/'),
        'model' => app_setting('ai_model', 'AI_MODEL', 'gpt-4o-mini'),
        'timeout' => app_setting_int('ai_timeout', 'AI_TIMEOUT', 60),
        'connect_timeout' => app_setting_int('ai_connect_timeout', 'AI_CONNECT_TIMEOUT', 10),
    ];
}

/** False when no key is configured, so callers can say so instead of failing opaquely. */
function ai_is_configured(): bool
{
    return ai_settings()['api_key'] !== '';
}

/**
 * Ask the model for a JSON object and return it decoded.
 *
 * @param string $system instructions, including the required JSON shape
 * @param string $user the data to reason about
 * @param array{max_tokens?:int,temperature?:float} $options
 * @return array<string,mixed>
 * @throws AiException
 */
function ai_chat_json(string $system, string $user, array $options = []): array
{
    $settings = ai_settings();
    if ($settings['api_key'] === '') {
        throw new AiException(
            'ai_not_configured',
            'No AI key configured. Set ai_api_key in api/config.local.php or AI_API_KEY.'
        );
    }

    $payload = [
        'model' => $settings['model'],
        'messages' => [
            ['role' => 'system', 'content' => $system],
            ['role' => 'user', 'content' => $user],
        ],
        'temperature' => $options['temperature'] ?? 0.2,
        'max_tokens' => $options['max_tokens'] ?? ai_default_max_tokens(),
    ];

    // PHP's own execution limit is commonly 30s, and where it is measured as
    // wall time - the built-in server, CLI, Windows - it fires while this script
    // is blocked on the socket. That turns a slow model into a fatal error and a
    // blank page, losing the clean, translatable ai_timeout that the curl
    // timeout below would otherwise have produced. Raise the limit to cover one
    // call plus the rate-limit retry, so the curl timeout is what decides and
    // PHP's limit is only a backstop. (PHP-FPM on Linux already excludes
    // blocking socket time from this limit, so this is belt-and-braces there.)
    set_time_limit($settings['timeout'] * 2 + 30);

    // JSON mode is the reason this returns arrays instead of prose. Not every
    // "compatible" server implements it, so one rejection is retried without it
    // and the answer is parsed defensively instead.
    $jsonMode = $payload;
    $jsonMode['response_format'] = ['type' => 'json_object'];

    $send = static function () use ($settings, $jsonMode, $payload) {
        $response = ai_post($settings, $jsonMode);

        return ai_rejected_json_mode($response) ? ai_post($settings, $payload) : $response;
    };

    $response = $send();
    if ($response['status'] === 429) {
        // A rate-limited request is rejected, not billed, and the free models
        // sit behind an upstream pool that rate-limits constantly - so one short
        // wait is worth it. Once, never twice: a genuine limit has to surface
        // as an error rather than as a slow hang.
        usleep(1500000);
        $response = $send();
    }

    $status = $response['status'];
    // These three are worth telling apart, because each one has a different
    // fix: wait, change the key, or file a bug. A single "the endpoint
    // returned an error" would make all three look the same.
    if ($status === 429) {
        throw new AiException('ai_rate_limited', 'The AI endpoint is rate limiting this server (HTTP 429).');
    }
    if ($status === 401 || $status === 403) {
        throw new AiException('ai_auth_failed', 'The AI endpoint rejected the configured key (HTTP ' . $status . ').');
    }
    if ($status < 200 || $status >= 300) {
        throw new AiException('ai_http_error', 'The AI endpoint returned HTTP ' . $status . '.');
    }

    return ai_extract_json(ai_message_content($response['body']));
}

/**
 * True when the server refused the request specifically over response_format,
 * i.e. a 4xx whose message names it.
 *
 * A JSON-mode rejection is the one failure worth re-sending, because the
 * fallback is the same request minus a field the server does not understand.
 * Anything else (401, 429, 5xx) is left alone here: the caller retries 429 once
 * on its own terms, and resending the rest would only cost twice.
 *
 * @param array{status:int,body:array} $response
 */
function ai_rejected_json_mode(array $response): bool
{
    if ($response['status'] !== 400 && $response['status'] !== 422) {
        return false;
    }

    $message = (string) ($response['body']['error']['message'] ?? '');
    return stripos($message, 'response_format') !== false
        || stripos($message, 'json_object') !== false
        || stripos($message, 'json mode') !== false;
}

/**
 * One HTTP round trip. Transport failures throw; an HTTP error status comes
 * back as data so the caller can decide whether retrying is worth it.
 *
 * @return array{status:int,body:array<string,mixed>}
 */
function ai_post(array $settings, array $payload): array
{
    $ch = curl_init($settings['base_url'] . '/chat/completions');
    if ($ch === false) {
        throw new AiException('ai_http_error', 'Could not initialise a request to the AI endpoint.');
    }

    curl_setopt_array($ch, [
        CURLOPT_POST => true,
        CURLOPT_RETURNTRANSFER => true,
        CURLOPT_POSTFIELDS => json_encode($payload),
        CURLOPT_HTTPHEADER => [
            'Content-Type: application/json',
            'Accept: application/json',
            'Authorization: Bearer ' . $settings['api_key'],
        ],
        CURLOPT_TIMEOUT => $settings['timeout'],
        CURLOPT_CONNECTTIMEOUT => $settings['connect_timeout'],
        // TLS verification stays on. A gateway with a self-signed certificate
        // is a deployment mistake to fix on the server, not something to
        // work around by trusting whatever answers.
        CURLOPT_SSL_VERIFYPEER => true,
        CURLOPT_SSL_VERIFYHOST => 2,
    ]);

    $raw = curl_exec($ch);
    $status = (int) curl_getinfo($ch, CURLINFO_RESPONSE_CODE);
    $curlError = curl_error($ch);
    // No curl_close(): the deployment runs PHP 8.5, where it is deprecated and
    // its notice lands in the response body, breaking the JSON. The handle is
    // freed when $ch goes out of scope, which is what 8.0+ does anyway.
    unset($ch);

    if ($raw === false) {
        // curl_error() never carries the request headers, so the key cannot
        // reach the log or the response through this message.
        $isTimeout = stripos($curlError, 'timed out') !== false;
        throw new AiException(
            $isTimeout ? 'ai_timeout' : 'ai_http_error',
            'Could not reach the AI endpoint' . ($curlError !== '' ? ': ' . $curlError : '')
        );
    }

    $body = json_decode($raw, true);
    if (!is_array($body)) {
        throw new AiException('ai_bad_response', 'The AI endpoint did not return JSON (HTTP ' . $status . ').');
    }

    if ($status < 200 || $status >= 300) {
        error_log('ai_post: HTTP ' . $status . ' ' . substr((string) ($body['error']['message'] ?? $raw), 0, 300));
    }

    return ['status' => $status, 'body' => $body];
}

/** Pull the assistant's text out of an OpenAI-shaped response. */
function ai_message_content(array $body): string
{
    $content = $body['choices'][0]['message']['content'] ?? '';
    if (is_array($content)) {
        // Some gateways return content as a list of parts.
        $content = implode('', array_map(
            static fn($part) => is_array($part) ? (string) ($part['text'] ?? '') : (string) $part,
            $content
        ));
    }

    if (!is_string($content) || trim($content) === '') {
        // A reasoning model that spent the whole token budget thinking answers
        // with an empty message, so this is the symptom to name rather than
        // "the response could not be read".
        $reasoning = $body['choices'][0]['message']['reasoning'] ?? '';
        error_log('ai_message_content: empty content'
            . ($reasoning !== '' ? ' after ' . strlen((string) $reasoning) . ' characters of reasoning' : '')
            . ' (finish_reason: ' . ($body['choices'][0]['finish_reason'] ?? 'unknown') . ')');
        throw new AiException(
            'ai_empty_response',
            'The model used its whole reply budget without answering.'
        );
    }

    return $content;
}

/**
 * Decode the model's answer as a JSON object. JSON mode makes this the normal
 * path; the brace-slice fallback covers a server that ignored it and answered
 * inside a code fence or with a sentence around the object.
 *
 * @return array<string,mixed>
 */
function ai_extract_json(string $content): array
{
    $content = trim($content);
    // ```json ... ``` fences survive when JSON mode is unavailable.
    if (preg_match('/```(?:json)?\s*(.+?)\s*```/s', $content, $m) === 1) {
        $content = trim($m[1]);
    }

    $decoded = json_decode($content, true);
    if (is_array($decoded)) {
        return $decoded;
    }

    $start = strpos($content, '{');
    $end = strrpos($content, '}');
    if ($start !== false && $end !== false && $end > $start) {
        $decoded = json_decode(substr($content, $start, $end - $start + 1), true);
        if (is_array($decoded)) {
            return $decoded;
        }
    }

    throw new AiException('ai_bad_response', 'The AI response was not valid JSON.');
}
