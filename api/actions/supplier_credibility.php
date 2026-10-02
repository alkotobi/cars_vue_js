<?php
// Supplier credibility checks: run the model over a supplier's own fields and
// keep every run.
//
// The model gets the supplier's NAME and nothing else, and is asked two
// questions: is this supplier credible, and does it have any court cases.
//
// The second question is the reason the prompt is written the way it is. A model
// with no tools cannot search, so asking it to "do a full search" and report court
// cases does not produce a search - it produces a confident, plausible,
// fabricated lawsuit against a real company. On a screen a person uses to decide
// whether to send cars and money, that is the worst possible output.
//
// So the prompt keeps the two questions, tells the model to answer question 2
// honestly when it has nothing, and requires can_verify (always false) plus a
// *_basis field on each answer so the UI can show how much weight a line of text
// deserves. Recollection is allowed but must be labelled as recollection.
//
// That is the whole feature: a second opinion to argue with, not a fact. If real
// registry or litigation data is ever needed it has to come from a provider that
// sells it, wired in as its own action - not from a bigger prompt here.
//
// Depends on getConnection(), getDbConfig() and apiErrorDie() from api.php, and
// require_api_admin() plus ai_chat_json() from lib/.

require_once __DIR__ . '/../lib/auth.php';
require_once __DIR__ . '/../lib/ai_client.php';

/** How many red flags one check may carry, and how long each may be. */
const SUPPLIER_CREDIBILITY_MAX_FLAGS = 10;
const SUPPLIER_CREDIBILITY_MAX_SUMMARY = 2000;
const SUPPLIER_CREDIBILITY_FLAG_LENGTH = 200;

/** Score at or above this is the low-risk band; below the medium cut is high. */
const SUPPLIER_CREDIBILITY_RISK_CUTS = [70, 40];

/**
 * Whether this server offers the feature at all.
 *
 * The server-side half of the off switch, and deliberately separate from
 * CREDIBILITY_ENABLED in the frontend. That flag hides the UI, which is a
 * courtesy; this one refuses the API, which is the guarantee. A cached bundle, a
 * stale tab, or anything else that can still reach api.php must not be able to
 * spend money on a model the operator has switched off.
 *
 * Defaults to ON. A server opts out with 'credibility_enabled' => '0', and the
 * default is ON so that upgrading a server that is already working does not
 * switch the feature off behind the operator's back. The two switches are
 * independent on purpose: this one alone decides whether an API call is served.
 */
function supplier_credibility_enabled(): bool
{
    return app_setting_bool('credibility_enabled', 'CREDIBILITY_ENABLED', true);
}

/**
 * Refuse the request when the server has the feature switched off.
 *
 * Checked after the admin check on every entry point, so an unauthorised caller
 * gets 'not_admin' as before and learns nothing about how the server is
 * configured.
 */
function supplier_credibility_require_enabled(): void
{
    if (!supplier_credibility_enabled()) {
        apiErrorDie('credibility_disabled');
    }
}

/**
 * Run a new check for one supplier. Admin only: it spends money.
 *
 * @param array<string,mixed> $postData
 */
function handle_assess_supplier_credibility(array $postData): void
{
    $conn = supplier_credibility_connection();
    $user = require_api_admin($conn, $postData);
    supplier_credibility_require_enabled();

    $supplierId = (int) ($postData['supplier_id'] ?? 0);
    if ($supplierId <= 0) {
        apiErrorDie('supplier_not_found');
    }

    $supplier = supplier_credibility_load($conn, $supplierId);
    if (!$supplier) {
        apiErrorDie('supplier_not_found');
    }

    $lang = supplier_credibility_lang((string) ($postData['lang'] ?? ''));
    [$system, $prompt] = supplier_credibility_prompt($supplier, $lang);

    try {
        $raw = ai_chat_json($system, $prompt);
    } catch (AiException $e) {
        error_log('assess_supplier_credibility: ' . $e->errorCode . ' ' . $e->getMessage());
        // The provider's own message is never echoed to the client: it can carry
        // a key fragment, a model name or an internal URL. The code is stable
        // and the client translates it.
        apiErrorDie($e->errorCode);
    }

    $check = supplier_credibility_normalize($raw, ai_settings()['model'], $lang);
    $check['id'] = supplier_credibility_insert($conn, $supplierId, (int) $user['id'], $check);

    echo json_encode(['success' => true, 'check' => $check]);
    exit;
}

/**
 * Past checks for one supplier, newest first.
 *
 * @param array<string,mixed> $postData
 */
