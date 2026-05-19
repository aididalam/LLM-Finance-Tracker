-- ─────────────────────────────────────────────────────────────────────────────
-- Seed script: 6 months of realistic expense + income data
-- Run: mysql -u root -proot1234 llmexpensetracker < scripts/seed_6months.sql
-- ─────────────────────────────────────────────────────────────────────────────

SET @food     = (SELECT category_id FROM categories WHERE name = 'Food'          LIMIT 1);
SET @trans    = (SELECT category_id FROM categories WHERE name = 'Transport'     LIMIT 1);
SET @shop     = (SELECT category_id FROM categories WHERE name = 'Shopping'      LIMIT 1);
SET @ent      = (SELECT category_id FROM categories WHERE name = 'Entertainment' LIMIT 1);
SET @health   = (SELECT category_id FROM categories WHERE name = 'Health'        LIMIT 1);
SET @bills    = (SELECT category_id FROM categories WHERE name = 'Bills'         LIMIT 1);
SET @edu      = (SELECT category_id FROM categories WHERE name = 'Education'     LIMIT 1);
SET @salary   = (SELECT category_id FROM categories WHERE name = 'Salary'        LIMIT 1);
SET @freelance= (SELECT category_id FROM categories WHERE name = 'Freelance'     LIMIT 1);
SET @other    = (SELECT category_id FROM categories WHERE name = 'Other'         LIMIT 1);

-- ── helper: today minus N months ─────────────────────────────────────────────
-- All dates are relative to today so the dashboard always looks current.

INSERT INTO entries (entry_id, transaction_type, category_id, amount, currency, description, merchant, expense_date, reporting_date, payment_method, source) VALUES

-- ════════════════ MONTH -6 (6 months ago) ════════════════
(UUID(),'income',  @salary,    75000.00,'BDT','Monthly salary',                   'Employer',        DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 6 MONTH), '%Y-%m-01'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 6 MONTH), '%Y-%m-01'),'other','web'),
(UUID(),'income',  @freelance,  12000.00,'BDT','Web project payment',             'Client A',        DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 6 MONTH), '%Y-%m-10'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 6 MONTH), '%Y-%m-10'),'other','web'),
(UUID(),'expense', @food,        450.00,'BDT','Lunch at restaurant',              'Kacchi Bhai',     DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 6 MONTH), '%Y-%m-03'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 6 MONTH), '%Y-%m-03'),'cash','web'),
(UUID(),'expense', @food,        120.00,'BDT','Morning tea and snacks',           'Corner café',     DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 6 MONTH), '%Y-%m-05'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 6 MONTH), '%Y-%m-05'),'cash','web'),
(UUID(),'expense', @food,        680.00,'BDT','Grocery shopping',                 'Shwapno',         DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 6 MONTH), '%Y-%m-07'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 6 MONTH), '%Y-%m-07'),'cash','web'),
(UUID(),'expense', @food,        350.00,'BDT','Dinner with family',               'Star Kabab',      DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 6 MONTH), '%Y-%m-12'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 6 MONTH), '%Y-%m-12'),'cash','web'),
(UUID(),'expense', @food,        220.00,'BDT','Office lunch',                     'Canteen',         DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 6 MONTH), '%Y-%m-15'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 6 MONTH), '%Y-%m-15'),'cash','web'),
(UUID(),'expense', @food,        900.00,'BDT','Weekly groceries',                 'Meena Bazar',     DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 6 MONTH), '%Y-%m-18'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 6 MONTH), '%Y-%m-18'),'cash','web'),
(UUID(),'expense', @trans,       200.00,'BDT','Uber ride to office',              'Uber',            DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 6 MONTH), '%Y-%m-04'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 6 MONTH), '%Y-%m-04'),'other','web'),
(UUID(),'expense', @trans,       150.00,'BDT','CNG fare',                         'Local CNG',       DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 6 MONTH), '%Y-%m-08'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 6 MONTH), '%Y-%m-08'),'cash','web'),
(UUID(),'expense', @trans,      1200.00,'BDT','Monthly bus pass',                 'BRTC',            DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 6 MONTH), '%Y-%m-01'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 6 MONTH), '%Y-%m-01'),'cash','web'),
(UUID(),'expense', @bills,      1500.00,'BDT','Electricity bill',                 'DESCO',           DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 6 MONTH), '%Y-%m-10'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 6 MONTH), '%Y-%m-10'),'other','web'),
(UUID(),'expense', @bills,       600.00,'BDT','Internet bill',                    'Grameenphone',    DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 6 MONTH), '%Y-%m-05'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 6 MONTH), '%Y-%m-05'),'other','web'),
(UUID(),'expense', @bills,      8000.00,'BDT','House rent',                       'Landlord',        DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 6 MONTH), '%Y-%m-01'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 6 MONTH), '%Y-%m-01'),'cash','web'),
(UUID(),'expense', @health,      800.00,'BDT','Doctor consultation',              'Square Hospital',  DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 6 MONTH), '%Y-%m-14'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 6 MONTH), '%Y-%m-14'),'cash','web'),
(UUID(),'expense', @health,      450.00,'BDT','Medicine purchase',                'Lazz Pharma',     DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 6 MONTH), '%Y-%m-16'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 6 MONTH), '%Y-%m-16'),'cash','web'),
(UUID(),'expense', @ent,         500.00,'BDT','Netflix subscription',             'Netflix',         DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 6 MONTH), '%Y-%m-06'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 6 MONTH), '%Y-%m-06'),'credit_card','web'),
(UUID(),'expense', @shop,       2500.00,'BDT','New shirt and pants',              'Aarong',          DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 6 MONTH), '%Y-%m-20'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 6 MONTH), '%Y-%m-20'),'credit_card','web'),

