#!/usr/bin/env python3
"""
Generate 6 months of realistic Bangladesh middle-class daily expense data.
Each month produces 60-90 entries covering daily life: rickshaw, bazar,
tea stalls, office lunch, rent, bills, mobile recharge, medicine, etc.

Usage:
    python3 scripts/generate_seed.py | mysql -u root -proot1234 llmexpensetracker
"""

import random
import uuid
from datetime import date, timedelta
from calendar import monthrange

random.seed(42)  # reproducible

# ── helpers ──────────────────────────────────────────────────────────────────
def uid(): return str(uuid.uuid4())

def days_in(y, m): return monthrange(y, m)[1]

def rand_day(y, m, skip=None):
    skip = skip or []
    candidates = [d for d in range(1, days_in(y,m)+1) if d not in skip]
    return random.choice(candidates)

def row(y, m, d, typ, cat, amount, desc, merchant, method='cash'):
    dt = f"{y}-{m:02d}-{d:02d}"
    return (
        f"(UUID(),'{typ}',{cat},{amount:.2f},'BDT',"
        f"'{desc}','{merchant}','{dt}','{dt}','{method}','web')"
    )

# ── 6-month window ending this month ─────────────────────────────────────────
today = date.today()
months = []
for i in range(5, -1, -1):
    y = today.year
    m = today.month - i
    while m <= 0:
        m += 12; y -= 1
    months.append((y, m))

# ── category SQL vars ─────────────────────────────────────────────────────────
CAT = {
    'food':      "@food",
    'trans':     "@trans",
    'shop':      "@shop",
    'ent':       "@ent",
    'health':    "@health",
    'bills':     "@bills",
    'edu':       "@edu",
    'salary':    "@salary",
    'freelance': "@freelance",
    'other':     "@other",
}

# ── realistic item pools ──────────────────────────────────────────────────────

TEA = [
    ("Cha o singara", "Corner tea stall", 15, 30),
    ("Morning tea", "Local cha stall", 10, 20),
    ("Tea with biscuit", "Road side stall", 12, 25),
    ("Afternoon cha", "Office canteen", 10, 20),
    ("Tea and samosa", "Stall near office", 20, 35),
]

LUNCH = [
    ("Dal bhat lunch", "Local restaurant", 70, 120),
    ("Biriyani lunch", "Haji Biriyani", 150, 200),
    ("Khichuri", "Canteen", 60, 90),
    ("Ruti and bhorta", "Small restaurant", 50, 80),
    ("Chicken curry lunch", "Office area restaurant", 100, 160),
    ("Fish curry and rice", "Local hotel", 80, 130),
    ("Beef kala bhuna", "Purbani Hotel", 130, 180),
    ("Doi bora and fuchka", "Street food", 40, 70),
    ("Chotpoti", "Street vendor", 30, 60),
    ("Shawarma", "Fast food", 120, 180),
    ("Office lunch", "Canteen", 70, 110),
    ("Lunch set", "Star Kabab", 100, 150),
]

DINNER = [
    ("Family dinner rice and fish", "Home market items", 250, 450),
    ("Takeaway biryani", "Kacchi Bhai", 350, 550),
    ("Dinner at local hotel", "Fakruddin", 300, 500),
    ("Evening snacks and dinner items", "Local bazar", 200, 350),
    ("Doi puri and halim", "Street vendor", 60, 120),
]

BREAKFAST = [
    ("Paratha and egg", "Breakfast stall", 40, 70),
    ("Ruti and halwa", "Local stall", 30, 55),
    ("Puri and alur dom", "Morning shop", 35, 60),
    ("Bread butter and egg", "Grocery", 50, 90),
    ("Luchi and chole", "Local restaurant", 50, 80),
]

