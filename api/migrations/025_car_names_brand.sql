-- ============================================
-- Migration: attach the default car names to their brands
-- ============================================
-- cars_names.id_brand was NULL for every default name, so Car Models
-- (src/views/CarNamesView.vue) rendered a blank Brand for all of them, and
-- nothing in the app could group or filter models by marque.
--
-- The brand was worked out from the model itself, not from the string, because
-- the strings are unreliable: "TIGO 3" is a Chery Tiggo, "CHERY TIGGO 3X"
-- spells the marque wrong, and the whole VW group is prefixed "T-" without ever
-- saying VW. Matching on the name alone would have got most of these wrong.
--
-- Every statement joins on the brand NAME rather than a hardcoded id, so this is
-- safe on a server whose brand ids were never 1-12. Re-running changes nothing:
-- it is a plain UPDATE, not an INSERT.
--
-- Nothing in the mapping below is left as an unverified guess. The names that
-- started out as inferences - TACOUA, the three THARUs and DASHING PRO - have
-- since been confirmed by the operator, and the comments say which.

-- ---- The one brand that has to be created first ----
--
-- 'LIVAN AUTO' and 'LIVAN MAN' are LIVAN cars, but there is no LIVAN row in the
-- brands table, so the UPDATE below would join against nothing and leave both
-- names NULL - silently, with no error. Create the row first. brands.brand is
-- UNIQUE and id is AUTO_INCREMENT, so INSERT IGNORE makes this a no-op on any
-- server that already has LIVAN, and no id is hardcoded.
--
-- The brand row is added rather than assumed because the names cannot be left
-- NULL: Car Models then renders a blank Brand and the app cannot group by
-- marque. If LIVAN is ever renamed, this is the line to change.
INSERT IGNORE INTO `brands` (`brand`) VALUES ('LIVAN');

-- ---- Certain: the marque is stated in the name, or the model is unambiguous ----
UPDATE `cars_names` cn JOIN `brands` b ON b.brand = 'AUDI'
SET cn.id_brand = b.id WHERE cn.car_name = 'A3';

UPDATE `cars_names` cn JOIN `brands` b ON b.brand = 'CHANGAN'
SET cn.id_brand = b.id WHERE cn.car_name = 'CHANGAN CS75 PLUS';

-- Chery. Joins on IN ('CHERY','CHERRY') rather than a single value because
-- migration 026 renames the brand row from the misspelled 'CHERRY' to 'CHERY',
-- and 025 may run before or after it. A single-value join would match nothing on
-- a renamed database and leave all eight names NULL with no error - the same
-- silent failure documented on the KIA statement below.
-- 'CHERRY TIGO 7' is listed in both spellings because migration 027 renames it
-- to 'CHERY TIGO 7', and this migration may run before or after. Listing only the
-- corrected spelling would leave the row NULL on a database that has not been
-- renamed yet, with no error.
UPDATE `cars_names` cn JOIN `brands` b ON b.brand IN ('CHERY', 'CHERRY')
SET cn.id_brand = b.id WHERE cn.car_name IN (
  'CHERY TIGO 7', 'CHERRY TIGO 7', 'CHERY TIGGO 3X', 'TIGO 3',
  'COOLRAY FULL NOT SPORT', 'COOLRAY SPORT',
  'COOLRAY SUPER AUTO NO SUNROOF', 'COOLRAY SUPER AUTO WITH SUNROOF',
  'COOLRAY SUPER MANUAL'
);

UPDATE `cars_names` cn JOIN `brands` b ON b.brand = 'FREIGHT'
SET cn.id_brand = b.id WHERE cn.car_name = 'FREIGHT';

UPDATE `cars_names` cn JOIN `brands` b ON b.brand = 'MG'
SET cn.id_brand = b.id WHERE cn.car_name IN ('MG5 BASE AUTO', 'MG5 MAN');

UPDATE `cars_names` cn JOIN `brands` b ON b.brand = 'SKODA'
SET cn.id_brand = b.id WHERE cn.car_name = 'KAMIQ GT';

UPDATE `cars_names` cn JOIN `brands` b ON b.brand = 'KIA'
SET cn.id_brand = b.id WHERE cn.car_name IN (
  'K3', 'KX1 20251.4LCVT SUNROOFEDITION',
  -- 'SELTOS LUXERY BLACK ROOF' is the historical misspelling, corrected to
  -- LUXURY by migration 027. Both spellings are listed so this works either side
  -- of that rename: an equality test against a single spelling matches nothing on
  -- the other side of it, and fails silently. The first version of this migration
  -- used only the corrected spelling and matched zero rows here.
  'SELTOS LUXURY BLACK ROOF', 'SELTOS LUXERY BLACK ROOF', 'SONET BLACK ROOF'
);

-- The VW group: the "T-" prefix is VW's, never says VW, and the GOLFs say it
-- nowhere at all.
--
-- TACOUA and the three THARUs match no model I could identify. I had placed them
-- here on weak evidence - they cluster with the VW "T-" names and nothing else in
-- the list claims that prefix. The operator has since confirmed VW, so they are
-- no longer flagged as a guess.
UPDATE `cars_names` cn JOIN `brands` b ON b.brand = 'VW'
SET cn.id_brand = b.id WHERE cn.car_name IN (
  'GOLF 300TSI R-LINE', 'GOLF R-LINE (FULL OPTION)',
  'T-CROSS', 'T-ROC STARLIGHT', 'T-ROC WITH PACKAKE', 'TIGUAN L',
  'TACOUA', 'THARU BASIC', 'THARU MID', 'THARU TOP'
);

-- 'JETTA VS5' is NOT here. It looks like a VW model and was originally mapped to
-- VW, but migration 026 establishes that JETTA is a brand in its own right, so
-- it is mapped there instead - after that migration, since running this first
-- would put it under the old 'JETA' row.
UPDATE `cars_names` cn JOIN `brands` b ON b.brand IN ('JETTA', 'JETA')
SET cn.id_brand = b.id WHERE cn.car_name = 'JETTA VS5';

-- Confirmed by the operator.

-- 'DASHING PRO 1.6 DCT' was mapped to Geely as a guess: I could not identify the
-- model and picked the marque that looked closest by lineage. It is in fact a
-- JETOUR, so this replaces that guess rather than adding a new one.
UPDATE `cars_names` cn JOIN `brands` b ON b.brand = 'JETOUR'
SET cn.id_brand = b.id WHERE cn.car_name = 'DASHING PRO 1.6 DCT';

-- Emgrand is definitively a Geely model.
UPDATE `cars_names` cn JOIN `brands` b ON b.brand = 'GEELY'
SET cn.id_brand = b.id WHERE cn.car_name IN ('EMGRAND MAN', 'EMGRAND USED CAR');

-- LIVAN is stated in the name, and the brand row above makes this resolve.
UPDATE `cars_names` cn JOIN `brands` b ON b.brand = 'LIVAN'
SET cn.id_brand = b.id WHERE cn.car_name IN ('LIVAN AUTO', 'LIVAN MAN');
