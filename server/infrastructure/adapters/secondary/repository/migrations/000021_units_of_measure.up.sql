CREATE TABLE units_of_measure (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code       VARCHAR(20) NOT NULL UNIQUE,
    name       VARCHAR(60) NOT NULL,
    is_active  BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO units_of_measure (code, name) VALUES
    ('kg',      'Kilogramo'),
    ('qq',      'Quintal'),
    ('lb',      'Libra'),
    ('tn',      'Tonelada'),
    ('unidad',  'Unidad'),
    ('docena',  'Docena');
