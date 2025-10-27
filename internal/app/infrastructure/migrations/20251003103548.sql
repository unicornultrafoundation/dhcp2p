-- Set comment to schema: "public"
COMMENT ON SCHEMA "public" IS 'Schema for dhcp2p';
-- Create "alloc_state" table
CREATE TABLE "public"."alloc_state" (
  "id" serial NOT NULL,
  "last_token_id" bigint NOT NULL,
  "max_token_id" bigint NOT NULL DEFAULT 168162304,
  PRIMARY KEY ("id")
);
-- Create "leases" table
CREATE TABLE "public"."leases" (
  "token_id" bigint NOT NULL,
  "peer_id" character varying(128) NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("token_id")
);
-- Create index "idx_leases_expires_at" to table: "leases"
CREATE INDEX "idx_leases_expires_at" ON "public"."leases" ("expires_at");
-- Create "nonces" table
CREATE TABLE "public"."nonces" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "peer_id" character varying(128) NOT NULL,
  "issued_at" timestamptz NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "used" boolean NOT NULL DEFAULT false,
  "used_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
