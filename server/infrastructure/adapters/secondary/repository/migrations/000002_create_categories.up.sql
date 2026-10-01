CREATE TABLE categories (
    id                         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                       VARCHAR(40) NOT NULL,
    description                TEXT NOT NULL DEFAULT '',
    main_category              VARCHAR(40) NOT NULL DEFAULT 'otros',
    is_active                  BOOLEAN NOT NULL DEFAULT TRUE,
    default_unit_of_measure_id UUID REFERENCES units_of_measure(id) ON DELETE SET NULL,

    CONSTRAINT uq_categories_name UNIQUE (name)
);

INSERT INTO categories (name, description, main_category, default_unit_of_measure_id) VALUES
    ('Frutales', 'Frutas yumeros y dulces', 'frutales', (SELECT id FROM units_of_measure WHERE code = 'kg')),
    ('Cítricos', 'Cítricos', 'citrusos', (SELECT id FROM units_of_measure WHERE code = 'qq')),
    ('Otros',    'Otros productos agropecuarios', 'otros', (SELECT id FROM units_of_measure WHERE code = 'unidad'))
ON CONFLICT (name) DO NOTHING;