function handle_get_supplier_credibility_checks(array $postData): void
{
    $conn = supplier_credibility_connection();
    require_api_admin($conn, $postData);
    supplier_credibility_require_enabled();

    $supplierId = (int) ($postData['supplier_id'] ?? 0);
    if ($supplierId <= 0) {
        apiErrorDie('supplier_not_found');
    }

    $limit = max(1, min(50, (int) ($postData['limit'] ?? 20)));
    $stmt = $conn->prepare(
        'SELECT c.id, c.id_supplier, c.score, c.risk_level, c.summary, c.red_flags,
                c.confidence, c.model, c.lang, c.id_user, c.date_create, u.username AS username
         FROM supplier_credibility_checks c
         LEFT JOIN users u ON u.id = c.id_user
         WHERE c.id_supplier = ?
         ORDER BY c.date_create DESC, c.id DESC
         LIMIT ' . $limit
    );
    $stmt->execute([$supplierId]);

    $checks = array_map('supplier_credibility_row', $stmt->fetchAll(PDO::FETCH_ASSOC));

    echo json_encode(['success' => true, 'checks' => $checks]);
    exit;
}

/**
 * The latest check per supplier, for the score badge in the suppliers table.
 * One call for the whole list rather than one per row.
 *
 * @param array<string,mixed> $postData
 */
function handle_get_supplier_credibility_latest(array $postData): void
{
    $conn = supplier_credibility_connection();
    require_api_admin($conn, $postData);
    supplier_credibility_require_enabled();

    $stmt = $conn->query(
        'SELECT c.id_supplier, c.id, c.score, c.risk_level, c.summary, c.red_flags,
                c.confidence, c.model, c.lang, c.id_user, c.date_create
         FROM supplier_credibility_checks c
         JOIN (
             SELECT id_supplier, MAX(id) AS id
             FROM supplier_credibility_checks
             GROUP BY id_supplier
         ) newest ON newest.id = c.id'
    );

    $latest = [];
    foreach ($stmt->fetchAll(PDO::FETCH_ASSOC) as $row) {
        $latest[] = supplier_credibility_row($row);
    }

    echo json_encode(['success' => true, 'checks' => $latest]);
    exit;
}

/**
 * The supplier plus the only history context that is actually verifiable: how
 * much has been bought from them so far. Money already moved is evidence; a
 * note claiming a payment record is not.
 *
 * @return array<string,mixed>|null
 */
function supplier_credibility_load($conn, int $supplierId): ?array
{
    try {
        // PDO sends the statement to MySQL at prepare time, so a column that
        // does not exist throws here rather than at execute.
        $stmt = $conn->prepare(
            'SELECT s.id, s.name FROM suppliers s WHERE s.id = ? LIMIT 1'
        );
        $stmt->execute([$supplierId]);
    } catch (PDOException $e) {
        // A server deployed before its migrations were run fails here on the
        // missing column, and left alone that is a fatal error with a stack
        // trace. Naming the migration turns a deploy-order mistake into one
        // sentence an admin can act on.
        error_log('supplier_credibility_load: ' . $e->getMessage());
        apiErrorDie('db_schema_outdated');
    }

    $row = $stmt->fetch(PDO::FETCH_ASSOC);

    return $row ?: null;
}

/**
 * Build the two halves of the prompt: rigid instructions and the data.
 *
 * The system half stays in English because JSON-shape compliance follows it
 * better, and it names the output language explicitly - the answer comes back in
 * the reader's language while the contract stays predictable.
 *
 * @return array{0:string,1:string}
 */