-- ════════════════ MONTH -5 ════════════════
(UUID(),'income',  @salary,    75000.00,'BDT','Monthly salary',                   'Employer',        DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 5 MONTH), '%Y-%m-01'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 5 MONTH), '%Y-%m-01'),'other','web'),
(UUID(),'income',  @freelance,   8500.00,'BDT','Logo design project',             'Client B',        DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 5 MONTH), '%Y-%m-14'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 5 MONTH), '%Y-%m-14'),'other','web'),
(UUID(),'expense', @food,        480.00,'BDT','Lunch at restaurant',              'Fakruddin',       DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 5 MONTH), '%Y-%m-02'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 5 MONTH), '%Y-%m-02'),'cash','web'),
(UUID(),'expense', @food,        750.00,'BDT','Weekly groceries',                 'Shwapno',         DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 5 MONTH), '%Y-%m-06'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 5 MONTH), '%Y-%m-06'),'cash','web'),
(UUID(),'expense', @food,        180.00,'BDT','Breakfast items',                  'Bakery',          DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 5 MONTH), '%Y-%m-09'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 5 MONTH), '%Y-%m-09'),'cash','web'),
(UUID(),'expense', @food,        420.00,'BDT','Family dinner',                    'Chillox',         DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 5 MONTH), '%Y-%m-16'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 5 MONTH), '%Y-%m-16'),'cash','web'),
(UUID(),'expense', @food,        860.00,'BDT','Groceries',                        'Meena Bazar',     DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 5 MONTH), '%Y-%m-22'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 5 MONTH), '%Y-%m-22'),'cash','web'),
(UUID(),'expense', @trans,       180.00,'BDT','Pathao ride',                      'Pathao',          DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 5 MONTH), '%Y-%m-03'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 5 MONTH), '%Y-%m-03'),'other','web'),
(UUID(),'expense', @trans,      1200.00,'BDT','Monthly bus pass',                 'BRTC',            DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 5 MONTH), '%Y-%m-01'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 5 MONTH), '%Y-%m-01'),'cash','web'),
(UUID(),'expense', @trans,       350.00,'BDT','Fuel top-up',                      'Padma Oil',       DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 5 MONTH), '%Y-%m-12'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 5 MONTH), '%Y-%m-12'),'cash','web'),
(UUID(),'expense', @bills,      1600.00,'BDT','Electricity bill',                 'DESCO',           DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 5 MONTH), '%Y-%m-10'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 5 MONTH), '%Y-%m-10'),'other','web'),
(UUID(),'expense', @bills,       600.00,'BDT','Internet bill',                    'Grameenphone',    DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 5 MONTH), '%Y-%m-05'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 5 MONTH), '%Y-%m-05'),'other','web'),
(UUID(),'expense', @bills,      8000.00,'BDT','House rent',                       'Landlord',        DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 5 MONTH), '%Y-%m-01'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 5 MONTH), '%Y-%m-01'),'cash','web'),
(UUID(),'expense', @edu,        3000.00,'BDT','Online course subscription',       'Udemy',           DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 5 MONTH), '%Y-%m-08'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 5 MONTH), '%Y-%m-08'),'credit_card','web'),
(UUID(),'expense', @ent,         500.00,'BDT','Netflix subscription',             'Netflix',         DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 5 MONTH), '%Y-%m-06'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 5 MONTH), '%Y-%m-06'),'credit_card','web'),
(UUID(),'expense', @ent,        1200.00,'BDT','Movie night with friends',         'Star Cineplex',   DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 5 MONTH), '%Y-%m-19'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 5 MONTH), '%Y-%m-19'),'cash','web'),
(UUID(),'expense', @shop,       1800.00,'BDT','Shoes',                            'Bata',            DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 5 MONTH), '%Y-%m-24'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 5 MONTH), '%Y-%m-24'),'credit_card','web'),
(UUID(),'expense', @health,      350.00,'BDT','Vitamins and supplements',         'Lazz Pharma',     DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 5 MONTH), '%Y-%m-13'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 5 MONTH), '%Y-%m-13'),'cash','web'),

