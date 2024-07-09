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
    "song" boolean DEFAULT false, 
    "viewers" integer NOT NULL, 
    "content" varchar NOT NULL, 
    "scheduled_start_time" timestamp, 
    "thumbnail" varchar NOT NULL, 
    "created_at" TIMESTAMP(0) NOT NULL DEFAULT CURRENT_TIMESTAMP, 
    "updated_at" TIMESTAMP(0) NOT NULL DEFAULT CURRENT_TIMESTAMP, 
    PRIMARY KEY ("id")
);

CREATE TABLE "users" (
    "token" varchar(1000) NOT NULL,
    "song" boolean DEFAULT false,
    "keyword" boolean DEFAULT false,
    "keyword_text" varchar(100),
    "created_at" TIMESTAMP(0) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(0) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("token")
);