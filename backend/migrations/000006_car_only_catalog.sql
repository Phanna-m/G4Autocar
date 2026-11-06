-- Keep historical orders intact while removing non-vehicle items from the
-- public catalog.
UPDATE products p
SET is_active = FALSE
FROM categories c
WHERE p.category_id = c.id
  AND c.type IN ('BICYCLE', 'SPARE_PART');

UPDATE categories
SET is_active = FALSE
WHERE type IN ('BICYCLE', 'SPARE_PART');

INSERT INTO products (
    id, category_id, sku, name, brand, model, year, base_price,
    condition, description, images, specs
)
VALUES
(
    '018e0003-0000-0000-0000-000000000001',
    'ca222222-2222-2222-2222-222222222222',
    'BMW-M4-COMP-2026',
    'BMW M4 Competition',
    'BMW',
    'M4 Competition',
    2026,
    82400.00,
    'NEW',
    'High-performance luxury coupe with a twin-turbocharged engine and track-focused handling.',
    '["https://images.unsplash.com/photo-1503376780353-7e6692767b70?auto=format&fit=crop&q=80&w=800"]'::jsonb,
    '{"Engine":"Twin-turbo inline-six","Drivetrain":"M xDrive available","Seating":"4"}'::jsonb
),
(
    '018e0003-0000-0000-0000-000000000002',
    'ca111111-1111-1111-1111-111111111111',
    'DODGE-CHARGER-DAYTONA-2026',
    'Dodge Charger Daytona',
    'Dodge',
    'Charger Daytona',
    2026,
    59995.00,
    'NEW',
    'Modern all-electric muscle car with bold styling and performance-focused all-wheel drive.',
    '["https://images.unsplash.com/photo-1560958089-b8a1929cea89?auto=format&fit=crop&q=80&w=800"]'::jsonb,
    '{"Powertrain":"All-electric","Drivetrain":"AWD","Body":"Muscle car"}'::jsonb
),
(
    '018e0003-0000-0000-0000-000000000003',
    'ca222222-2222-2222-2222-222222222222',
    'NISSAN-GTR-NISMO-2026',
    'Nissan GT-R NISMO',
    'Nissan',
    'GT-R NISMO',
    2026,
    220000.00,
    'NEW',
    'Showcase supercar listing inspired by the GT-R NISMO, with a twin-turbo performance setup and all-wheel drive.',
    '["https://images.unsplash.com/photo-1549317661-bd32c8ce0db2?auto=format&fit=crop&q=80&w=800"]'::jsonb,
    '{"Engine":"Twin-turbo V6","Drivetrain":"AWD","Body":"Supercar"}'::jsonb
),
(
    '018e0003-0000-0000-0000-000000000004',
    'ca111111-1111-1111-1111-111111111111',
    'LAMBORGHINI-REVUELTO-2026',
    'Lamborghini Revuelto',
    'Lamborghini',
    'Revuelto',
    2026,
    608000.00,
    'NEW',
    'Hybrid V12 supercar combining dramatic design with high-performance electric assistance.',
    '["https://images.unsplash.com/photo-1583121274602-3e2820c69888?auto=format&fit=crop&q=80&w=800"]'::jsonb,
    '{"Powertrain":"V12 plug-in hybrid","Drivetrain":"AWD","Body":"Supercar"}'::jsonb
),
(
    '018e0003-0000-0000-0000-000000000005',
    'ca222222-2222-2222-2222-222222222222',
    'FORD-MUSTANG-DARK-HORSE-2026',
    'Ford Mustang Dark Horse',
    'Ford',
    'Mustang Dark Horse',
    2026,
    63995.00,
    'NEW',
    'Track-inspired Mustang coupe with a naturally aspirated V8 and rear-wheel drive.',
    '["https://images.unsplash.com/photo-1494976388531-d1058494cdd8?auto=format&fit=crop&q=80&w=800"]'::jsonb,
    '{"Engine":"5.0L V8","Drivetrain":"RWD","Body":"Sports coupe"}'::jsonb
),
(
    '018e0003-0000-0000-0000-000000000006',
    'ca111111-1111-1111-1111-111111111111',
    'AUDI-RS-E-TRON-GT-2026',
    'Audi RS e-tron GT',
    'Audi',
    'RS e-tron GT',
    2026,
    167000.00,
    'NEW',
    'Electric grand-touring sedan with a performance-tuned dual-motor all-wheel-drive system.',
    '["https://images.unsplash.com/photo-1560958089-b8a1929cea89?auto=format&fit=crop&q=80&w=800"]'::jsonb,
    '{"Powertrain":"Dual-motor electric","Drivetrain":"AWD","Body":"Grand tourer"}'::jsonb
),
(
    '018e0003-0000-0000-0000-000000000007',
    'ca222222-2222-2222-2222-222222222222',
    'CHEVROLET-CAMARO-ZL1-2026',
    'Chevrolet Camaro ZL1',
    'Chevrolet',
    'Camaro ZL1',
    2026,
    72995.00,
    'NEW',
    'Showcase muscle-car listing with supercharged V8 performance and aggressive coupe styling.',
    '["https://images.unsplash.com/photo-1552519507-da3b142c6e3d?auto=format&fit=crop&q=80&w=800"]'::jsonb,
    '{"Engine":"Supercharged V8","Drivetrain":"RWD","Body":"Muscle car"}'::jsonb
),
(
    '018e0003-0000-0000-0000-000000000008',
    'ca222222-2222-2222-2222-222222222222',
    'MERCEDES-CLA-300-2026',
    'Mercedes-Benz CLA 300',
    'Mercedes-Benz',
    'CLA 300',
    2026,
    46500.00,
    'NEW',
    'Compact luxury four-door coupe with a refined cabin and modern driver assistance.',
    '["https://images.unsplash.com/photo-1563720223185-11003d516935?auto=format&fit=crop&q=80&w=800"]'::jsonb,
    '{"Engine":"Turbocharged four-cylinder","Drivetrain":"FWD or AWD","Body":"Luxury sedan"}'::jsonb
),
(
    '018e0003-0000-0000-0000-000000000009',
    'ca222222-2222-2222-2222-222222222222',
    'LEXUS-LX-570-2026',
    'Lexus LX 570',
    'Lexus',
    'LX 570',
    2026,
    93000.00,
    'NEW',
    'Full-size luxury SUV showcase with a spacious interior and confident four-wheel-drive styling.',
    '["https://images.unsplash.com/photo-1533473359331-0135ef1b58bf?auto=format&fit=crop&q=80&w=800"]'::jsonb,
    '{"Drivetrain":"4WD","Body":"Luxury SUV","Seating":"7"}'::jsonb
),
(
    '018e0003-0000-0000-0000-000000000010',
    'ca222222-2222-2222-2222-222222222222',
    'CHEVROLET-CORVETTE-Z06-2026',
    'Chevrolet Corvette Z06',
    'Chevrolet',
    'Corvette Z06',
    2026,
    115000.00,
    'NEW',
    'Mid-engine American supercar with a high-revving V8 and track-ready aerodynamics.',
    '["https://images.unsplash.com/photo-1583121274602-3e2820c69888?auto=format&fit=crop&q=80&w=800"]'::jsonb,
    '{"Engine":"Naturally aspirated V8","Drivetrain":"RWD","Body":"Supercar"}'::jsonb
)
ON CONFLICT (sku) DO NOTHING;