-- ════════════════ MONTH -4 ════════════════
(UUID(),'income',  @salary,    75000.00,'BDT','Monthly salary',                   'Employer',        DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 4 MONTH), '%Y-%m-01'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 4 MONTH), '%Y-%m-01'),'other','web'),
(UUID(),'income',  @freelance,  18000.00,'BDT','Mobile app UI design',            'Client C',        DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 4 MONTH), '%Y-%m-12'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 4 MONTH), '%Y-%m-12'),'other','web'),
(UUID(),'income',  @freelance,   5000.00,'BDT','Content writing',                 'Client D',        DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 4 MONTH), '%Y-%m-20'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 4 MONTH), '%Y-%m-20'),'other','web'),
(UUID(),'expense', @food,        520.00,'BDT','Lunch with colleague',             'Bashundhara Food Court', DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 4 MONTH), '%Y-%m-04'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 4 MONTH), '%Y-%m-04'),'cash','web'),
(UUID(),'expense', @food,        800.00,'BDT','Grocery run',                      'Shwapno',         DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 4 MONTH), '%Y-%m-07'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 4 MONTH), '%Y-%m-07'),'cash','web'),
(UUID(),'expense', @food,        260.00,'BDT','Tea and snacks',                   'Local shop',      DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 4 MONTH), '%Y-%m-11'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 4 MONTH), '%Y-%m-11'),'cash','web'),
(UUID(),'expense', @food,        650.00,'BDT','Weekend dinner',                   'Dhaka Regency',   DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 4 MONTH), '%Y-%m-17'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 4 MONTH), '%Y-%m-17'),'credit_card','web'),
(UUID(),'expense', @food,        920.00,'BDT','Monthly groceries top-up',         'Meena Bazar',     DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 4 MONTH), '%Y-%m-23'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 4 MONTH), '%Y-%m-23'),'cash','web'),
(UUID(),'expense', @trans,       220.00,'BDT','Uber to airport',                  'Uber',            DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 4 MONTH), '%Y-%m-05'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 4 MONTH), '%Y-%m-05'),'other','web'),
(UUID(),'expense', @trans,      1200.00,'BDT','Monthly bus pass',                 'BRTC',            DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 4 MONTH), '%Y-%m-01'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 4 MONTH), '%Y-%m-01'),'cash','web'),
(UUID(),'expense', @bills,      1750.00,'BDT','Electricity bill',                 'DESCO',           DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 4 MONTH), '%Y-%m-10'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 4 MONTH), '%Y-%m-10'),'other','web'),
(UUID(),'expense', @bills,       600.00,'BDT','Internet bill',                    'Grameenphone',    DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 4 MONTH), '%Y-%m-05'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 4 MONTH), '%Y-%m-05'),'other','web'),
(UUID(),'expense', @bills,      8000.00,'BDT','House rent',                       'Landlord',        DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 4 MONTH), '%Y-%m-01'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 4 MONTH), '%Y-%m-01'),'cash','web'),
(UUID(),'expense', @shop,       4500.00,'BDT','New laptop bag',                   'Daraz',           DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 4 MONTH), '%Y-%m-15'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 4 MONTH), '%Y-%m-15'),'credit_card','web'),
(UUID(),'expense', @ent,         500.00,'BDT','Netflix subscription',             'Netflix',         DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 4 MONTH), '%Y-%m-06'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 4 MONTH), '%Y-%m-06'),'credit_card','web'),
(UUID(),'expense', @ent,        2000.00,'BDT','Concert tickets',                  'Ticketly',        DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 4 MONTH), '%Y-%m-22'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 4 MONTH), '%Y-%m-22'),'credit_card','web'),
(UUID(),'expense', @health,     2500.00,'BDT','Annual health checkup',            'Ibn Sina',        DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 4 MONTH), '%Y-%m-18'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 4 MONTH), '%Y-%m-18'),'credit_card','web'),
(UUID(),'expense', @edu,        1500.00,'BDT','Books',                            'Nilkhet',         DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 4 MONTH), '%Y-%m-10'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 4 MONTH), '%Y-%m-10'),'cash','web'),

