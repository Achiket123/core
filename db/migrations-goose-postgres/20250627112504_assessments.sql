-- +goose Up
-- modify "groups" table
ALTER TABLE "groups" ADD COLUMN "assessment_blocked_groups" character varying NULL, ADD COLUMN "assessment_editors" character varying NULL, ADD COLUMN "assessment_viewers" character varying NULL;
-- create "assessments" table
CREATE TABLE "assessments" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "assessment_type" character varying NOT NULL DEFAULT 'INTERNAL', "template_id" character varying NOT NULL, "assessment_owner_id" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "assessment_id" to table: "assessments"
CREATE UNIQUE INDEX "assessment_id" ON "assessments" ("id");
-- create index "assessment_name_owner_id" to table: "assessments"
CREATE UNIQUE INDEX "assessment_name_owner_id" ON "assessments" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- create index "assessment_owner_id" to table: "assessments"
CREATE INDEX "assessment_owner_id" ON "assessments" ("owner_id") WHERE (deleted_at IS NULL);
-- create "assessment_history" table
CREATE TABLE "assessment_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "name" character varying NOT NULL, "assessment_type" character varying NOT NULL DEFAULT 'INTERNAL', "template_id" character varying NOT NULL, "assessment_owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "assessmenthistory_history_time" to table: "assessment_history"
CREATE INDEX "assessmenthistory_history_time" ON "assessment_history" ("history_time");
-- create "assessment_responses" table
CREATE TABLE "assessment_responses" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "status" character varying NOT NULL DEFAULT 'NOT_STARTED', "assigned_at" timestamptz NULL, "started_at" timestamptz NOT NULL, "completed_at" timestamptz NULL, "due_date" timestamptz NULL, "assessment_id" character varying NOT NULL, "user_id" character varying NOT NULL, "response_data_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "assessmentresponse_assessment_id_user_id" to table: "assessment_responses"
CREATE UNIQUE INDEX "assessmentresponse_assessment_id_user_id" ON "assessment_responses" ("assessment_id", "user_id") WHERE (deleted_at IS NULL);
-- create index "assessmentresponse_completed_at" to table: "assessment_responses"
CREATE INDEX "assessmentresponse_completed_at" ON "assessment_responses" ("completed_at");
-- create index "assessmentresponse_due_date" to table: "assessment_responses"
CREATE INDEX "assessmentresponse_due_date" ON "assessment_responses" ("due_date");
-- create index "assessmentresponse_id" to table: "assessment_responses"
CREATE UNIQUE INDEX "assessmentresponse_id" ON "assessment_responses" ("id");
-- create index "assessmentresponse_status" to table: "assessment_responses"
CREATE INDEX "assessmentresponse_status" ON "assessment_responses" ("status");
-- create "assessment_response_history" table
CREATE TABLE "assessment_response_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "assessment_id" character varying NOT NULL, "user_id" character varying NOT NULL, "status" character varying NOT NULL DEFAULT 'NOT_STARTED', "assigned_at" timestamptz NULL, "started_at" timestamptz NOT NULL, "completed_at" timestamptz NULL, "due_date" timestamptz NULL, "response_data_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "assessmentresponsehistory_history_time" to table: "assessment_response_history"
CREATE INDEX "assessmentresponsehistory_history_time" ON "assessment_response_history" ("history_time");
-- create "assessment_users" table
CREATE TABLE "assessment_users" ("assessment_id" character varying NOT NULL, "user_id" character varying NOT NULL, PRIMARY KEY ("assessment_id", "user_id"));
-- modify "groups" table
ALTER TABLE "groups" ADD CONSTRAINT "groups_assessments_blocked_groups" FOREIGN KEY ("assessment_blocked_groups") REFERENCES "assessments" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_assessments_editors" FOREIGN KEY ("assessment_editors") REFERENCES "assessments" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_assessments_viewers" FOREIGN KEY ("assessment_viewers") REFERENCES "assessments" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "assessments" table
ALTER TABLE "assessments" ADD CONSTRAINT "assessments_groups_assessment_owner" FOREIGN KEY ("assessment_owner_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assessments_organizations_assessments" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assessments_templates_template" FOREIGN KEY ("template_id") REFERENCES "templates" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "assessment_responses" table
ALTER TABLE "assessment_responses" ADD CONSTRAINT "assessment_responses_assessments_assessment_responses" FOREIGN KEY ("assessment_id") REFERENCES "assessments" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "assessment_responses_document_data_document" FOREIGN KEY ("response_data_id") REFERENCES "document_data" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assessment_responses_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "assessment_users" table
ALTER TABLE "assessment_users" ADD CONSTRAINT "assessment_users_assessment_id" FOREIGN KEY ("assessment_id") REFERENCES "assessments" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "assessment_users_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;

-- +goose Down
-- reverse: modify "assessment_users" table
ALTER TABLE "assessment_users" DROP CONSTRAINT "assessment_users_user_id", DROP CONSTRAINT "assessment_users_assessment_id";
-- reverse: modify "assessment_responses" table
ALTER TABLE "assessment_responses" DROP CONSTRAINT "assessment_responses_users_user", DROP CONSTRAINT "assessment_responses_document_data_document", DROP CONSTRAINT "assessment_responses_assessments_assessment_responses";
-- reverse: modify "assessments" table
ALTER TABLE "assessments" DROP CONSTRAINT "assessments_templates_template", DROP CONSTRAINT "assessments_organizations_assessments", DROP CONSTRAINT "assessments_groups_assessment_owner";
-- reverse: modify "groups" table
ALTER TABLE "groups" DROP CONSTRAINT "groups_assessments_viewers", DROP CONSTRAINT "groups_assessments_editors", DROP CONSTRAINT "groups_assessments_blocked_groups";
-- reverse: create "assessment_users" table
DROP TABLE "assessment_users";
-- reverse: create index "assessmentresponsehistory_history_time" to table: "assessment_response_history"
DROP INDEX "assessmentresponsehistory_history_time";
-- reverse: create "assessment_response_history" table
DROP TABLE "assessment_response_history";
-- reverse: create index "assessmentresponse_status" to table: "assessment_responses"
DROP INDEX "assessmentresponse_status";
-- reverse: create index "assessmentresponse_id" to table: "assessment_responses"
DROP INDEX "assessmentresponse_id";
-- reverse: create index "assessmentresponse_due_date" to table: "assessment_responses"
DROP INDEX "assessmentresponse_due_date";
-- reverse: create index "assessmentresponse_completed_at" to table: "assessment_responses"
DROP INDEX "assessmentresponse_completed_at";
-- reverse: create index "assessmentresponse_assessment_id_user_id" to table: "assessment_responses"
DROP INDEX "assessmentresponse_assessment_id_user_id";
-- reverse: create "assessment_responses" table
DROP TABLE "assessment_responses";
-- reverse: create index "assessmenthistory_history_time" to table: "assessment_history"
DROP INDEX "assessmenthistory_history_time";
-- reverse: create "assessment_history" table
DROP TABLE "assessment_history";
-- reverse: create index "assessment_owner_id" to table: "assessments"
DROP INDEX "assessment_owner_id";
-- reverse: create index "assessment_name_owner_id" to table: "assessments"
DROP INDEX "assessment_name_owner_id";
-- reverse: create index "assessment_id" to table: "assessments"
DROP INDEX "assessment_id";
-- reverse: create "assessments" table
DROP TABLE "assessments";
-- reverse: modify "groups" table
ALTER TABLE "groups" DROP COLUMN "assessment_viewers", DROP COLUMN "assessment_editors", DROP COLUMN "assessment_blocked_groups";
