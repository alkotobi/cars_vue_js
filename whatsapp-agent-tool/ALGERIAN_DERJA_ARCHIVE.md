# Archive — glossary history & superseded entries

Working history for [`ALGERIAN_DERJA.md`](./ALGERIAN_DERJA.md) and
[`CANNED_REPLIES.md`](./CANNED_REPLIES.md). **Not loaded during normal turns** — the live files
point here when something has been superseded.

**Append new change-log lines here**, not in the live files.

---

## 1. Algerian Derja — change log

| Date | Change |
|---|---|
| 2026-10-06 | File created. Confirmed: Emigrand=ايميڨراند, طنوبيلات=cars, خرجت غير…="just out from factory". Currency rule added (centimes, 1 DZD = 100 centimes). Business context added (China → Algeria car trade). ~40 inferred entries from 3 chats pending confirmation. |
| 2026-10-06 | **Correction from the user:** "shipping included" = **واصلة لميناء الجزائر** (arrives at Algiers port) — NOT "السعر يوصلك" and NOT door-to-door. `واصلة` on its own only means "arriving". Also fixed reply style: replies go out in warm Arabic with 1–3 emoji, never dry, never sounding like a refusal → see `CANNED_REPLIES.md`. |
| 2026-10-06 | Sweep of 50 unread chats added ~20 inferred words (دزير، شحال، نسلكك، علبة السرعة، فول ابشن، ابتداء من، التصريح بالضياع، الطابلات…) + 4 new open questions. All still **inferred** — awaiting the user's confirmation. |
| 2026-10-06 | **User confirmed 9 items, promoted to §3:** لاكوت = a place in Algiers · دزير = **Algiers** (not Algeria) · البطاقة الرمادية = the car registration card ("grey card") · التصريح بالضياع = police-stamped declaration of a lost document · الطابلات = touch screen · كي راك = how are you · رَواي i5 = Roewe i5 · **"تدخل 2027" = the plate shows "27"** (we quote by plate year) · **"عنددي 300" = 300M centimes = 3,000,000 DZD**. |
| 2026-10-06 | **Auto-reply round:** fetched the last 5 unread chats and replied to all 5 (documents-from-China, Coolray white/black + voice note, 3-year stock, Algiers address, voice-only chat). Added ~8 inferred words (مرحاب، أول بأول، من ولاية جنوبية، وبارك الله فيك مسبقا…) + 3 pending (المجمولة، `_3ans`، GAC GS3). New helper `src/autoreply.ts` (fetch/peek/send). |
| 2026-10-08 | Trim: §5b rows already promoted to §3, the resolved open questions, both change logs and this history moved out of the live files (token hygiene). Everything below is kept verbatim. |
| 2026-10-17 | **User confirmed (catalog vocabulary):** clients say **"MEDIUM"** for the trim our catalog calls **MID** (e.g. GAC GS3 MEDIUM → `GS3 MID`), and **`MAN` = manual** gearbox. Added to §3 and to `client-reply.md` §3.10 (name sanity). |

## 2. §5b rows already confirmed in §3 (kept for history)

| Phrase (Arabic) | Confirmed meaning | From |
|---|---|---|
| كي راك / كي راك خويا | how are you | +213 541 24 85 34 |
| تدخل 2027 | the plate shows "27" | +213 699 62 62 00 |
| التصريح بالضياع | police stamped declaration of lost document | group "Hicham الطبيب …" |
| الطابلات | touch screen | +213 555 87 63 70 |
| دزير | Algiers | +213 671 61 34 94 |
| رواي i5 | Roewe i5 | +213 664 88 42 31 |
| لاكوت | a place in Algiers | group "Coolray manual + battle" |
| البطاقة الرمادية | car registration card (grey card) | +213 672 26 33 57 |

## 3. Resolved / superseded open questions

- ~~The exact spelling behind "كيما من الوزين"~~ — user confirmed it means "just out from the
  factory"; seen once as `الوزين`. Ask again only if it reappears differently.
- ~~Budgets quoted bare ("عنددي 300")~~ → **resolved**: centimes. 300 = 3,000,000 DZD.

## 4. Canned replies — change log

| Date | Change |
|---|---|
| 2026-10-06 | Created. First version was derja-only and read as a refusal — rewritten in warm Arabic with light emoji; #1 reframed as a service promise. 12 templates. |
| 2026-10-06 | Style rules locked: Arabic script (not stiff MSA, not broken derja), 1–3 emoji, never sound like a refusal, always append the AI disclosure, placeholders listed. |
