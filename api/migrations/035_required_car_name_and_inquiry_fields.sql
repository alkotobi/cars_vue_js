-- ============================================
-- Migration: make key fields required
-- ============================================
-- Enforces choices that used to be silently defaulted, so a car model can no
-- longer be created without saying big/small, and an inquiry line can no longer
-- omit the color, quantity, target price or whether the car is used/new.
--
--   1. cars_names.is_big_car        -> NOT NULL, no default (must be supplied)
--   2. inquiry.date_inquiry         -> NOT NULL (the Go tool fills NOW() by default)
--   3. inquiry_details.id_color     -> NOT NULL
--   4. inquiry_details.QTY          -> NOT NULL
--   5. inquiry_details.target_price -> NOT NULL
--   6. inquiry_details.is_used_car  -> NOT NULL, no default (must be supplied)
--
-- Idempotent: each column is backfilled from NULL before the MODIFY, and a
-- MODIFY to the same definition is a no-op, so re-running is safe.
--
-- NOTE: dropping cars_names.is_big_car's default means every INSERT must now
-- pass it. The app forms (CarNameFormDialog.vue, CarStockForm.vue quick-add) and
-- the Go tool's `car-names add` were updated to require it in the same change.

UPDATE `cars_names` SET `is_big_car` = 0 WHERE `is_big_car` IS NULL;
ALTER TABLE `cars_names` MODIFY `is_big_car` tinyint(1) NOT NULL;

UPDATE `inquiry` SET `date_inquiry` = NOW() WHERE `date_inquiry` IS NULL;
ALTER TABLE `inquiry` MODIFY `date_inquiry` datetime NOT NULL;

UPDATE `inquiry_details` SET `id_color` = 0 WHERE `id_color` IS NULL;
UPDATE `inquiry_details` SET `QTY` = 0 WHERE `QTY` IS NULL;
UPDATE `inquiry_details` SET `target_price` = 0 WHERE `target_price` IS NULL;
UPDATE `inquiry_details` SET `is_used_car` = 0 WHERE `is_used_car` IS NULL;

ALTER TABLE `inquiry_details` MODIFY `id_color` int NOT NULL;
ALTER TABLE `inquiry_details` MODIFY `QTY` int NOT NULL;
ALTER TABLE `inquiry_details` MODIFY `target_price` decimal(10,2) NOT NULL;
ALTER TABLE `inquiry_details` MODIFY `is_used_car` tinyint(1) NOT NULL;