BAZAR = [
    ("Shobji bazar - vegetables", "Local bazar", 150, 400),
    ("Mach bazar - fish", "Local fish market", 250, 600),
    ("Mangsho - beef or mutton", "Local butcher", 400, 900),
    ("Daal and moshla shopping", "Grocery shop", 200, 450),
    ("Rice and atta", "Grocery store", 300, 700),
    ("Weekly bazar shopping", "Karwan Bazar", 500, 1200),
    ("Vegetables and eggs", "Nearby shop", 200, 380),
    ("Onion, garlic, ginger", "Local shop", 80, 200),
    ("Mustard oil and spices", "Grocery", 150, 350),
    ("Lentils and pulses", "Shop", 150, 300),
    ("Milk and dairy", "Aarong Dairy", 100, 220),
    ("Frozen items", "Shwapno", 200, 500),
    ("Monthly grocery stock", "Meena Bazar", 800, 1800),
]

RICKSHAW = [
    ("Rickshaw to bazar", "Local rickshaw", 20, 50),
    ("Rickshaw to bus stand", "Rickshaw", 20, 40),
    ("Rickshaw from office", "Rickshaw", 30, 60),
    ("Baby taxi to market", "Baby taxi", 50, 100),
    ("Rickshaw for kids school", "Rickshaw", 30, 60),
    ("Rickshaw round trip", "Rickshaw", 40, 80),
]

BUS_CNG = [
    ("Bus fare to office", "Local bus", 20, 40),
    ("CNG to Gulshan", "CNG", 80, 150),
    ("Bus fare", "BRTC", 15, 30),
    ("Local bus fare", "City bus", 10, 25),
    ("Uber to meeting", "Uber", 150, 350),
    ("Pathao bike", "Pathao", 60, 150),
    ("CNG auto", "CNG", 70, 130),
    ("Bus and rickshaw commute", "Local transport", 50, 90),
    ("Leguna fare", "Leguna", 15, 25),
]

MOBILE = [
    ("GP mobile recharge", "Grameenphone", 50, 200),
    ("Robi recharge", "Robi", 50, 150),
    ("Banglalink recharge", "Banglalink", 50, 100),
    ("Mobile data pack", "Grameenphone", 100, 250),
    ("Internet data recharge", "Robi", 80, 200),
]

MEDICINE = [
    ("Paracetamol and medicine", "Lazz Pharma", 80, 200),
    ("Stomach medicine", "Local pharmacy", 60, 150),
    ("Vitamins and supplements", "Pharmacy", 150, 400),
    ("Allergy medicine", "Pharmacy", 100, 250),
    ("ORS and saline", "Pharmacy", 50, 120),
    ("Antacid medicine", "Drug store", 60, 130),
    ("Cough syrup", "Pharmacy", 80, 180),
]

DOCTOR = [
    ("Doctor consultation", "Square Hospital", 600, 1500),
    ("Chamber visit - MBBS", "Local chamber", 400, 800),
    ("Child doctor visit", "Ibn Sina", 600, 1200),
    ("Eye checkup", "Eye hospital", 500, 1000),
    ("Dental visit", "Popular Dental", 800, 2000),
]

CLOTHING = [
    ("Punjabi", "Aarong", 800, 2500),
    ("Sari for wife", "Tangail Saree Kutir", 600, 2000),
    ("School uniform", "Local tailor", 400, 900),
    ("Kids clothes", "Yellow", 500, 1200),
    ("Shirt and pant", "Richman", 600, 1800),
    ("Sandal/chappal", "Bata", 300, 800),
    ("Undergarments and socks", "Local shop", 150, 400),
    ("Lungi", "Local shop", 200, 500),
]

UTILITY_BILLS = [
    ("DESCO electricity bill", "DESCO", 1200, 2500),
    ("Titas gas bill", "Titas Gas", 950, 1100),
    ("WASA water bill", "WASA", 400, 600),
    ("Internet bill", "Grameenphone Home", 600, 700),
    ("Satellite TV bill", "BD Cable", 250, 400),
]

KIDS = [
    ("School tuition fee", "School", 2500, 5000),
    ("Kids school van fee", "Van driver", 1000, 1500),
    ("School books and copies", "Nilkhet", 300, 800),
    ("Kids tiffin money", "School", 500, 1000),
    ("Coaching class fee", "Coaching center", 1500, 3000),
]

SALON = [
    ("Haircut", "Local salon", 60, 150),
    ("Shave and haircut", "Barber", 50, 100),
    ("Hair cut for kids", "Local salon", 50, 80),
]

