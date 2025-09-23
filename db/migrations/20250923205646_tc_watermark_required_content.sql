-- Modify "trust_center_watermark_configs" table
ALTER TABLE "trust_center_watermark_configs" ADD CONSTRAINT "text_or_logo_id_not_null" CHECK ((text IS NOT NULL) OR (logo_id IS NOT NULL));