-- ════════════════ MONTH -3 ════════════════
(UUID(),'income',  @salary,    80000.00,'BDT','Monthly salary (increment)',        'Employer',        DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 3 MONTH), '%Y-%m-01'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 3 MONTH), '%Y-%m-01'),'other','web'),
(UUID(),'income',  @freelance,  22000.00,'BDT','Full-stack development project',  'Client E',        DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 3 MONTH), '%Y-%m-08'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 3 MONTH), '%Y-%m-08'),'other','web'),
(UUID(),'expense', @food,        560.00,'BDT','Lunch',                            'Kacchi Bhai',     DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 3 MONTH), '%Y-%m-03'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 3 MONTH), '%Y-%m-03'),'cash','web'),
(UUID(),'expense', @food,        830.00,'BDT','Groceries',                        'Shwapno',         DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 3 MONTH), '%Y-%m-08'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 3 MONTH), '%Y-%m-08'),'cash','web'),
(UUID(),'expense', @food,        300.00,'BDT','Breakfast out',                    'Café Mango',      DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 3 MONTH), '%Y-%m-11'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 3 MONTH), '%Y-%m-11'),'cash','web'),
(UUID(),'expense', @food,        480.00,'BDT','Office team lunch',                'Canteen',         DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 3 MONTH), '%Y-%m-14'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 3 MONTH), '%Y-%m-14'),'cash','web'),
(UUID(),'expense', @food,        720.00,'BDT','Eid special groceries',            'Meena Bazar',     DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 3 MONTH), '%Y-%m-19'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 3 MONTH), '%Y-%m-19'),'cash','web'),
(UUID(),'expense', @food,       1400.00,'BDT','Family Eid dinner',                'Westin Dhaka',    DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 3 MONTH), '%Y-%m-22'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 3 MONTH), '%Y-%m-22'),'credit_card','web'),
(UUID(),'expense', @trans,       250.00,'BDT','Ride share',                       'Pathao',          DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 3 MONTH), '%Y-%m-04'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 3 MONTH), '%Y-%m-04'),'other','web'),
(UUID(),'expense', @trans,      1200.00,'BDT','Monthly bus pass',                 'BRTC',            DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 3 MONTH), '%Y-%m-01'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 3 MONTH), '%Y-%m-01'),'cash','web'),
(UUID(),'expense', @trans,      3500.00,'BDT','Bus ticket home (Eid travel)',     'Shyamoli Paribahan', DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 3 MONTH), '%Y-%m-18'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 3 MONTH), '%Y-%m-18'),'cash','web'),
(UUID(),'expense', @bills,      1800.00,'BDT','Electricity bill',                 'DESCO',           DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 3 MONTH), '%Y-%m-10'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 3 MONTH), '%Y-%m-10'),'other','web'),
(UUID(),'expense', @bills,       600.00,'BDT','Internet bill',                    'Grameenphone',    DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 3 MONTH), '%Y-%m-05'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 3 MONTH), '%Y-%m-05'),'other','web'),
(UUID(),'expense', @bills,      8000.00,'BDT','House rent',                       'Landlord',        DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 3 MONTH), '%Y-%m-01'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 3 MONTH), '%Y-%m-01'),'cash','web'),
(UUID(),'expense', @shop,       6000.00,'BDT','Eid clothing',                     'Aarong',          DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 3 MONTH), '%Y-%m-15'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 3 MONTH), '%Y-%m-15'),'credit_card','web'),
(UUID(),'expense', @shop,       2200.00,'BDT','Gift for parents',                 'Online',          DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 3 MONTH), '%Y-%m-20'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 3 MONTH), '%Y-%m-20'),'credit_card','web'),
(UUID(),'expense', @ent,         500.00,'BDT','Netflix subscription',             'Netflix',         DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 3 MONTH), '%Y-%m-06'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 3 MONTH), '%Y-%m-06'),'credit_card','web'),
(UUID(),'expense', @health,      600.00,'BDT','Medicine',                         'Lazz Pharma',     DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 3 MONTH), '%Y-%m-13'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 3 MONTH), '%Y-%m-13'),'cash','web'),