function supplier_credibility_prompt(array $supplier, string $lang): array
{
    $language = SUPPLIER_CREDIBILITY_LANGUAGES[$lang] ?? 'English';

    $system = <<<PROMPT
You are asked about one company: a car-export trading supplier. Answer exactly two questions about it, in {$language}.

Question 1: is this supplier credible?
Question 2: does this supplier have any court cases, and if so what are they?

What you actually have is the company's NAME and nothing else. No address, no contact, no registration number, no documents, no references.

Be clear-eyed about that, because it decides how you are allowed to answer:

- You have no internet access, no company register, no court register, no
  sanctions list and no way to verify anything. Nothing you write may suggest
  that you looked anything up, and you must never say that you did a search.
- You may use the name itself. The country and place it names, the legal form
  in the name, whether it names a legal entity or only a trading label, whether
  the name is generic or oddly specific - that is real analysis and you should
  do it properly.
- You may also use what you happen to know about a well-known company from
  training. If you do, say plainly that it is your own recollection, that it may
  be out of date, and that you could not check it.
- You must not invent anything to fill the gap. No court, case number,
  judgment, lawsuit, fine, address, phone number, registration number, amount or
  date that you cannot attribute to your own recollection. Inventing a lawsuit
  against a real company causes real harm, so an honest "I don't know" is worth
  far more than a confident guess. "I have no way to check this company's court
  record, and I do not recall any" is a complete, acceptable answer to
  question 2. Prefer it every time you are unsure.

Answer with one JSON object and nothing else:
{
  "can_verify": false,
  "credible": "<your direct answer to question 1, in {$language}>",
  "credible_basis": "name_analysis" | "recollection" | "no_information",
  "courts": "<your direct answer to question 2, in {$language}>",
  "courts_basis": "recollection" | "no_information",
  "score": <integer 0-100>,
  "risk_level": "low" | "medium" | "high",
  "red_flags": ["<short phrase, in {$language}>"],
  "confidence": "low" | "medium" | "high"
}
can_verify is always false: you did not verify anything and cannot. credible_basis is "recollection" only if you really know this company from training, and "name_analysis" if you are reasoning from the name alone. courts_basis is "recollection" only if you can actually recall this company's court history; use "no_information" whenever you are guessing or have nothing. red_flags is a list of specific concerns, each under 15 words; use [] when there are none. Write every text field in {$language}.
PROMPT;

    $facts = ['supplier_name' => supplier_credibility_utf8((string) $supplier['name'])];

    $factsJson = json_encode($facts, JSON_PRETTY_PRINT | JSON_UNESCAPED_UNICODE);
    if ($factsJson === false) {
        // Never send the model an empty message: it would answer about nothing.
        error_log('supplier_credibility_prompt: could not encode the supplier name');
        $factsJson = json_encode(['error' => 'The supplier name could not be encoded.']);
    }

    return [$system, $factsJson];
}

/**
 * Coerce whatever the model returned into a check that is safe to store and
 * render. A missing or nonsense field is filled in or dropped here rather than
 * reaching the UI as null.
 *
 * @return array<string,mixed>
 */
function supplier_credibility_normalize(array $raw, string $model, string $lang): array
{
    $score = (int) round((float) ($raw['score'] ?? 0));
    $score = max(0, min(100, $score));

    $risk = strtolower(trim((string) ($raw['risk_level'] ?? '')));
    if (!in_array($risk, ['low', 'medium', 'high'], true)) {
        $risk = supplier_credibility_risk_from_score($score);
    }

    $confidence = strtolower(trim((string) ($raw['confidence'] ?? '')));
    if (!in_array($confidence, ['low', 'medium', 'high'], true)) {
        $confidence = 'low';
    }

    $credible = supplier_credibility_truncate((string) ($raw['credible'] ?? ''), SUPPLIER_CREDIBILITY_MAX_SUMMARY);
    if ($credible === '') {
        $credible = 'The model returned no answer.';
    }

    $courts = supplier_credibility_truncate((string) ($raw['courts'] ?? ''), SUPPLIER_CREDIBILITY_MAX_SUMMARY);
    if ($courts === '') {
        $courts = 'The model returned no answer.';
    }

    // Recollection is the only basis that outranks no_information, and the model
    // is only trusted to claim it if it says so explicitly. Anything else falls
    // back to saying it does not know, so a default can never read as knowledge.
    $credibleBasis = strtolower(trim((string) ($raw['credible_basis'] ?? '')));
    if (!in_array($credibleBasis, ['name_analysis', 'recollection'], true)) {
        $credibleBasis = 'no_information';
    }

    $courtsBasis = strtolower(trim((string) ($raw['courts_basis'] ?? '')));
    if ($courtsBasis !== 'recollection') {
        // The model is far more willing to claim court history than to admit it
        // has none, so this defaults to the honest answer.
        $courtsBasis = 'no_information';
    }

    $flags = [];
    foreach ((array) ($raw['red_flags'] ?? []) as $flag) {
        $flag = supplier_credibility_truncate((string) $flag, SUPPLIER_CREDIBILITY_FLAG_LENGTH);
        if ($flag === '') {
            continue;
        }
        $flags[] = $flag;
        if (count($flags) >= SUPPLIER_CREDIBILITY_MAX_FLAGS) {
            break;
        }
    }

    return [
        'score' => $score,
        'risk_level' => $risk,
        'credible' => $credible,
        'credible_basis' => $credibleBasis,
        'courts' => $courts,
        'courts_basis' => $courtsBasis,
        'red_flags' => $flags,
        'confidence' => $confidence,
        'model' => $model,
        'lang' => $lang,
    ];
}