MISC = [
    ("Newspaper subscription", "Prothom Alo", 350, 420),
    ("Prayer items - attar, tasbih", "Islamic shop", 100, 300),
    ("Household cleaning items", "Shop", 150, 400),
    ("Utensils repair", "Local repair", 80, 200),
    ("Umbrella", "Shop", 150, 350),
    ("Postage and courier", "Sundarban courier", 80, 200),
    ("Mobile phone repair", "Local shop", 200, 800),
    ("Pen drive and stationery", "Stationery shop", 100, 300),
    ("Cigarette", "Shop", 10, 30),
    ("Betel leaf - paan", "Paan shop", 10, 30),
    ("Charitable donation - mosque", "Mosque", 100, 500),
    ("Charitable donation - poor", "Individual", 50, 200),
    ("Parents money send bKash", "bKash", 2000, 5000),
    ("Birthday cake", "Bakery", 300, 800),
    ("Photography", "Studio", 500, 1500),
]

ENTERTAINMENT = [
    ("Bkash recharge for OTT", "bKash", 300, 600),
    ("Hoichoi subscription", "Hoichoi", 200, 350),
    ("Netflix", "Netflix", 400, 500),
    ("Cinema ticket", "Star Cineplex", 200, 500),
    ("YouTube premium", "Google", 189, 189),
    ("Tea and adda outing", "Tea house", 100, 250),
    ("Family picnic expense", "Local spot", 500, 2000),
    ("Eid fair entry", "Fair", 100, 300),
]

def pick(pool, price_override=None):
    item = random.choice(pool)
    desc, merchant = item[0], item[1]
    lo, hi = item[2], item[3]
    amount = price_override if price_override else round(random.uniform(lo, hi) / 5) * 5
    if amount < 1: amount = lo
    return desc, merchant, amount


