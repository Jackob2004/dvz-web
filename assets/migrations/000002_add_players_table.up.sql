CREATE TABLE "players" (
    "id" INTEGER,
    "uuid" TEXT NOT NULL UNIQUE,
    "username" TEXT NOT NULL,
    "experience_points" INTEGER NOT NULL DEFAULT 0, -- 100 experience points = 1 level
    "dwarf_kills" INTEGER NOT NULL DEFAULT 0,
    "zombie_kills" INTEGER NOT NULL DEFAULT 0,
    "ai_zombie_kills" INTEGER NOT NULL DEFAULT 0,
    "shrine_kills" INTEGER NOT NULL DEFAULT 0,
    "deaths" INTEGER NOT NULL DEFAULT 0,
    "longest_time_survived" NUMERIC NOT NULL DEFAULT 0,
    "total_play_time" NUMERIC NOT NULL DEFAULT 0,
    "games_played" INTEGER NOT NULL DEFAULT 0,
    "rampage_kill_streak" INTEGER NOT NULL DEFAULT 0,
    "bow_shoots" INTEGER NOT NULL DEFAULT 0,
    "bow_hits" INTEGER NOT NULL DEFAULT 0,
    "bow_accuracy" REAL NOT NULL GENERATED ALWAYS AS (
        CASE WHEN "bow_shoots" = 0 THEN 0
            ELSE "bow_hits" * 100.0 / "bow_shoots"
        END
        ) STORED,
    "level" INTEGER NOT NULL GENERATED ALWAYS AS ("experience_points" / 100) STORED,
    PRIMARY KEY("id")
);

CREATE INDEX "uuid_index" ON "players" ("uuid");
CREATE INDEX "level_index" ON "players" ("level");