-- +goose Up
-- +goose StatementBegin
-- Memories are a derived recall cache. Existing rows may have embeddings from a
-- different model width, so clear them before resizing the vector column.
-- Run with the app stopped or no concurrent memory writes; this migration takes
-- an exclusive lock and resets the derived semantic-memory cache.
DO $$
BEGIN
  IF to_regclass('public.memories') IS NOT NULL THEN
    DROP INDEX IF EXISTS idx_memories_embedding_hnsw;
    LOCK TABLE memories IN ACCESS EXCLUSIVE MODE;
    DELETE FROM memories;
    ALTER TABLE memories
      ALTER COLUMN embedding TYPE vector(768)
      USING NULL::vector(768);
    CREATE INDEX idx_memories_embedding_hnsw
      ON memories USING hnsw (embedding vector_cosine_ops)
      WITH (m = 16, ef_construction = 64);
  END IF;
END
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $$
BEGIN
  IF to_regclass('public.memories') IS NOT NULL THEN
    DROP INDEX IF EXISTS idx_memories_embedding_hnsw;
    LOCK TABLE memories IN ACCESS EXCLUSIVE MODE;
    DELETE FROM memories;
    ALTER TABLE memories
      ALTER COLUMN embedding TYPE vector(1536)
      USING NULL::vector(1536);
    CREATE INDEX idx_memories_embedding_hnsw
      ON memories USING hnsw (embedding vector_cosine_ops)
      WITH (m = 16, ef_construction = 64);
  END IF;
END
$$;
-- +goose StatementEnd