# ── generate month data ───────────────────────────────────────────────────────
def generate_month(y, m, salary, extra_income=None):
    rows = []
    used_days = set()
    last_day = days_in(y, m)

    def add(d, typ, cat, amount, desc, merchant, method='cash'):
        rows.append(row(y, m, d, typ, cat, amount, desc, merchant, method))

    # ── Income ────────────────────────────────────────────────────────────────
    add(1, 'income', CAT['salary'], salary, 'Monthly salary', 'Employer', 'other')
    if extra_income:
        for amt, desc, day in extra_income:
            add(day, 'income', CAT['freelance'], amt, desc, 'Client', 'other')

    # ── Fixed monthly bills (1st week) ───────────────────────────────────────
    add(1,  'expense', CAT['bills'], random.randint(14000,18000), 'House rent', 'Landlord', 'cash')
    add(random.randint(8,12), 'expense', CAT['bills'], round(random.uniform(1200,2400)/10)*10, 'DESCO electricity bill', 'DESCO', 'other')
    add(random.randint(5,8),  'expense', CAT['bills'], random.randint(950,1100), 'Titas gas bill', 'Titas Gas', 'cash')
    add(random.randint(5,7),  'expense', CAT['bills'], random.randint(400,600), 'WASA water bill', 'WASA', 'cash')
    add(random.randint(5,8),  'expense', CAT['bills'], random.randint(600,700), 'Internet bill', 'Grameenphone Home', 'other')
    add(random.randint(3,6),  'expense', CAT['other'], random.randint(350,420), 'Newspaper subscription', 'Prothom Alo', 'cash')
    add(random.randint(5,10), 'expense', CAT['bills'], random.randint(250,400), 'Cable TV bill', 'BD Cable', 'cash')

    # ── School / kids (if applicable) ────────────────────────────────────────
    add(1, 'expense', CAT['edu'], random.randint(3000,5500), 'School tuition fee', 'School', 'cash')
    add(1, 'expense', CAT['trans'], random.randint(1000,1500), 'School van monthly fee', 'Van driver', 'cash')
    if random.random() > 0.5:
        add(random.randint(2,10), 'expense', CAT['edu'], random.randint(1500,3000), 'Coaching class fee', 'Coaching center', 'cash')

    # ── Mobile recharge (2-4 times a month) ──────────────────────────────────
    for _ in range(random.randint(2, 4)):
        d, mer, amt = pick(MOBILE)
        add(rand_day(y,m), 'expense', CAT['bills'], amt, d, mer, 'other')

    # ── Rickshaw/transport (daily-ish: 15-20 entries) ─────────────────────────
    transport_days = random.sample(range(1, last_day+1), random.randint(15, 22))
    for d in transport_days:
        if random.random() < 0.6:
            desc, mer, amt = pick(RICKSHAW)
        else:
            desc, mer, amt = pick(BUS_CNG)
        add(d, 'expense', CAT['trans'], amt, desc, mer, 'cash')

    # ── Tea (almost every day, 20-25 entries) ────────────────────────────────
    tea_days = random.sample(range(1, last_day+1), random.randint(20, 26))
    for d in tea_days:
        desc, mer, amt = pick(TEA)
        add(d, 'expense', CAT['food'], amt, desc, mer, 'cash')

    # ── Office lunch / midday meal (18-22 entries) ────────────────────────────
    lunch_days = random.sample(range(1, last_day+1), random.randint(18, 24))
    for d in lunch_days:
        desc, mer, amt = pick(LUNCH)
        add(d, 'expense', CAT['food'], amt, desc, mer, 'cash')

    # ── Breakfast bought (6-10 times) ─────────────────────────────────────────
    for _ in range(random.randint(5, 10)):
        desc, mer, amt = pick(BREAKFAST)
        add(rand_day(y,m), 'expense', CAT['food'], amt, desc, mer, 'cash')

    # ── Bazar / local market (2-3 times a week = 8-12 entries) ───────────────
    bazar_days = random.sample(range(1, last_day+1), random.randint(8, 13))
    for d in bazar_days:
        desc, mer, amt = pick(BAZAR)
        add(d, 'expense', CAT['food'], amt, desc, mer, 'cash')

    # ── Evening dinner / takeaway (4-7 times) ────────────────────────────────
    for _ in range(random.randint(4, 7)):
        desc, mer, amt = pick(DINNER)
        add(rand_day(y,m), 'expense', CAT['food'], amt, desc, mer, 'cash')

    # ── Medicine / pharmacy (3-6 times) ──────────────────────────────────────
    for _ in range(random.randint(3, 6)):
        desc, mer, amt = pick(MEDICINE)
        add(rand_day(y,m), 'expense', CAT['health'], amt, desc, mer, 'cash')

    # ── Doctor (0-2 times) ────────────────────────────────────────────────────
    for _ in range(random.randint(0, 2)):
        desc, mer, amt = pick(DOCTOR)
        method = random.choice(['cash', 'cash', 'other'])
        add(rand_day(y,m), 'expense', CAT['health'], amt, desc, mer, method)

    # ── Haircut / salon (1-2 times) ───────────────────────────────────────────
    for _ in range(random.randint(1, 2)):
        desc, mer, amt = pick(SALON)
        add(rand_day(y,m), 'expense', CAT['other'], amt, desc, mer, 'cash')

    # ── Clothing / shoes (0-2 items) ─────────────────────────────────────────
    for _ in range(random.randint(0, 2)):
        desc, mer, amt = pick(CLOTHING)
        add(rand_day(y,m), 'expense', CAT['shop'], amt, desc, mer, random.choice(['cash','credit_card']))

    # ── Entertainment (1-3) ───────────────────────────────────────────────────
    for _ in range(random.randint(1, 3)):
        desc, mer, amt = pick(ENTERTAINMENT)
        add(rand_day(y,m), 'expense', CAT['ent'], amt, desc, mer, random.choice(['cash','credit_card','other']))

    # ── Miscellaneous (2-5) ───────────────────────────────────────────────────
    for _ in range(random.randint(2, 5)):
        desc, mer, amt = pick(MISC)
        add(rand_day(y,m), 'expense', CAT['other'], amt, desc, mer, 'cash')

    # ── Charity / mosque (1-2) ───────────────────────────────────────────────
    for _ in range(random.randint(1, 2)):
        amt = round(random.randint(50, 500) / 50) * 50
        desc = random.choice(["Mosque donation", "Friday mosque donation", "Poor person help", "Zakat contribution"])
        add(rand_day(y,m), 'expense', CAT['other'], amt, desc, 'Individual/Mosque', 'cash')

    return rows


