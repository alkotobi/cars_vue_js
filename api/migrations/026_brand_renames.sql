-- ============================================
-- Migration: correct two misspelled brand names
-- ============================================
-- brands.brand held two typos, both confirmed by the operator:
--
--   'CHERRY' -> 'CHERY'   the actual marque
--   'JETA'   -> 'JETTA'   a brand in its own right, not a mangled VW model
--
-- These are renames, not new rows. The ids are deliberately left alone, so every
-- cars_names.id_brand, cars_stock.id_brand and anything else already pointing at
-- id 3 or id 6 keeps pointing at the same row and needs no update.
--
-- brands.brand is UNIQUE, so a plain UPDATE would abort with a duplicate-key
-- error on any database that already has the corrected spelling - which would
-- leave the second rename unapplied with no explanation. Each rename is
-- therefore guarded on the target name not already existing. If both spellings
-- are present that is a genuine data conflict needing a human, and skipping is
-- the safe outcome: it is visible, it fails loudly on the UNIQUE index later,
-- and it cannot silently merge two marques.
--
-- Idempotent: re-running is a no-op once the names are correct.

-- 'CHERRY' -> 'CHERY'
UPDATE `brands`
SET `brand` = 'CHERY'
WHERE `brand` = 'CHERRY'
  AND NOT EXISTS (
    SELECT 1 FROM (SELECT `brand` FROM `brands`) AS existing WHERE existing.`brand` = 'CHERY'
  );

-- 'JETA' -> 'JETTA'
UPDATE `brands`
SET `brand` = 'JETTA'
WHERE `brand` = 'JETA'
  AND NOT EXISTS (
    SELECT 1 FROM (SELECT `brand` FROM `brands`) AS existing WHERE existing.`brand` = 'JETTA'
  );

-- 'JETTA VS5' was mapped to VW by migration 025, on the reasoning that Jetta is a
-- VW model and the nearby 'JETA' brand row was a typo for it. That reasoning was
-- wrong in a way worth recording: 'JETA' was not a mangled 'Jetta', it was a
-- mangled 'JETTA', a separate brand. So the car moves from VW to JETTA.
--
-- Joining on both spellings so this is right whether or not the rename above has
-- been applied yet, and whether or not 025 has already run.
UPDATE `cars_names` cn JOIN `brands` b ON b.`brand` IN ('JETTA', 'JETA')
SET cn.id_brand = b.id WHERE cn.car_name = 'JETTA VS5';

-- 'CHERY TIGGO 3X' keeps its spelling as a car name. It is a car's name, not a
-- brand, and renaming one is the operator's call - the same reasoning that left
-- 'SELTOS LUXERY BLACK ROOF' alone.
