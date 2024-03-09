CREATE TABLE "vtubers" (
    "id" varchar(24) NOT NULL,
    "name" varchar NOT NULL,
    "item_count" integer DEFAULT 0,
    "created_at" TIMESTAMP(0) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(0) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("id")
);

CREATE TABLE "videos" (
    "id" varchar(11) NOT NULL,
    "title" varchar NOT NULL,
    "duration" varchar NOT NULL,
    "viewers" integer NOT NULL,
    "content" varchar NOT NULL,
    "announced" boolean DEFAULT false,
    "scheduled_start_time" timestamp,
    "created_at" TIMESTAMP(0) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(0) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("id")
);

CREATE TABLE "users" (
    "token" varchar(200) NOT NULL,
    "created_at" TIMESTAMP(0) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(0) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("token")
);