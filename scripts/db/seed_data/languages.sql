-- Seed data for languages table
-- Clear existing data and reset sequence
DELETE FROM languages;
ALTER SEQUENCE languages_id_seq RESTART WITH 1;

-- Insert default languages
INSERT INTO languages (id, name, code, created_at, updated_at) VALUES
(1, 'English', 'en', NOW(), NOW()),
(3, 'Українська', 'uk', NOW(), NOW());

-- Update sequence to next available ID
SELECT setval('languages_id_seq', (SELECT MAX(id) FROM languages) + 1);