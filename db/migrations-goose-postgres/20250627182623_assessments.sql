-- +goose Up
-- create "assessment_history" table
CREATE TABLE "assessment_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "name" character varying NOT NULL, "assessment_type" character varying NOT NULL DEFAULT 'INTERNAL', "template_id" character varying NOT NULL, "assessment_owner_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "assessmenthistory_history_time" to table: "assessment_history"
CREATE INDEX "assessmenthistory_history_time" ON "assessment_history" ("history_time");
-- create "assessment_response_history" table
CREATE TABLE "assessment_response_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "assessment_id" character varying NOT NULL, "user_id" character varying NOT NULL, "status" character varying NOT NULL DEFAULT 'NOT_STARTED', "assigned_at" timestamptz NULL, "started_at" timestamptz NOT NULL, "completed_at" timestamptz NULL, "due_date" timestamptz NULL, "response_data_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "assessmentresponsehistory_history_time" to table: "assessment_response_history"
CREATE INDEX "assessmentresponsehistory_history_time" ON "assessment_response_history" ("history_time");
-- create "assessments" table
CREATE TABLE "assessments" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "name" character varying NOT NULL, "assessment_type" character varying NOT NULL DEFAULT 'INTERNAL', "template_id" character varying NOT NULL, "assessment_owner_id" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "assessments_groups_assessment_owner" FOREIGN KEY ("assessment_owner_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "assessments_organizations_assessments" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "assessments_templates_template" FOREIGN KEY ("template_id") REFERENCES "templates" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- create index "assessment_id" to table: "assessments"
CREATE UNIQUE INDEX "assessment_id" ON "assessments" ("id");
-- create index "assessment_name_owner_id" to table: "assessments"
CREATE UNIQUE INDEX "assessment_name_owner_id" ON "assessments" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- create index "assessment_owner_id" to table: "assessments"
CREATE INDEX "assessment_owner_id" ON "assessments" ("owner_id") WHERE (deleted_at IS NULL);
-- create "assessment_blocked_groups" table
CREATE TABLE "assessment_blocked_groups" ("assessment_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("assessment_id", "group_id"), CONSTRAINT "assessment_blocked_groups_assessment_id" FOREIGN KEY ("assessment_id") REFERENCES "assessments" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "assessment_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "assessment_editors" table
CREATE TABLE "assessment_editors" ("assessment_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("assessment_id", "group_id"), CONSTRAINT "assessment_editors_assessment_id" FOREIGN KEY ("assessment_id") REFERENCES "assessments" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "assessment_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "assessment_responses" table
CREATE TABLE "assessment_responses" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "status" character varying NOT NULL DEFAULT 'NOT_STARTED', "assigned_at" timestamptz NULL, "started_at" timestamptz NOT NULL, "completed_at" timestamptz NULL, "due_date" timestamptz NULL, "assessment_id" character varying NOT NULL, "user_id" character varying NOT NULL, "response_data_id" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "assessment_responses_assessments_assessment_responses" FOREIGN KEY ("assessment_id") REFERENCES "assessments" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "assessment_responses_document_data_document" FOREIGN KEY ("response_data_id") REFERENCES "document_data" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "assessment_responses_organizations_assessment_responses" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "assessment_responses_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- create index "assessmentresponse_assessment_id_user_id" to table: "assessment_responses"
CREATE UNIQUE INDEX "assessmentresponse_assessment_id_user_id" ON "assessment_responses" ("assessment_id", "user_id") WHERE (deleted_at IS NULL);
-- create index "assessmentresponse_completed_at" to table: "assessment_responses"
CREATE INDEX "assessmentresponse_completed_at" ON "assessment_responses" ("completed_at");
-- create index "assessmentresponse_due_date" to table: "assessment_responses"
CREATE INDEX "assessmentresponse_due_date" ON "assessment_responses" ("due_date");
-- create index "assessmentresponse_id" to table: "assessment_responses"
CREATE UNIQUE INDEX "assessmentresponse_id" ON "assessment_responses" ("id");
-- create index "assessmentresponse_owner_id" to table: "assessment_responses"
CREATE INDEX "assessmentresponse_owner_id" ON "assessment_responses" ("owner_id") WHERE (deleted_at IS NULL);
-- create index "assessmentresponse_status" to table: "assessment_responses"
CREATE INDEX "assessmentresponse_status" ON "assessment_responses" ("status");
-- create "assessment_users" table
CREATE TABLE "assessment_users" ("assessment_id" character varying NOT NULL, "user_id" character varying NOT NULL, PRIMARY KEY ("assessment_id", "user_id"), CONSTRAINT "assessment_users_assessment_id" FOREIGN KEY ("assessment_id") REFERENCES "assessments" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "assessment_users_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "assessment_viewers" table
CREATE TABLE "assessment_viewers" ("assessment_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("assessment_id", "group_id"), CONSTRAINT "assessment_viewers_assessment_id" FOREIGN KEY ("assessment_id") REFERENCES "assessments" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "assessment_viewers_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);

-- +goose Down
-- reverse: create "assessment_viewers" table
DROP TABLE "assessment_viewers";
-- reverse: create "assessment_users" table
DROP TABLE "assessment_users";
-- reverse: create index "assessmentresponse_status" to table: "assessment_responses"
DROP INDEX "assessmentresponse_status";
-- reverse: create index "assessmentresponse_owner_id" to table: "assessment_responses"
DROP INDEX "assessmentresponse_owner_id";
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
-- reverse: create "assessment_editors" table
DROP TABLE "assessment_editors";
-- reverse: create "assessment_blocked_groups" table
DROP TABLE "assessment_blocked_groups";
-- reverse: create index "assessment_owner_id" to table: "assessments"
DROP INDEX "assessment_owner_id";
-- reverse: create index "assessment_name_owner_id" to table: "assessments"
DROP INDEX "assessment_name_owner_id";
-- reverse: create index "assessment_id" to table: "assessments"
DROP INDEX "assessment_id";
-- reverse: create "assessments" table
DROP TABLE "assessments";
-- reverse: create index "assessmentresponsehistory_history_time" to table: "assessment_response_history"
DROP INDEX "assessmentresponsehistory_history_time";
-- reverse: create "assessment_response_history" table
DROP TABLE "assessment_response_history";
-- reverse: create index "assessmenthistory_history_time" to table: "assessment_history"
DROP INDEX "assessmenthistory_history_time";
-- reverse: create "assessment_history" table
DROP TABLE "assessment_history";
