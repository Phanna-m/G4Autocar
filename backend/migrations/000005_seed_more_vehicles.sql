-- Add individually named vehicles and e-bikes to the public catalog.
INSERT INTO products (
    id, category_id, sku, name, brand, model, year, base_price,
    condition, description, images, specs
)
VALUES
(
    '018e0002-0000-0000-0000-000000000001',
    'ca111111-1111-1111-1111-111111111111',
    'HND-CRV-HEV-2026',
    'Honda CR-V e:HEV',
    'Honda',
    'CR-V e:HEV',
    2026,
    42000.00,
    'NEW',
    'Comfortable hybrid SUV with a spacious cabin and efficient two-motor powertrain.',
    '["https://images.unsplash.com/photo-1533473359331-0135ef1b58bf?auto=format&fit=crop&q=80&w=800"]'::jsonb,
    '{"Powertrain":"2.0L Hybrid","Drivetrain":"AWD","Seating":"5"}'::jsonb
),
(
    '018e0002-0000-0000-0000-000000000002',
    'ca111111-1111-1111-1111-111111111111',
    'TYT-RAV4-HEV-2026',
    'Toyota RAV4 Hybrid',
    'Toyota',
    'RAV4 Hybrid',
    2026,
    38700.00,
    'NEW',
    'Versatile hybrid crossover with electronic all-wheel drive and everyday practicality.',
    '["https://tmna.aemassets.toyota.com/is/image/toyota/toyota/jellies/relative/2026/rav4/base.png?bg=fff&fmt=webp&qlt=90&wid=600"]'::jsonb,
    '{"Powertrain":"2.5L Hybrid","Drivetrain":"Electronic AWD","Seating":"5"}'::jsonb
),
(
    '018e0002-0000-0000-0000-000000000003',
    'ca222222-2222-2222-2222-222222222222',
    'TYT-FORTUNER-2026',
    'Toyota Fortuner 2.8 Legender',
    'Toyota',
    'Fortuner 2.8 Legender',
    2026,
    79500.00,
    'NEW',
    'Seven-seat body-on-frame SUV with a turbo-diesel engine and four-wheel drive.',
    '["https://images.unsplash.com/photo-1551830820-330a71b99659?auto=format&fit=crop&q=80&w=800"]'::jsonb,
    '{"Engine":"2.8L Turbo Diesel","Drivetrain":"4WD","Seating":"7"}'::jsonb
),
(
    '018e0002-0000-0000-0000-000000000004',
    'ca222222-2222-2222-2222-222222222222',
    'HND-CIVIC-RS-2026',
    'Honda Civic RS Turbo',
    'Honda',
    'Civic RS Turbo',
    2026,
    37000.00,
    'NEW',
    'Sport-focused sedan with a turbocharged engine and a modern driver-oriented cabin.',
    '["https://images.unsplash.com/photo-1503376780353-7e6692767b70?auto=format&fit=crop&q=80&w=800"]'::jsonb,
    '{"Engine":"1.5L Turbo","Transmission":"CVT","Seating":"5"}'::jsonb
),
(
    '018e0002-0000-0000-0000-000000000005',
    'ca333333-3333-3333-3333-333333333333',
    'GNT-EXPLORE-E2-2026',
    'Giant Explore E+ 2',
    'Giant',
    'Explore E+ 2',
    2026,
    3500.00,
    'NEW',
    'All-round electric trekking bicycle for city commutes and weekend rides.',
    '["https://images.unsplash.com/photo-1571068316344-75bc76f77890?auto=format&fit=crop&q=80&w=800"]'::jsonb,
    '{"Motor":"SyncDrive Sport","Range":"Up to 150 km","Brakes":"Hydraulic disc"}'::jsonb
),
(
    '018e0002-0000-0000-0000-000000000006',
    'ca333333-3333-3333-3333-333333333333',
    'TRK-MARLIN-PLUS-2026',
    'Trek Marlin+ 6',
    'Trek',
    'Marlin+ 6',
    2026,
    4200.00,
    'NEW',
    'Electric hardtail mountain bike built for trails and everyday adventure.',
    '["https://images.unsplash.com/photo-1485965120184-e220f721d03e?auto=format&fit=crop&q=80&w=800"]'::jsonb,
    '{"Motor":"Bosch Active Line","Battery":"400 Wh","Brakes":"Hydraulic disc"}'::jsonb
),
(
    '018e0002-0000-0000-0000-000000000007',
    'ca333333-3333-3333-3333-333333333333',
    'SPC-TURBO-VADO-2026',
    'Specialized Turbo Vado 4.0',
    'Specialized',
    'Turbo Vado 4.0',
    2026,
    4000.00,
    'NEW',
    'Comfortable urban electric bicycle with integrated lights and practical commuter features.',
    '[]'::jsonb,
    '{"Motor":"Specialized 2.0","Range":"Up to 120 km","Brakes":"Hydraulic disc"}'::jsonb
)
ON CONFLICT (sku) DO NOTHING;
