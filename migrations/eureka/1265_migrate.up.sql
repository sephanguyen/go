ALTER TABLE learning_material DROP CONSTRAINT IF EXISTS vendor_type_check;
ALTER TABLE learning_objectives DROP CONSTRAINT IF EXISTS vendor_type_check;

UPDATE learning_material SET vendor_type = 'LM_VENDOR_TYPE_MANABIE' WHERE 1 > 0;
UPDATE learning_objectives SET vendor_type = 'LM_VENDOR_TYPE_MANABIE' WHERE 1 > 0;

ALTER TABLE learning_material ALTER vendor_type SET DEFAULT 'LM_VENDOR_TYPE_MANABIE';
ALTER TABLE learning_objectives ALTER vendor_type SET DEFAULT 'LM_VENDOR_TYPE_MANABIE';

ALTER TABLE learning_material ADD CONSTRAINT vendor_type_check CHECK(vendor_type IN('LM_VENDOR_TYPE_MANABIE', 'LM_VENDOR_TYPE_LEARNOSITY'));
ALTER TABLE learning_objectives ADD CONSTRAINT vendor_type_check CHECK(vendor_type IN('LM_VENDOR_TYPE_MANABIE', 'LM_VENDOR_TYPE_LEARNOSITY'));
