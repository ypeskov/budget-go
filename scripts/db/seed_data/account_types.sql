-- Seed data for account_types table
-- Clear existing data and reset sequence
DELETE FROM account_types;
ALTER SEQUENCE account_types_id_seq RESTART WITH 1;

-- Insert default account types
INSERT INTO account_types (id, type_name, is_credit, created_at, updated_at) VALUES
(1, 'cash', false, NOW(), NOW()),
(2, 'regular_bank', false, NOW(), NOW()),
(3, 'debit_card', false, NOW(), NOW()),
(4, 'credit_card', true, NOW(), NOW()),
(5, 'loan', true, NOW(), NOW());

-- Update sequence to next available ID
SELECT setval('account_types_id_seq', (SELECT MAX(id) FROM account_types) + 1);