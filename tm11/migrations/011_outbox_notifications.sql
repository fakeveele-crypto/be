-- Migration 011: outbox dan notifications untuk email asinkron.
-- psql -U postgres -d praktikum_backend -f migrations/011_outbox_notifications.sql

-- outbox = kotak surat keluar. API menaruh kejadian di sini bersama KRS,
-- lalu relay mengambil surat itu dan mengirim ke Redis sebagai task.
CREATE TABLE IF NOT EXISTS outbox (
    id           BIGSERIAL    PRIMARY KEY,
    event_type   VARCHAR(50)  NOT NULL,
    payload      JSONB        NOT NULL,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ
);

-- Relay hanya mencari surat yang belum dikirim, jadi index-nya parsial.
CREATE INDEX IF NOT EXISTS outbox_pending_idx ON outbox (id) WHERE processed_at IS NULL;

-- notifications = buku tanda terima. Bila task sudah masuk ke buku ini,
-- maka email tidak boleh dikirim dua kali untuk outbox yang sama.
CREATE TABLE IF NOT EXISTS notifications (
    id         BIGSERIAL    PRIMARY KEY,
    outbox_id  BIGINT       NOT NULL,
    user_id    INTEGER      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title      VARCHAR(150) NOT NULL,
    message    TEXT         NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    -- Buku tanda terima: satu surat (id outbox) hanya boleh tercatat SATU kali.
    -- Inilah penjaga idempotency, bukan kode aplikasi.
    CONSTRAINT notifications_outbox_id_unique UNIQUE (outbox_id)
);

CREATE INDEX IF NOT EXISTS notifications_user_idx ON notifications (user_id, id DESC);
