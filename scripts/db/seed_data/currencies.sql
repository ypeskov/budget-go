-- Seed data for currencies table
-- Clear existing data and reset sequence
DELETE FROM currencies;
ALTER SEQUENCE currencies_id_seq RESTART WITH 1;

-- Insert default currencies
INSERT INTO currencies (id, code, name, created_at, updated_at) VALUES
(1, 'USD', 'United States Dollar', NOW(), NOW()),
(2, 'UAH', 'Ukrainian Hryvna', NOW(), NOW()),
(3, 'EUR', 'Euro', NOW(), NOW()),
(4, 'BGN', 'Bulgarian Lev', NOW(), NOW()),
(5, 'AUD', 'Australian Dollar', NOW(), NOW()),
(6, 'BRL', 'Brazilian Real', NOW(), NOW()),
(7, 'CAD', 'Canadian Dollar', NOW(), NOW()),
(8, 'CHF', 'Swiss Franc', NOW(), NOW()),
(9, 'CNY', 'Chinese Yuan', NOW(), NOW()),
(10, 'CZK', 'Czech Koruna', NOW(), NOW()),
(11, 'DKK', 'Danish Krone', NOW(), NOW()),
(12, 'GBP', 'British Pound', NOW(), NOW()),
(13, 'HKD', 'Hong Kong Dollar', NOW(), NOW()),
(14, 'HRK', 'Croatian Kuna', NOW(), NOW()),
(15, 'HUF', 'Hungarian Forint', NOW(), NOW()),
(16, 'IDR', 'Indonesian Rupiah', NOW(), NOW()),
(17, 'ILS', 'Israeli New Shekel', NOW(), NOW()),
(18, 'INR', 'Indian Rupee', NOW(), NOW()),
(19, 'ISK', 'Icelandic Krona', NOW(), NOW()),
(20, 'JPY', 'Japanese Yen', NOW(), NOW()),
(21, 'KRW', 'South Korean Won', NOW(), NOW()),
(22, 'MXN', 'Mexican Peso', NOW(), NOW()),
(23, 'MYR', 'Malaysian Ringgit', NOW(), NOW()),
(24, 'NOK', 'Norwegian Krone', NOW(), NOW()),
(25, 'NZD', 'New Zealand Dollar', NOW(), NOW()),
(26, 'PHP', 'Philippine Peso', NOW(), NOW()),
(27, 'PLN', 'Polish Zloty', NOW(), NOW()),
(28, 'RON', 'Romanian Leu', NOW(), NOW()),
(29, 'RUB', 'Russian Ruble', NOW(), NOW()),
(30, 'SEK', 'Swedish Krona', NOW(), NOW()),
(31, 'SGD', 'Singapore Dollar', NOW(), NOW()),
(32, 'THB', 'Thai Baht', NOW(), NOW()),
(33, 'TRY', 'Turkish Lira', NOW(), NOW()),
(34, 'ZAR', 'South African Rand', NOW(), NOW());

-- Update sequence to next available ID
SELECT setval('currencies_id_seq', (SELECT MAX(id) FROM currencies) + 1);