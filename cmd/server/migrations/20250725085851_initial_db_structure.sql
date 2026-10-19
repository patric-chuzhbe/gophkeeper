-- +goose Up
-- +goose StatementBegin
CREATE TABLE metadata
(
    metadata_id UUID         NOT NULL DEFAULT gen_random_uuid(),
    record_id   UUID NULL,
    name        VARCHAR(255) NOT NULL,
    value       TEXT NULL,
    CONSTRAINT PK_METADATA PRIMARY KEY (metadata_id)
);

CREATE TABLE users
(
    user_id UUID         NOT NULL DEFAULT gen_random_uuid(),
    login   VARCHAR(255) NOT NULL,
    pass    VARCHAR(255) NULL,
    CONSTRAINT PK_USERS PRIMARY KEY (user_id)
);

CREATE UNIQUE INDEX Index_1 ON users (login);

CREATE TABLE vault_records
(
    record_id UUID         NOT NULL DEFAULT gen_random_uuid(),
    user_id   UUID NULL,
    data      TEXT NULL,
    type      VARCHAR(255) NOT NULL DEFAULT 'LOG_PASS'
        CONSTRAINT CKC_TYPE_VAULT_RE CHECK (type IN ('LOG_PASS', 'TEXT', 'BINARY', 'BANK_CARD')),
    nonce     TEXT         not null,
    CONSTRAINT PK_VAULT_RECORDS PRIMARY KEY (record_id)
);

ALTER TABLE metadata
    ADD CONSTRAINT FK_METADATA_REFERENCE_VAULT_RE FOREIGN KEY (record_id)
        REFERENCES vault_records (record_id)
        ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE vault_records
    ADD CONSTRAINT FK_VAULT_RE_REFERENCE_USERS FOREIGN KEY (user_id)
        REFERENCES users (user_id)
        ON DELETE CASCADE ON UPDATE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE metadata DROP CONSTRAINT FK_METADATA_REFERENCE_VAULT_RE;
ALTER TABLE vault_records DROP CONSTRAINT FK_VAULT_RE_REFERENCE_USERS;

DROP TABLE IF EXISTS metadata;
DROP TABLE IF EXISTS vault_records;
DROP TABLE IF EXISTS users;

DROP INDEX IF EXISTS Index_1;
-- +goose StatementEnd
