# Algerian Derja (الجزائرية) — living glossary

**This file grows every time the user explains a word.** Read it *before* translating or
summarizing any WhatsApp chat. It is the only source of truth for dialect meaning — never invent a
translation.

---

## 0. How to use this file (protocol for AI agents)

1. **Before** translating/summarizing a chat, scan it for words not listed here.
2. **Never guess.** If a word matters and you don't know it, keep a short "unknown words" list and
   ask the user at the end of your answer (or mid-task if the task depends on it).
3. **When the user explains a word**, add it to the tables below immediately:
   - Arabic spelling + transliteration you heard
   - English meaning
   - `confirmed` = user explained it · `inferred` = you deduced it, still unverified
   - add the date and the chat it came from
4. Re-read this file at the start of every WhatsApp task. The goal is that the list converges and
   we stop asking.
5. **To answer a client, use [`CANNED_REPLIES.md`](./CANNED_REPLIES.md)** — standard derja replies
   (inspection policy, price, shipping/customs, bulk, follow-ups). Every outgoing message ends with
   the AI disclosure line defined there.

---

## 1. Business context (read this to interpret chats correctly)

- **We are car dealers / exporters based in China.** Our counterparts on WhatsApp are **clients and
  partners in Algeria**.
- So chats are about: sourcing vehicles in China, their specs/condition/year, sea freight,
  customs, paperwork (papiers), Algerian registration, and delivery. Prices quoted in these chats
  are for **export from China to Algeria**, which is why they look high compared to local Algerian
  used-car prices.
- Both sides are car people, so slang is dense: model names are transliterated from Latin script
  (Corolla → كولراي, Toyota → طويوتا, Emigrand → ايميڨراند), gearbox/trim words come from French
  (manual → مانيال), and car-part words from English.
- **Years are quoted by plate year**: when we write a year (e.g. "تدخل 2027") the client reads the
  **last 2 digits on the matriculation plate**. Say the plate digits, not the model year.
- **"بالبطاقة الرمادية"** in a price request = the price that comes with a registrable car
  (registration card available) — usually the grey-import/customs-cleared number.
- **Lost papers**: client needs a police-stamped declaration of loss (التصريح بالضياع) — this comes up
  before we can re-issue documents for a vehicle.
- The user is the **dealer/seller** side in most chats; clients ask for a specific model, year and
  price, and often negotiate the customs/registration angle.

## 2. Currency convention (critical — easy to get wrong)

- Algerians usually talk in **centimes (سنتيم)**, not dinars: **1 DZD = 100 centimes**.
- **Rule: if a price is quoted without "da" / "دج" / "دينار", treat the number as CENTIMES.**
  - `248 مليون` = 248,000,000 centimes = **2.48 million DZD** (≈ €17k) — NOT 248 million dinars.
  - A bare `110` in the same sentence means the same unit (≈ 1.1 million DZD), not 110 DZD.
- When reporting a converted amount, **state the arithmetic** (`248M centimes = 2.48M DZD`) so the
  user can catch a mistake instantly.
- If an amount is ambiguous (no unit, unusual magnitude), say so and ask instead of converting.

## 3. Confirmed words (user explained these)

| Arabic | Transliteration | English | Notes |
|---|---|---|---|
| ايميڨراند | Emigrand | Emigrand (vehicle/model name) | Latin name is **Emigrand**; written ايميڨراند in derja. `ڨ` = "g" |
| واصلة لميناء الجزائر | wsala lmitaa3 al-jazaair | **shipping included** (up to Algiers port) | `واصلة` alone = only "arriving". It does **NOT** mean delivered to the client's door, and must never be translated as "the price reaches you" (`السعر يوصلك`). Buyer pays customs clearance, registration, insurance, and transport from the port to his city |
| طنوبيلات / طنوبيلة | Tunobila/Tnoubila | **cars** (generic slang for vehicles) | NOT specifically a Fiat Doblò — it is the general Algerian word for a car/vehicle |
| خرجت غير كيما من … | khrejt ghir kima men … | **just out from the factory** (brand new) | The trailing noun was rendered "الوزين" in our capture; spelling varies between senders — confirm next time |
| لاكوت | lakwet | **a place / district in Algiers** (user confirmed: a location in Algiers) | Used as an answer to "أعطيني موقع مكتب وين في جزائر" → "your office is in La Kouat?" |
| دزير | dziyer | **Algiers** (NOT "Algeria" in general) | e.g. "شكون في دزير يبيع Alipay" = who in Algiers sells Alipay |
| البطاقة الرمادية | el-bate9a er-rmidia | **the car registration card in Algeria** ("grey card") | Clients ask for prices "بالبطاقة الرمادية" = the price that lets them register the car |
| التصريح بالضياع | et-tasri7 bel-dhiya3 | **police declaration of a lost document** — when an official paper is lost you go to the police station and get a stamped declaration | Needed before re-issuing papers / for customs and paperwork |
| الطابلات | et-tablat | **touch screen** (in a car) | "ما تتحكمش في المبرد عن طريق الطابلات" = the cooling can't be controlled from the touch screen |
| مديم / MEDIUM | meydiyum | **MID** (the trim level) | Clients write "gac gs3 medium"; our model/catalog name is **MID** (GS3 MID). User confirmed 2026-10-17. `quote` on the client's word returns MISSING → map to the exact catalog name (`MID`) |
| مان / MAN | man | **manual** (gearbox) | The `MAN` suffix in our car names = manual transmission. User confirmed 2026-10-17. MAN ≠ AUTO (different cars, different costs) |
| كي راك | ki rak | **how are you** | "شيخ نور الدين كي راك" = how are you, Sheikh Nour Eddine |
| رواي i5 | Roewe i5 | **Roewe i5** (the model the client transliterated as "rouoi i5") | client also asked MG5, Corolla without turbo, Sanguai |

