CREATE TABLE "player_skins" (
    "player_id" INTEGER,
    "head_skin" BLOB NOT NULL,
    "skin_hash" TEXT NOT NULL,
    "fetched_at" INTEGER NOT NULL, -- Unix seconds
    PRIMARY KEY("player_id"),
    FOREIGN KEY("player_id") REFERENCES "players"("id")
);