-- ════════════════ MONTH -2 ════════════════
(UUID(),'income',  @salary,    80000.00,'BDT','Monthly salary',                   'Employer',        DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 2 MONTH), '%Y-%m-01'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 2 MONTH), '%Y-%m-01'),'other','web'),
(UUID(),'income',  @freelance,  15000.00,'BDT','API integration project',         'Client F',        DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 2 MONTH), '%Y-%m-18'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 2 MONTH), '%Y-%m-18'),'other','web'),
(UUID(),'expense', @food,        490.00,'BDT','Lunch',                            'Fakruddin',       DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 2 MONTH), '%Y-%m-02'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 2 MONTH), '%Y-%m-02'),'cash','web'),
(UUID(),'expense', @food,        810.00,'BDT','Groceries',                        'Shwapno',         DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 2 MONTH), '%Y-%m-06'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 2 MONTH), '%Y-%m-06'),'cash','web'),
(UUID(),'expense', @food,        140.00,'BDT','Afternoon snacks',                 'Corner store',    DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 2 MONTH), '%Y-%m-10'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 2 MONTH), '%Y-%m-10'),'cash','web'),
(UUID(),'expense', @food,        390.00,'BDT','Dinner takeaway',                  'Foodpanda',       DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 2 MONTH), '%Y-%m-14'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 2 MONTH), '%Y-%m-14'),'other','web'),
(UUID(),'expense', @food,        760.00,'BDT','Weekly groceries',                 'Meena Bazar',     DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 2 MONTH), '%Y-%m-20'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 2 MONTH), '%Y-%m-20'),'cash','web'),
(UUID(),'expense', @food,        550.00,'BDT','Birthday dinner',                  'Chillox',         DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 2 MONTH), '%Y-%m-25'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 2 MONTH), '%Y-%m-25'),'credit_card','web'),
(UUID(),'expense', @trans,       170.00,'BDT','Uber',                             'Uber',            DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 2 MONTH), '%Y-%m-03'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 2 MONTH), '%Y-%m-03'),'other','web'),
(UUID(),'expense', @trans,      1200.00,'BDT','Monthly bus pass',                 'BRTC',            DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 2 MONTH), '%Y-%m-01'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 2 MONTH), '%Y-%m-01'),'cash','web'),
(UUID(),'expense', @bills,      1650.00,'BDT','Electricity bill',                 'DESCO',           DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 2 MONTH), '%Y-%m-10'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 2 MONTH), '%Y-%m-10'),'other','web'),
(UUID(),'expense', @bills,       600.00,'BDT','Internet bill',                    'Grameenphone',    DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 2 MONTH), '%Y-%m-05'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 2 MONTH), '%Y-%m-05'),'other','web'),
(UUID(),'expense', @bills,      8000.00,'BDT','House rent',                       'Landlord',        DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 2 MONTH), '%Y-%m-01'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 2 MONTH), '%Y-%m-01'),'cash','web'),
(UUID(),'expense', @ent,         500.00,'BDT','Netflix subscription',             'Netflix',         DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 2 MONTH), '%Y-%m-06'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 2 MONTH), '%Y-%m-06'),'credit_card','web'),
(UUID(),'expense', @ent,         800.00,'BDT','Spotify family plan',              'Spotify',         DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 2 MONTH), '%Y-%m-06'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 2 MONTH), '%Y-%m-06'),'credit_card','web'),
(UUID(),'expense', @shop,       3200.00,'BDT','Household items',                  'Daraz',           DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 2 MONTH), '%Y-%m-22'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 2 MONTH), '%Y-%m-22'),'credit_card','web'),
(UUID(),'expense', @edu,        2000.00,'BDT','Coursera annual plan',             'Coursera',        DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 2 MONTH), '%Y-%m-12'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 2 MONTH), '%Y-%m-12'),'credit_card','web'),
(UUID(),'expense', @health,     1200.00,'BDT','Gym membership',                   'Gold Gym',        DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 2 MONTH), '%Y-%m-01'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 2 MONTH), '%Y-%m-01'),'credit_card','web'),