### Confirmed *business* facts (not words)

| Item | Confirmed meaning |
|---|---|
| **"تدخل 2027"** (our own quote: `246 مليون واصلة تدخل 2027`) | The **matriculation plate shows "27"** as the 2-digit year, i.e. we quote by *plate year*, not model year |
| **"عنددي 300"** (+213 661 87 93 85) | **300 million centimes = 3,000,000 DZD** — that client's budget |
| Bare `300` style budgets | Bare large numbers in these chats are **centimes** (see §2) |

## 4. Confirmed phrases & sentences seen in chats

| Arabic | English | From |
|---|---|---|
| شعال سومة الاخ | "she's as old as you, brother" (teasing the vehicle's year) | +213 796 30 29 86 |
| 248 مليون واصلة بلا جمركة | "2.48M DZD, shipping included up to the port (Algiers), customs duty not paid" | +213 796 30 29 86 |
| وحد 110 المصاريف الباقية | "so, 1.1M DZD for the remaining costs?" | +213 796 30 29 86 |
| منقدروش ديرو 2026 و تدخل ‑3 سنين | "can't we register it as 2026 and import it as 3 years older?" (paper year = new, to pay less customs) | +213 796 30 29 86 |
| ما نقدرش نجخلها انا ‑3 سنوات | "I can't make it 3 years older" | +213 796 30 29 86 |
| قصدي معندكش طنوبيلات تع ‑3 سنين حاليا | "I mean — do you have cars that are 3 years old in stock right now?" | +213 796 30 29 86 |
| كاين ايميڨراند | "do you have an Emigrand?" | +213 796 30 29 86 |
| واش تحتاج عندنا بزاف | "what do you need? we've got plenty" | +213 796 30 29 86 |
| ولا كولراي مانيال | "or a manual Corolla" | +213 796 30 29 86 |
| جملة ؟ | "in bulk / wholesale?" | +213 796 30 29 86 |
| لا حبة وحدة | "not a single one — just one unit" | +213 796 30 29 86 |
| راني رايح لبارك فيه الكورولا | "I'm going to the dealership (البارك فيه = the car park = the lot/showroom) for the Corolla" | شراد |
| نصورلك ونرسل لك | "I'll take photos and send them to you" | شراد |
| كيما تتمنى لنفسك ربي يعطيك ماتتمنى | "may God grant you what you wish for yourself" (a blessing; also a jab meaning "you only care about yourself") | شراد |
| ستسأل عني يوم القايمة … نضع فيك ثقة كبيرة | "you'll be asked about me on the Day of Resurrection — I place great trust in you" | شراد |
| نحتاج طويوتا مستعملة | "I need a used Toyota" | شراد |
| التيوتا تخدموها | "which Toyota do you handle/deal in?" | شراد |

## 5. Inferred words (AI-deduced, NOT yet confirmed — verify or correct)

| Arabic | Transliteration | Guess | Notes |
|---|---|---|---|
| Slm / السلام | salam | hello | |
| وعليكم السلام ورحمة الله وبركاته | wa 3alaykum es-salam | and peace be upon you too | |
| الاخ | el-akh | brother (respectful address) | |
| واش | wash | what | from "واش هي" |
| عندنا / معندكش | 3andna / ma3andkosh | we have / you don't have | negative prefixes: ما…ش |
| بزاف | bzzaf | a lot, plenty | |
| كاين | kayen | there is / we have | |
| ولا | wala | or | |
| كولراي | kollray | Corolla | Latin transliteration |
| طويوتا | touyouta | Toyota | spelling varies: طويوطا |
| مانيال | maniyel | manual gearbox | from French "manuel" |
| مليون | million | million | |
| واصلة | wsala | arriving | **alone it does NOT mean "shipping included"** — the full phrase is below |
| بلا جمركة | bla jamrika | without customs duty | grey import |
| يعني | ya3ni | i.e. / so | |
| المصاريف الباقية | el-mesharif el-baqiya | the remaining costs/fees | |
| تقريبا | ta9riban | approximately | |
| منقدروش | man9edrouch | we can't | |
| تدخّل (دخل) | tdekhol | to import | |
| جديدة | jdida | new, brand new | |
| ما نقدرش | ma n9edrouch | I can't | |
| حبة وحدة | habba we7da | a single unit | حبة = piece |
| لبارك فيه | el-barek fih | the lot / dealership (literally "the parking where it is") | slang |
| راك | rak | are you (masculine) | "راك مشرف" = are you an admin/moderator |
| مشرف | moshaffef | moderator / admin | of a group or Facebook page |
| خوي / خوك | khawi / khok | my brother | |
| يالي | yali | come on / tell me | |
| فوتلي | foutli | forward to me | |
| خاص | khas | private / special | "مساج خاص" = private message |
| مساج / ميساج | mesaj | message | |
| نور الدين | Nour Eddine | person name | |
| دزير | dziyer | ~~Algeria~~ → **CONFIRMED = Algiers** | see §3 |
| ابحث | ib7eth | I search | |
| يخصني شكون | yekhsni chkoun | who do I ask (for a recommendation) | |
| علبة السرعة / علبة يدوية | 3la9a s3a / 3la9a yadwiya | gearbox (يدوية = manual) | "العلبة" = the box |
| فول ابشن / FULL OPTION | full option | fully-loaded trim | |
| باتل / BATTLE | battle | trim name (keep Latin) | |
| اقل من 3 سنوات | a9al men 3 snin | less than 3 years old (year criterion that drives customs) | |
| كولراي / كولاري / كورولا | kollray | Corolla | spelling varies wildly |
| جيلي | geely | Geely (Coolray) | |
| ليفان | livan | Livan (X3 Pro, GX3 Pro) | |
| شحال / شحال عندكم | cha7al | how much / what's your price | |
| بدولار / بالدولار | bdolar | in dollars (the other currency, vs dinar/centimes) | |
| بدينار | bdinyar | in dinars | |
| نسلكك / نسلككم | nsallek | we send you (a quote) | |
| مع كل التكاليف | ma3a kol el-tkhalif | all-in cost | |
| ابتداء من | ibda2 min | starting from | |
| كاين بال | kayen bel | there is with (time) | |
| LCC / certificat de conformité | — | the Chinese vehicle conformity certificate | clients ask for the English copy — confirm the exact document name |

## 5b. Inferred phrases seen in chats (NOT yet confirmed)

Rows promoted to §3 are kept in [`ALGERIAN_DERJA_ARCHIVE.md`](./ALGERIAN_DERJA_ARCHIVE.md).

| Phrase (Arabic) | Guess | From |
|---|---|---|
| لاباس | all good? / how's it going (labas) | bou3elam ceramic |
| صاحبي | my friend | +86 180 2974 1592 |
| تحتاج | you need | +86 180 2974 1592 |
| ملون و بدقة جيدة | colored and good quality (about photo quality) | group "Hicham الطبيب …" |
| فيها خير | it's fine / all good | group "Hicham الطبيب …" |
| راكم بخير | may you be well | +213 555 87 63 70 |
| المجمولة | el-mejmoula | the assembled/complete vehicle (vs CKD?) — **ask the user** | used by *us* in +213 661 42 29 46 |
| `_3ans` | 3 ans | client shorthand: wants a car **3 years old / under 3 years** — **ask the user** | +213 552 64 02 15 |
| GAG GS3 medium | GAC GS3 Medium | GAC GS3, Medium trim (client typed "Gag") — spelling **ask the user** | +213 552 64 02 15 |
| مرحاب | merhaba | our name — clients address us "اخي مرحاب" (brother Merhaba) | +213 555 87 63 70 |
| وبارك الله فيك مسبقا | barakallah fik masbi9an | thanks in advance / may God bless you beforehand | +213 555 87 63 70 |
| هل يوجد حل؟ | hal yujad hal? | is there a solution? | +213 555 87 63 70 |
| بخصوص وثائق سيارة من الصين | bisakhkhas watha2iq siyara men essin | about car documents from China (paperwork request) | +213 555 87 63 70 |
| منعرف حتى واحد في الصين | ma3reft 7atta wa7ed fi essin | I don't know anyone in China | +213 555 87 63 70 |
| أول بأول | awwal bi-awwal | right away / as soon as there's news (keep me posted) | +213 552 64 02 15 |
| من ولاية جنوبية | men wilaya jnubiya | from a southern wilaya (they will travel) | +213 671 41 56 57 |


## 6. Open questions / to confirm next time

- **"Certificat de conformité" / LCC**: clients ask for the **English copy** of the Chinese
  conformity certificate. Confirm: original Chinese, notarized translation, or the manufacturer's
  English certificate?
- **`السنة 3-` / `-3 سنين` / `_3ans`**: the year-on-paper trick (register as new, import as 3
  years older) is refused by us ("ما نقدرش نجخلها") but requested often — clients asking for
  "أقل من 3 سنوات" want **in-stock cars older than 3 years**, not a paper trick.
- **Group "Coolray manual + battle"** quoted *us* a price (Coolray manual 8,100 USD + shipping
  from 1,535 USD) — confirm whether they are competitors, an intermediary, or a serious buyer.
- Still-unconfirmed single words live in §5/§5b with **ask the user** markers — confirm them the
  next time they come up.

## 7. Change log

Lives in [`ALGERIAN_DERJA_ARCHIVE.md`](./ALGERIAN_DERJA_ARCHIVE.md) — **append new lines there**
(history is not loaded during normal turns).
