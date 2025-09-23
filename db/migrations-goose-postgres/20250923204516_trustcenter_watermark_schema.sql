-- +goose Up
-- create "trust_center_watermark_config_history" table
CREATE TABLE "trust_center_watermark_config_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "trust_center_id" character varying NULL, "logo_id" character varying NULL, "text" character varying NULL, "font_size" double precision NULL DEFAULT 48, "opacity" double precision NULL DEFAULT 0.3, "rotation" double precision NULL DEFAULT 45, "color" character varying NULL, "font" character varying NULL, PRIMARY KEY ("id"));
-- create index "trustcenterwatermarkconfighistory_history_time" to table: "trust_center_watermark_config_history"
CREATE INDEX "trustcenterwatermarkconfighistory_history_time" ON "trust_center_watermark_config_history" ("history_time");
-- create "trust_center_watermark_configs" table
CREATE TABLE "trust_center_watermark_configs" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "text" character varying NULL, "font_size" double precision NULL DEFAULT 48, "opacity" double precision NULL DEFAULT 0.3, "rotation" double precision NULL DEFAULT 45, "color" character varying NULL, "font" character varying NULL, "trust_center_id" character varying NULL, "logo_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "trust_center_watermark_configs_files_file" FOREIGN KEY ("logo_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "trust_center_watermark_configs_trust_centers_watermark_config" FOREIGN KEY ("trust_center_id") REFERENCES "trust_centers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "trust_center_watermark_configs_trust_center_id_key" to table: "trust_center_watermark_configs"
CREATE UNIQUE INDEX "trust_center_watermark_configs_trust_center_id_key" ON "trust_center_watermark_configs" ("trust_center_id");
-- create index "trustcenterwatermarkconfig_trust_center_id" to table: "trust_center_watermark_configs"
CREATE UNIQUE INDEX "trustcenterwatermarkconfig_trust_center_id" ON "trust_center_watermark_configs" ("trust_center_id") WHERE (deleted_at IS NULL);

-- +goose Down
-- reverse: create index "trustcenterwatermarkconfig_trust_center_id" to table: "trust_center_watermark_configs"
DROP INDEX "trustcenterwatermarkconfig_trust_center_id";
-- reverse: create index "trust_center_watermark_configs_trust_center_id_key" to table: "trust_center_watermark_configs"
DROP INDEX "trust_center_watermark_configs_trust_center_id_key";
-- reverse: create "trust_center_watermark_configs" table
DROP TABLE "trust_center_watermark_configs";
-- reverse: create index "trustcenterwatermarkconfighistory_history_time" to table: "trust_center_watermark_config_history"
DROP INDEX "trustcenterwatermarkconfighistory_history_time";
-- reverse: create "trust_center_watermark_config_history" table
DROP TABLE "trust_center_watermark_config_history";