-- ════════════════ MONTH -1 ════════════════
(UUID(),'income',  @salary,    80000.00,'BDT','Monthly salary',                   'Employer',        DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), '%Y-%m-01'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), '%Y-%m-01'),'other','web'),
(UUID(),'income',  @freelance,  30000.00,'BDT','E-commerce website build',        'Client G',        DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), '%Y-%m-10'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), '%Y-%m-10'),'other','web'),
(UUID(),'income',  @freelance,   7500.00,'BDT','SEO consulting',                  'Client H',        DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), '%Y-%m-22'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), '%Y-%m-22'),'other','web'),
(UUID(),'expense', @food,        510.00,'BDT','Lunch at restaurant',              'Kacchi Bhai',     DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), '%Y-%m-03'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), '%Y-%m-03'),'cash','web'),
(UUID(),'expense', @food,        870.00,'BDT','Groceries',                        'Shwapno',         DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), '%Y-%m-07'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), '%Y-%m-07'),'cash','web'),
(UUID(),'expense', @food,        190.00,'BDT','Snacks',                           'Local shop',      DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), '%Y-%m-11'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), '%Y-%m-11'),'cash','web'),
(UUID(),'expense', @food,        430.00,'BDT','Office lunch',                     'Canteen',         DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), '%Y-%m-15'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), '%Y-%m-15'),'cash','web'),
(UUID(),'expense', @food,        950.00,'BDT','Weekend groceries',                'Meena Bazar',     DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), '%Y-%m-21'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), '%Y-%m-21'),'cash','web'),
(UUID(),'expense', @food,        680.00,'BDT','Dinner out',                       'Chillox',         DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), '%Y-%m-26'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), '%Y-%m-26'),'credit_card','web'),
(UUID(),'expense', @trans,       200.00,'BDT','Pathao bike',                      'Pathao',          DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), '%Y-%m-04'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), '%Y-%m-04'),'other','web'),
(UUID(),'expense', @trans,      1200.00,'BDT','Monthly bus pass',                 'BRTC',            DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), '%Y-%m-01'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), '%Y-%m-01'),'cash','web'),
(UUID(),'expense', @trans,       450.00,'BDT','Fuel',                             'Padma Oil',       DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), '%Y-%m-14'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), '%Y-%m-14'),'cash','web'),
(UUID(),'expense', @bills,      1700.00,'BDT','Electricity bill',                 'DESCO',           DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), '%Y-%m-10'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), '%Y-%m-10'),'other','web'),
(UUID(),'expense', @bills,       600.00,'BDT','Internet bill',                    'Grameenphone',    DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), '%Y-%m-05'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), '%Y-%m-05'),'other','web'),
(UUID(),'expense', @bills,      8000.00,'BDT','House rent',                       'Landlord',        DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), '%Y-%m-01'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), '%Y-%m-01'),'cash','web'),
(UUID(),'expense', @shop,       5500.00,'BDT','New headphones',                   'Daraz',           DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), '%Y-%m-16'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), '%Y-%m-16'),'credit_card','web'),
(UUID(),'expense', @ent,         500.00,'BDT','Netflix subscription',             'Netflix',         DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), '%Y-%m-06'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), '%Y-%m-06'),'credit_card','web'),
(UUID(),'expense', @ent,        1500.00,'BDT','Weekend trip accommodation',       'Airbnb',          DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), '%Y-%m-23'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), '%Y-%m-23'),'credit_card','web'),
(UUID(),'expense', @health,     1200.00,'BDT','Gym membership',                   'Gold Gym',        DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), '%Y-%m-01'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), '%Y-%m-01'),'credit_card','web'),
(UUID(),'expense', @health,      900.00,'BDT','Dental checkup',                   'Popular Dental',  DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), '%Y-%m-18'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), '%Y-%m-18'),'cash','web'),
(UUID(),'expense', @edu,        4500.00,'BDT','Laptop for courses',               'Ryans',           DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), '%Y-%m-20'), DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), '%Y-%m-20'),'credit_card','web'),

