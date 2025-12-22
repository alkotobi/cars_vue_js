-- Migration: Add share_token to clients table for secure client details access
-- This makes client detail links non-guessable

-- Add share_token column to clients table
ALTER TABLE `clients` 
ADD COLUMN `share_token` VARCHAR(64) NULL UNIQUE AFTER `id`;

-- Generate unique tokens for existing clients
-- Using SHA2 to generate a 64-character unique token
UPDATE `clients` 
SET `share_token` = SUBSTRING(
  SHA2(CONCAT(id, name, id_no, NOW(), RAND()), 256), 
  1, 
  64
)
WHERE `share_token` IS NULL OR `share_token` = '';

-- Create index on share_token for faster lookups
CREATE INDEX `idx_share_token` ON `clients` (`share_token`);

-- Add a trigger to auto-generate token for new clients
DELIMITER $$

CREATE TRIGGER `generate_client_share_token` 
BEFORE INSERT ON `clients`
FOR EACH ROW
BEGIN
  IF NEW.share_token IS NULL OR NEW.share_token = '' THEN
    SET NEW.share_token = SUBSTRING(
      SHA2(CONCAT(NEW.id, NEW.name, COALESCE(NEW.id_no, ''), NOW(), RAND()), 256), 
      1, 
      64
    );
  END IF;
END$$

DELIMITER ;