/**
 * Drop bytes that are not valid UTF-8.
 *
 * A single stray byte makes json_encode() fail, and because the API encodes the
 * whole response in one call that becomes an empty body rather than one bad
 * field. The report is written in the reader's language, so this is a real
 * possibility, not a theoretical one.
 */
function supplier_credibility_utf8(string $text): string
{
    if ($text === '' || json_encode($text) !== false) {
        return $text;
    }

    if (function_exists('mb_convert_encoding')) {
        $converted = mb_convert_encoding($text, 'UTF-8', 'UTF-8');
    } elseif (function_exists('iconv')) {
        $converted = iconv('UTF-8', 'UTF-8//IGNORE', $text);
    } else {
        $converted = false;
    }

    return $converted === false ? '' : $converted;
}

/**
 * Valid UTF-8, at most $limit characters, never cut mid-character.
 *
 * Characters, not bytes: an Arabic or Chinese summary is measured in
 * characters, and a byte-wise cut would leave half a character behind.
 */
function supplier_credibility_truncate(string $text, int $limit): string
{
    $text = trim(supplier_credibility_utf8($text));
    if ($text === '') {
        return '';
    }

    if (function_exists('mb_substr')) {
        return mb_substr($text, 0, $limit);
    }

    $chars = preg_split('//u', $text, -1, PREG_SPLIT_NO_EMPTY);

    return $chars === false ? substr($text, 0, $limit) : implode('', array_slice($chars, 0, $limit));
}

function supplier_credibility_risk_from_score(int $score): string
{
    [$lowCut, $mediumCut] = SUPPLIER_CREDIBILITY_RISK_CUTS;

    return $score >= $lowCut ? 'low' : ($score >= $mediumCut ? 'medium' : 'high');
}

/** @return int the new row id */
function supplier_credibility_insert($conn, int $supplierId, int $userId, array $check): int
{
    // The column is CHECK (JSON_VALID(...)), and supplier_credibility_truncate()
    // has already removed anything un-encodable - this is the belt to that
    // braces, so a bad row never becomes a failed INSERT.
    $flagsJson = json_encode(array_values($check['red_flags']));
    if ($flagsJson === false) {
        $flagsJson = '[]';
    }

    $stmt = $conn->prepare(
        'INSERT INTO supplier_credibility_checks
            (id_supplier, score, risk_level, summary, court_records, courts_basis, red_flags, confidence, model, lang, id_user, date_create)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, UTC_TIMESTAMP())'
    );
    $stmt->execute([
        $supplierId,
        $check['score'],
        $check['risk_level'],
        $check['credible'],
        $check['courts'],
        $check['courts_basis'],
        $flagsJson,
        $check['confidence'],
        $check['model'],
        $check['lang'],
        $userId,
    ]);

    return (int) $conn->lastInsertId();
}

/** A stored row, ready for the client. */
function supplier_credibility_row(array $row): array
{
    $flags = json_decode((string) ($row['red_flags'] ?? '[]'), true);

    return [
        'id' => (int) $row['id'],
        'supplier_id' => (int) $row['id_supplier'],
        'score' => (int) $row['score'],
        'risk_level' => (string) $row['risk_level'],
        'summary' => (string) $row['summary'],
        'court_records' => (string) ($row['court_records'] ?? ''),
        'courts_basis' => (string) ($row['courts_basis'] ?? 'no_information'),
        'red_flags' => is_array($flags) ? $flags : [],
        'confidence' => (string) $row['confidence'],
        'model' => (string) $row['model'],
        'lang' => (string) $row['lang'],
        'username' => isset($row['username']) ? (string) $row['username'] : null,
        'checked_at' => (string) $row['date_create'],
    ];
}

/** Report languages the prompt can ask for. */
const SUPPLIER_CREDIBILITY_LANGUAGES = [
    'en' => 'English',
    'ar' => 'Arabic',
    'fr' => 'French',
    'zh' => 'Chinese',
];

function supplier_credibility_lang(string $lang): string
{
    $lang = strtolower(substr(trim($lang), 0, 2));

    return isset(SUPPLIER_CREDIBILITY_LANGUAGES[$lang]) ? $lang : 'en';
}

function supplier_credibility_connection(): PDO
{
    $conn = getConnection(getDbConfig());
    if (is_array($conn) && isset($conn['error'])) {
        error_log('supplier_credibility_connection: ' . $conn['error']);
        apiErrorDie('db_unavailable');
    }

    return $conn;
}
