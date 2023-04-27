CREATE TABLE "public"."wallets" (
    "name" text NOT NULL,
    "amount" int8 NOT NULL DEFAULT 0 CHECK (amount >= (0)::bigint),
    "create_time" timestamp NOT NULL DEFAULT now(),
    PRIMARY KEY ("name")
);

CREATE SEQUENCE transfer_history_id_seq;

CREATE TYPE "public"."transfer_direction" AS ENUM ('deposit', 'transfer', 'withdrawal');

CREATE TABLE "public"."transfer_history" (
    "id" int8 NOT NULL DEFAULT nextval('transfer_history_id_seq'::regclass),
    "wallet" text NOT NULL,
    "direction" "public"."transfer_direction" NOT NULL,
    "amount" int8 NOT NULL,
    "meta" jsonb,
    "time" timestamp NOT NULL DEFAULT now(),
    PRIMARY KEY ("id")
);

CREATE INDEX "wallet_time_key" ON "public"."transfer_history" USING BTREE ("wallet","time");