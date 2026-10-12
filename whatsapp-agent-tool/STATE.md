# State files — schemas & creation rules

**When to read this file:** only when you are about to **create or update** an entry in
`ai_questions.json` / `unanswered_chats.json`, or when you have to post a question to the
`ai questions` group. Reading the pending report, triaging and replying do **not** need it
(AGENTS.md §14).

`whatsapp-tool pending` owns the mechanical parts: 60-day trim, archiving to `data/state-archive.json`,
`waiting_for_data` → `data_ready` promotion, group scan. Never do those by hand.

## 1. `ai_questions.json` — questions waiting on the boss/team

```
id            number   incremental, unique within this profile (see §3 for the group tag)
question      string   the exact question
askedAt       ISO UTC
answered      boolean
answeredAt    ISO UTC | null
answer        string | null   boss/team answer
answeredBy    "boss" | "group" | null
sentToGroup   boolean
sentToGroupAt ISO UTC | null
groupLastScanAt ISO UTC | null
profileId      number?        which WhatsApp account (profiles.json id)
context       { chatKey?, chatName?, lastMessage?, lastMessageType?, lastMessageAt?, needed?[], reason?, profileId? }
```

- `chatKey`: prefer the `+213…` number, else the exact `nameOrPhone`.
- **Never store a voice transcription**: if the last message was a voice note, set
  `lastMessage: "[voice message]"` or omit it.

Rules:
1. **Create** when a client question needs data that is missing/stale and you just replied
   `سأطلب من المدير وأعود إليك` + the AI disclosure line. `answered: false`.
2. **Dedup first** — scan unanswered questions for the same `chatKey` + same/covering `needed`;
   if one exists, reuse that `id` (merge `needed`). New `id` only for a genuinely different ask.
3. **Answer found** (boss or group): `answered=true`, `answeredAt=now(UTC)`, `answer`,
   `answeredBy`.
4. **Never delete** an unanswered question that is still linked from an active
   `unanswered_chats` entry (unless both are > 60 days old).
5. Incremental ids stay monotonic; `nextId` = max(id)+1.

## 2. `unanswered_chats.json` — chats waiting for data

```
chatKey        string (required, normalized as above)
chatName       string?
status         "waiting_for_data" | "data_ready" | "replied" | "closed" | "expired"
reason         string?
needed         string[]
lastMessage    string?          (voice → "[voice message]" or omit)
lastMessageType "text" | "voice" | null
lastMessageAt  ISO UTC | null
askedAt        ISO UTC
questionIds    number[]         links to ai_questions.json
dataReceivedAt ISO UTC | null
repliedAt      ISO UTC | null
answered       boolean
notes          string?
profileId      number?        which WhatsApp account (profiles.json id)
```

Rules:
1. **Create/update in place** for the `chatKey` when you send the holding reply
   (`waiting_for_data`). Update, never duplicate.
2. **Follow-up while waiting**: same request → do nothing (do not resend the holding reply).
   New missing item → merge into `needed` and append the new `questionId` (dedup).
3. Stays `waiting_for_data` until **every** `needed` item is covered by an answer/data.
4. Only ever act on `data_ready`. After answering → `replied`, then `closed`. Never touch
   `replied`/`closed`.
5. Trim > 60 days (`whatsapp-tool pending` does it).

## 3. Posting a question to the `ai questions` group

Exact group name: **`ai questions`**. Post when the entry is created (never re-post when
`sentToGroup == true`), then set `sentToGroup`/`sentToGroupAt`. One message:

```
[ID: {profile}#{id}]
AR: {question in clear Arabic}
هذا الرد مولد بالذكاء الاصطناعي

EN: {question in clear English}
This is an AI-generated message

ZH: {question in clear Chinese}
这是一条由AI生成的消息

Chat: {chatKey}
Needed: {comma separated needed fields}
```

- **The id tag is `[ID: {profile}#{id}]` — always include the profile.** Every account has its
  own `ai_questions.json` whose ids start at 1, and all accounts post to this SAME group, so a
  bare `[ID: n]` would collide (default `[ID: 1]` and `sofia [ID: 1]`). The profile prefix makes
  the key globally unique and lets the scan attribute each reply to the right account.
  - Use `toSendToGroup[].groupKey` from the `whatsapp-tool pending` report verbatim (it is already
    `{profile}#{id}`); the profile name is the `WA_PROFILE` value (default account =
    `whatsapp-profile`). Example: `[ID: sofia#5]`.
  - The scan only accepts tags whose profile matches the running account, so answers for other
    accounts are never mis-attributed.
  - Legacy bare `[ID: n]` messages are ignored by the scan (ambiguous); re-post as
    `[ID: {profile}#n]` if one is still open.
- Arabic part is clear Arabic (no derja); the AI line goes after each language block.
- The sidebar must be on the **All** filter or the group is hidden — `whatsapp-tool pending` restores
  it; if you open the group manually, switch back to All first.
- Scanning `[ID: {profile}#n]` replies is done by `whatsapp-tool pending`, never by hand.
