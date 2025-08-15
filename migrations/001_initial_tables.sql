-- +goose Up
-- +goose StatementBegin

-- Create ENUM types
CREATE TYPE periodenum AS ENUM (
    'DAILY',
    'WEEKLY',
    'MONTHLY',
    'YEARLY',
    'CUSTOM'
);

-- Create tables
CREATE TABLE account_types (
    id SERIAL PRIMARY KEY,
    type_name VARCHAR(100) NOT NULL,
    is_credit BOOLEAN DEFAULT FALSE NOT NULL,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

CREATE TABLE currencies (
    id SERIAL PRIMARY KEY,
    code VARCHAR(3) NOT NULL,
    name VARCHAR NOT NULL,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR NOT NULL,
    first_name VARCHAR,
    last_name VARCHAR,
    password_hash VARCHAR NOT NULL,
    is_active BOOLEAN DEFAULT TRUE NOT NULL,
    base_currency_id INTEGER NOT NULL,
    is_deleted BOOLEAN DEFAULT FALSE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

CREATE TABLE accounts (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    account_type_id INTEGER NOT NULL,
    currency_id INTEGER NOT NULL,
    initial_balance NUMERIC NOT NULL,
    balance NUMERIC NOT NULL,
    name VARCHAR(100) NOT NULL,
    opening_date TIMESTAMPTZ,
    comment VARCHAR,
    is_hidden BOOLEAN NOT NULL,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    show_in_reports BOOLEAN,
    credit_limit NUMERIC,
    is_archived BOOLEAN DEFAULT FALSE,
    archived_at TIMESTAMPTZ
);

CREATE TABLE activation_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    token VARCHAR(32) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

CREATE TABLE budgets (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    name VARCHAR(200) NOT NULL,
    target_amount NUMERIC NOT NULL,
    collected_amount NUMERIC NOT NULL,
    period periodenum NOT NULL,
    repeat BOOLEAN NOT NULL,
    start_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ NOT NULL,
    included_categories TEXT,
    comment VARCHAR,
    is_deleted BOOLEAN DEFAULT FALSE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    is_archived BOOLEAN DEFAULT FALSE NOT NULL,
    currency_id INTEGER NOT NULL
);

CREATE TABLE default_categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR NOT NULL,
    parent_id INTEGER,
    is_income BOOLEAN DEFAULT FALSE NOT NULL,
    is_deleted BOOLEAN DEFAULT FALSE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

CREATE TABLE exchange_rates (
    id SERIAL PRIMARY KEY,
    rates JSONB,
    actual_date DATE NOT NULL,
    base_currency_code VARCHAR(3) NOT NULL,
    service_name VARCHAR(50) NOT NULL,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

CREATE TABLE languages (
    id SERIAL PRIMARY KEY,
    name VARCHAR NOT NULL,
    code VARCHAR(50) NOT NULL,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

CREATE TABLE user_categories (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    name VARCHAR NOT NULL,
    parent_id INTEGER,
    is_income BOOLEAN DEFAULT FALSE NOT NULL,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

CREATE TABLE transaction_templates (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    label VARCHAR(255),
    category_id INTEGER,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    account_id INTEGER NOT NULL,
    amount NUMERIC NOT NULL,
    new_balance NUMERIC,
    category_id INTEGER,
    label VARCHAR(50),
    is_income BOOLEAN NOT NULL,
    is_transfer BOOLEAN DEFAULT FALSE NOT NULL,
    linked_transaction_id INTEGER,
    base_currency_amount NUMERIC,
    notes VARCHAR,
    date_time TIMESTAMPTZ,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

CREATE TABLE user_settings (
    id SERIAL PRIMARY KEY,
    settings JSON NOT NULL,
    user_id INTEGER NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- Add foreign key constraints
ALTER TABLE users ADD CONSTRAINT users_base_currency_id_fkey FOREIGN KEY (base_currency_id) REFERENCES currencies(id) ON DELETE CASCADE;

ALTER TABLE accounts ADD CONSTRAINT accounts_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE accounts ADD CONSTRAINT accounts_account_type_id_fkey FOREIGN KEY (account_type_id) REFERENCES account_types(id) ON DELETE CASCADE;
ALTER TABLE accounts ADD CONSTRAINT accounts_currency_id_fkey FOREIGN KEY (currency_id) REFERENCES currencies(id) ON DELETE CASCADE;

ALTER TABLE activation_tokens ADD CONSTRAINT activation_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE budgets ADD CONSTRAINT budgets_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE budgets ADD CONSTRAINT budgets_currency_id_fkey FOREIGN KEY (currency_id) REFERENCES currencies(id) ON DELETE CASCADE;

ALTER TABLE default_categories ADD CONSTRAINT default_categories_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES default_categories(id) ON DELETE CASCADE;

ALTER TABLE user_categories ADD CONSTRAINT user_categories_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE user_categories ADD CONSTRAINT user_categories_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES user_categories(id) ON DELETE CASCADE;

ALTER TABLE transaction_templates ADD CONSTRAINT transaction_templates_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE transaction_templates ADD CONSTRAINT transaction_templates_category_id_fkey FOREIGN KEY (category_id) REFERENCES user_categories(id) ON DELETE CASCADE;

ALTER TABLE transactions ADD CONSTRAINT transactions_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE transactions ADD CONSTRAINT transactions_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE;
ALTER TABLE transactions ADD CONSTRAINT transactions_category_id_fkey FOREIGN KEY (category_id) REFERENCES user_categories(id) ON DELETE CASCADE;
ALTER TABLE transactions ADD CONSTRAINT transactions_linked_transaction_id_fkey FOREIGN KEY (linked_transaction_id) REFERENCES transactions(id) ON DELETE CASCADE;

ALTER TABLE user_settings ADD CONSTRAINT user_settings_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Add unique constraints
ALTER TABLE activation_tokens ADD CONSTRAINT activation_tokens_token_key UNIQUE (token);
ALTER TABLE exchange_rates ADD CONSTRAINT unique_service_date UNIQUE (service_name, actual_date);

-- Create indexes
CREATE INDEX ix_account_types_type_name ON account_types USING btree (type_name);

CREATE INDEX ix_accounts_name ON accounts USING btree (name);
CREATE INDEX ix_accounts_user_id ON accounts USING btree (user_id);

CREATE INDEX ix_budgets_currency_id ON budgets USING btree (currency_id);
CREATE INDEX ix_budgets_name ON budgets USING btree (name);
CREATE INDEX ix_budgets_user_id ON budgets USING btree (user_id);

CREATE INDEX ix_currencies_code ON currencies USING btree (code);
CREATE INDEX ix_currencies_name ON currencies USING btree (name);

CREATE INDEX ix_default_categories_name ON default_categories USING btree (name);

CREATE INDEX ix_exchange_rates_actual_date ON exchange_rates USING btree (actual_date);

CREATE INDEX ix_transaction_templates_category_id ON transaction_templates USING btree (category_id);
CREATE INDEX ix_transaction_templates_label ON transaction_templates USING btree (label);
CREATE INDEX ix_transaction_templates_user_id ON transaction_templates USING btree (user_id);

CREATE INDEX ix_transactions_account_id ON transactions USING btree (account_id);
CREATE INDEX ix_transactions_category_id ON transactions USING btree (category_id);
CREATE INDEX ix_transactions_date_time ON transactions USING btree (date_time);
CREATE INDEX ix_transactions_label ON transactions USING btree (label);
CREATE INDEX ix_transactions_linked_transaction_id ON transactions USING btree (linked_transaction_id);
CREATE INDEX ix_transactions_user_id ON transactions USING btree (user_id);

CREATE INDEX ix_user_categories_name ON user_categories USING btree (name);
CREATE INDEX ix_user_categories_parent_id ON user_categories USING btree (parent_id);
CREATE INDEX ix_user_categories_user_id ON user_categories USING btree (user_id);

CREATE UNIQUE INDEX ix_users_email ON users USING btree (email);
CREATE INDEX ix_users_first_name ON users USING btree (first_name);
CREATE INDEX ix_users_last_name ON users USING btree (last_name);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop indexes (not needed as they'll be dropped with tables)

-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS user_settings CASCADE;
DROP TABLE IF EXISTS transactions CASCADE;
DROP TABLE IF EXISTS transaction_templates CASCADE;
DROP TABLE IF EXISTS user_categories CASCADE;
DROP TABLE IF EXISTS exchange_rates CASCADE;
DROP TABLE IF EXISTS default_categories CASCADE;
DROP TABLE IF EXISTS budgets CASCADE;
DROP TABLE IF EXISTS activation_tokens CASCADE;
DROP TABLE IF EXISTS accounts CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS currencies CASCADE;
DROP TABLE IF EXISTS account_types CASCADE;
DROP TABLE IF EXISTS languages CASCADE;

-- Drop ENUM types
DROP TYPE IF EXISTS periodenum;

-- +goose StatementEnd