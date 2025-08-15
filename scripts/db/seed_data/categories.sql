-- Seed data for categories table
-- Clear existing data and reset sequence
DELETE FROM categories;
ALTER SEQUENCE categories_id_seq RESTART WITH 1;

-- Insert default categories (expenses)
INSERT INTO categories (id, name, parent_id, is_income, created_at, updated_at) VALUES
(1, 'Life', NULL, false, NOW(), NOW()),
(2, 'Food', NULL, false, NOW(), NOW()),
(3, 'Automobile', NULL, false, NOW(), NOW()),
(4, 'Transport', NULL, false, NOW(), NOW()),
(5, 'Housing', NULL, false, NOW(), NOW()),
(6, 'Health', NULL, false, NOW(), NOW()),
(7, 'Education', NULL, false, NOW(), NOW()),
(8, 'Entertainment', NULL, false, NOW(), NOW()),
(9, 'Finances', NULL, false, NOW(), NOW()),
(10, 'Other', NULL, false, NOW(), NOW()),
-- Subcategories
(11, 'Parking', 3, false, NOW(), NOW()),
(12, 'Fuel', 3, false, NOW(), NOW()),
(13, 'Service', 3, false, NOW(), NOW()),
(14, 'Taxi', 4, false, NOW(), NOW()),
(15, 'Meat', 2, false, NOW(), NOW()),
-- Income categories
(16, 'Salary', NULL, true, NOW(), NOW()),
(17, 'Deposit', NULL, true, NOW(), NOW()),
(18, 'Present', NULL, true, NOW(), NOW()),
(19, 'Rent', NULL, true, NOW(), NOW()),
(20, 'Social', NULL, true, NOW(), NOW()),
(21, 'Other', NULL, true, NOW(), NOW());

-- Update sequence to next available ID
SELECT setval('categories_id_seq', (SELECT MAX(id) FROM categories) + 1);