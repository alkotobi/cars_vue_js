-- ============================================
-- Migration: attach the default car names to their brands
-- ============================================
-- cars_names.id_brand was NULL for every default name, so Car Models
-- (src/views/CarModelsView.vue) rendered a blank Brand for all of them, and
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
-- Entries marked LOW below are inferences from the model's lineage, not from the
-- text. They are the ones worth eyeballing.

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

-- CHERRY is the spelling this app uses for Chery. The brand row itself is
-- probably the typo - see DEPLOYMENT.md - but it is the only Chery row there
-- is, and 'CHERY TIGGO 3X' has to land somewhere.
UPDATE `cars_names` cn JOIN `brands` b ON b.brand = 'CHERRY'
SET cn.id_brand = b.id WHERE cn.car_name IN (
  'CHERRY TIGO 7', 'CHERY TIGGO 3X', 'TIGO 3',
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
  -- Spelled 'LUXERY' in the data, not 'LUXURY'. Match the stored value exactly:
  -- an equality test against the corrected spelling matches nothing at all, and
  -- this statement fails silently. The typo itself is left alone - fixing it
  -- would rename a car, which is the operator's call, not a migration's.
  'SELTOS LUXERY BLACK ROOF', 'SONET BLACK ROOF'
);

-- The VW group: the "T-" prefix is VW's, never says VW, and the GOLFs say it
-- nowhere at all.
UPDATE `cars_names` cn JOIN `brands` b ON b.brand = 'VW'
SET cn.id_brand = b.id WHERE cn.car_name IN (
  'GOLF 300TSI R-LINE', 'GOLF R-LINE (FULL OPTION)', 'JETTA VS5',
  'T-CROSS', 'T-ROC STARLIGHT', 'T-ROC WITH PACKAKE', 'TIGUAN L'
);

-- ---- LOW: inferred from the model's lineage, worth confirming ----

-- LOW: could not place this name with confidence. Geely is the best fit among
-- the brands present.
UPDATE `cars_names` cn JOIN `brands` b ON b.brand = 'GEELY'
SET cn.id_brand = b.id WHERE cn.car_name = 'DASHING PRO 1.6 DCT';

-- Emgrand is definitively a Geely model.
UPDATE `cars_names` cn JOIN `brands` b ON b.brand = 'GEELY'
SET cn.id_brand = b.id WHERE cn.car_name IN ('EMGRAND MAN', 'EMGRAND USED CAR');

-- LIVAN is stated in the name, and the brand row above makes this resolve.
UPDATE `cars_names` cn JOIN `brands` b ON b.brand = 'LIVAN'
SET cn.id_brand = b.id WHERE cn.car_name IN ('LIVAN AUTO', 'LIVAN MAN');

-- LOW: 'TACOUA' matches no known model. It sits alphabetically inside the VW
-- T- family and nothing else in the list claims that prefix, so VW is the
-- likely owner, but this is a guess at a typo.
UPDATE `cars_names` cn JOIN `brands` b ON b.brand = 'VW'
SET cn.id_brand = b.id WHERE cn.car_name = 'TACOUA';

-- LOW: 'THARU' matches no known model either. Same reasoning as TACOUA: it
-- clusters with the VW T- names. If these are meant to be something else -
-- Karoq, Teramont - the brand needs correcting.
UPDATE `cars_names` cn JOIN `brands` b ON b.brand = 'VW'
SET cn.id_brand = b.id WHERE cn.car_name IN ('THARU BASIC', 'THARU MID', 'THARU TOP');