-- ════════════════ CURRENT MONTH ════════════════
(UUID(),'income',  @salary,    80000.00,'BDT','Monthly salary',                   'Employer',        DATE_FORMAT(CURDATE(), '%Y-%m-01'), DATE_FORMAT(CURDATE(), '%Y-%m-01'),'other','web'),
(UUID(),'income',  @freelance,  10000.00,'BDT','Dashboard UI project',            'Client I',        DATE_FORMAT(CURDATE(), '%Y-%m-05'), DATE_FORMAT(CURDATE(), '%Y-%m-05'),'other','web'),
(UUID(),'expense', @food,        530.00,'BDT','Lunch',                            'Fakruddin',       DATE_FORMAT(CURDATE(), '%Y-%m-02'), DATE_FORMAT(CURDATE(), '%Y-%m-02'),'cash','web'),
(UUID(),'expense', @food,        890.00,'BDT','Groceries',                        'Shwapno',         DATE_FORMAT(CURDATE(), '%Y-%m-04'), DATE_FORMAT(CURDATE(), '%Y-%m-04'),'cash','web'),
(UUID(),'expense', @food,        160.00,'BDT','Snacks',                           'Local shop',      DATE_FORMAT(CURDATE(), '%Y-%m-06'), DATE_FORMAT(CURDATE(), '%Y-%m-06'),'cash','web'),
(UUID(),'expense', @food,        410.00,'BDT','Dinner takeaway',                  'Foodpanda',       DATE_FORMAT(CURDATE(), '%Y-%m-08'), DATE_FORMAT(CURDATE(), '%Y-%m-08'),'other','web'),
(UUID(),'expense', @trans,       190.00,'BDT','Uber ride',                        'Uber',            DATE_FORMAT(CURDATE(), '%Y-%m-03'), DATE_FORMAT(CURDATE(), '%Y-%m-03'),'other','web'),
(UUID(),'expense', @trans,      1200.00,'BDT','Monthly bus pass',                 'BRTC',            DATE_FORMAT(CURDATE(), '%Y-%m-01'), DATE_FORMAT(CURDATE(), '%Y-%m-01'),'cash','web'),
(UUID(),'expense', @bills,      1720.00,'BDT','Electricity bill',                 'DESCO',           DATE_FORMAT(CURDATE(), '%Y-%m-10'), DATE_FORMAT(CURDATE(), '%Y-%m-10'),'other','web'),
(UUID(),'expense', @bills,       600.00,'BDT','Internet bill',                    'Grameenphone',    DATE_FORMAT(CURDATE(), '%Y-%m-05'), DATE_FORMAT(CURDATE(), '%Y-%m-05'),'other','web'),
(UUID(),'expense', @bills,      8000.00,'BDT','House rent',                       'Landlord',        DATE_FORMAT(CURDATE(), '%Y-%m-01'), DATE_FORMAT(CURDATE(), '%Y-%m-01'),'cash','web'),
(UUID(),'expense', @ent,         500.00,'BDT','Netflix subscription',             'Netflix',         DATE_FORMAT(CURDATE(), '%Y-%m-06'), DATE_FORMAT(CURDATE(), '%Y-%m-06'),'credit_card','web'),
(UUID(),'expense', @health,     1200.00,'BDT','Gym membership',                   'Gold Gym',        DATE_FORMAT(CURDATE(), '%Y-%m-01'), DATE_FORMAT(CURDATE(), '%Y-%m-01'),'credit_card','web'),
(UUID(),'expense', @shop,       1800.00,'BDT','New clothes',                      'Yellow',          DATE_FORMAT(CURDATE(), '%Y-%m-09'), DATE_FORMAT(CURDATE(), '%Y-%m-09'),'credit_card','web');

SELECT CONCAT('Seeded ', COUNT(*), ' entries') AS result FROM entries WHERE source = 'web';
