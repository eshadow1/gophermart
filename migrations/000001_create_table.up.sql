-- +migrate Up

CREATE TABLE IF NOT EXISTS users (
    id            BIGSERIAL PRIMARY KEY,
    login         VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE users IS 'Хранилище зарегистрированных пользователей';

CREATE TYPE order_status AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');
CREATE TABLE IF NOT EXISTS orders (
                                      id              BIGSERIAL PRIMARY KEY,
                                      user_id         BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                      number          VARCHAR(255) NOT NULL UNIQUE,
                                      status          order_status NOT NULL DEFAULT 'NEW',
                                      accrual         NUMERIC(12,2),
                                      uploaded_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                      last_checked_at TIMESTAMPTZ
);

COMMENT ON TABLE orders IS 'Хранилище обрабатываемых заказов';

CREATE TABLE IF NOT EXISTS user_balances (
                                             user_id   BIGINT       PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
                                             current   NUMERIC(12,2) NOT NULL DEFAULT 0.00 CHECK (current >= 0),
                                             withdrawn NUMERIC(12,2) NOT NULL DEFAULT 0.00 CHECK (withdrawn >= 0)
);

COMMENT ON TABLE user_balances IS 'Хранилище баланса системы лояльности для пользователей';

CREATE TABLE IF NOT EXISTS withdrawals (
                                           id           BIGSERIAL PRIMARY KEY,
                                           user_id      BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                           order_number VARCHAR(255) NOT NULL,
                                           sum          NUMERIC(12,2) NOT NULL CHECK (sum > 0),
                                           processed_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE withdrawals IS 'Хранилище обрабатываемых заказов с системой лояльности для пользователей';