# ── special months ─────────────────────────────────────────────────────────────
def extra_clothes(y, m):
    """Extra Eid / festival clothing spike"""
    rows = []
    for _ in range(random.randint(3, 6)):
        desc, mer, amt = pick(CLOTHING)
        amt = amt * random.uniform(1.5, 2.5)  # Eid premium
        amt = round(amt / 50) * 50
        rows.append(row(y, m, rand_day(y,m), 'expense', CAT['shop'], amt, desc + ' (Eid)', mer,
                        random.choice(['cash','credit_card'])))
    return rows


# ── main ──────────────────────────────────────────────────────────────────────
print("-- Auto-generated seed: 6 months Bangladesh middle-class expenses")
print("-- Run: python3 scripts/generate_seed.py | mysql -u root -proot1234 llmexpensetracker")
print()
print("SET @food     = (SELECT category_id FROM categories WHERE name = 'Food'          LIMIT 1);")
print("SET @trans    = (SELECT category_id FROM categories WHERE name = 'Transport'     LIMIT 1);")
print("SET @shop     = (SELECT category_id FROM categories WHERE name = 'Shopping'      LIMIT 1);")
print("SET @ent      = (SELECT category_id FROM categories WHERE name = 'Entertainment' LIMIT 1);")
print("SET @health   = (SELECT category_id FROM categories WHERE name = 'Health'        LIMIT 1);")
print("SET @bills    = (SELECT category_id FROM categories WHERE name = 'Bills'         LIMIT 1);")
print("SET @edu      = (SELECT category_id FROM categories WHERE name = 'Education'     LIMIT 1);")
print("SET @salary   = (SELECT category_id FROM categories WHERE name = 'Salary'        LIMIT 1);")
print("SET @freelance= (SELECT category_id FROM categories WHERE name = 'Freelance'     LIMIT 1);")
print("SET @other    = (SELECT category_id FROM categories WHERE name = 'Other'         LIMIT 1);")
print()

all_rows = []

freelance_pools = [
    [(12000, "Logo design project", 10)],
    [(8500, "Banner and poster design", 14), (5000, "Content writing", 22)],
    [(18000, "Mobile app UI design", 12), (6000, "Social media management", 20)],
    [(22000, "Full-stack dev project", 8)],
    [(15000, "API integration", 18), (7000, "SEO consulting", 25)],
    [(10000, "Dashboard UI project", 5), (5500, "Blog writing", 20)],
]

salaries = [60000, 60000, 62000, 62000, 65000, 65000]

for i, (y, m) in enumerate(months):
    extra = freelance_pools[i] if random.random() > 0.3 else []
    rows = generate_month(y, m, salaries[i], extra if extra else None)

    # Eid bump — add to whichever month looks like Eid (month 3 from end feels right)
    if i == 2:
        rows += extra_clothes(y, m)
        # Eid travel
        rows.append(row(y, m, rand_day(y,m), 'expense', CAT['trans'],
                        random.randint(2500,4500), 'Bus ticket home (Eid)', 'Shyamoli Paribahan', 'cash'))
        rows.append(row(y, m, rand_day(y,m), 'expense', CAT['food'],
                        random.randint(800,1500), 'Eid special food and sweets', 'Local shop', 'cash'))
        rows.append(row(y, m, rand_day(y,m), 'expense', CAT['other'],
                        random.randint(1000,3000), 'Eid salami for relatives', 'Family', 'cash'))

    all_rows.extend(rows)

# Write in batches of 50 rows
BATCH = 50
header = "INSERT INTO entries (entry_id, transaction_type, category_id, amount, currency, description, merchant, expense_date, reporting_date, payment_method, source) VALUES"
for start in range(0, len(all_rows), BATCH):
    batch = all_rows[start:start+BATCH]
    print(header)
    print(',\n'.join(batch) + ';')
    print()

print(f"SELECT CONCAT('Seeded ', COUNT(*), ' entries across ', COUNT(DISTINCT DATE_FORMAT(expense_date, '%Y-%m')), ' months') AS result FROM entries WHERE source='web';")
