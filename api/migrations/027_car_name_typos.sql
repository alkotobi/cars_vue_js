-- ============================================
-- Migration: fix two misspelled car names
-- ============================================
-- Two default car names are misspelled in the seed data. Both are confirmed
-- corrections from the operator:
--
--   'CHERRY TIGO 7'              -> 'CHERY TIGO 7'
--   'SELTOS LUXERY BLACK ROOF'   -> 'SELTOS LUXURY BLACK ROOF'
--
-- These are renames of cars_names rows, so the ids do not change. Everything that
-- refers to a car name refers to it by id - buy_details.id_car_name and
-- car_name_media.car_name_id - so no other table needs updating and no row is
-- orphaned. That is also why this is safe where renaming a brand was not: a
-- brand is joined on its name in several queries, but a car name is only ever
-- looked up by id.
--
-- cars_names.car_name is UNIQUE, so each rename is guarded on the corrected name
-- not already existing. Without the guard, a database that somehow already has
-- the correct spelling aborts with a duplicate-key error partway through and
-- leaves the second rename unapplied.
--
-- Idempotent, and order-independent with 025: that migration now matches both
-- spellings of these two names, so it produces the right brand whether it runs
-- before or after this one. See the CHERY note there for why a single-spelling
-- match would be a silent no-op instead.

-- 'CHERRY TIGO 7' -> 'CHERY TIGO 7'
UPDATE `cars_names`
SET `car_name` = 'CHERY TIGO 7'
WHERE `car_name` = 'CHERRY TIGO 7'
  AND NOT EXISTS (
    SELECT 1 FROM (SELECT `car_name` FROM `cars_names`) AS existing
    WHERE existing.`car_name` = 'CHERY TIGO 7'
  );

-- 'SELTOS LUXERY BLACK ROOF' -> 'SELTOS LUXURY BLACK ROOF'
UPDATE `cars_names`
SET `car_name` = 'SELTOS LUXURY BLACK ROOF'
WHERE `car_name` = 'SELTOS LUXERY BLACK ROOF'
  AND NOT EXISTS (
    SELECT 1 FROM (SELECT `car_name` FROM `cars_names`) AS existing
    WHERE existing.`car_name` = 'SELTOS LUXURY BLACK ROOF'
  );
