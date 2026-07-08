-- Clean Architecture migration for Dynamic Pricing Engine

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'viewer' CHECK (role IN ('admin', 'viewer')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS pricing_config (
    id SERIAL PRIMARY KEY,
    base_price_per_hour DOUBLE PRECISION NOT NULL DEFAULT 6250,
    currency VARCHAR(10) NOT NULL DEFAULT 'IDR',
    surge_cap_multiplier DOUBLE PRECISION NOT NULL DEFAULT 2.0,
    demand_multipliers JSONB NOT NULL DEFAULT '{}',
    zone_surge_config JSONB NOT NULL DEFAULT '{}',
    battery_discount_tiers JSONB NOT NULL DEFAULT '{}',
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS pricing_config_history (
    id SERIAL PRIMARY KEY,
    config_id INT REFERENCES pricing_config(id),
    version INT NOT NULL,
    base_price_per_hour DOUBLE PRECISION NOT NULL,
    currency VARCHAR(10) NOT NULL,
    surge_cap_multiplier DOUBLE PRECISION NOT NULL,
    demand_multipliers JSONB NOT NULL,
    zone_surge_config JSONB NOT NULL,
    battery_discount_tiers JSONB NOT NULL,
    changed_by VARCHAR(100),
    changed_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS vehicles (
    id VARCHAR(20) PRIMARY KEY,
    zone VARCHAR(100) NOT NULL,
    soc DOUBLE PRECISION NOT NULL DEFAULT 100.0 CHECK (soc >= 0 AND soc <= 100),
    model VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS audit_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    vehicle_id VARCHAR(20) NOT NULL,
    zone VARCHAR(100) NOT NULL,
    duration_hours INT NOT NULL,
    input_data JSONB NOT NULL,
    factors_applied JSONB NOT NULL,
    final_price DOUBLE PRECISION NOT NULL,
    config_version INT NOT NULL,
    signature TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS zone_utilization (
    zone VARCHAR(100) PRIMARY KEY,
    utilization DOUBLE PRECISION NOT NULL DEFAULT 0.0 CHECK (utilization >= 0 AND utilization <= 100),
    name VARCHAR(100) NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_vehicle ON audit_log(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_audit_timestamp ON audit_log(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_audit_zone ON audit_log(zone);
CREATE INDEX IF NOT EXISTS idx_config_version ON pricing_config(version DESC);
CREATE INDEX IF NOT EXISTS idx_vehicles_zone ON vehicles(zone);

-- Seed data
INSERT INTO pricing_config (base_price_per_hour, currency, surge_cap_multiplier, demand_multipliers, zone_surge_config, battery_discount_tiers)
VALUES (
    6250, 'IDR', 2.0,
    '{"default":1.0,"rules":[{"days":[1,2,3,4,5],"hours":[17,18,19],"multiplier":1.3},{"days":[1,2,3,4,5],"hours":[0,1,2,3,4],"multiplier":0.8},{"days":[6,0],"hours":[9,10,11,12,13,14,15,16,17,18,19,20,21],"multiplier":1.2}]}',
    '{"thresholds":[{"max_utilization":50,"factor":1.0},{"max_utilization":80,"factor":1.2},{"max_utilization":100,"factor":1.5}]}',
    '{"thresholds":[{"max_soc":40,"discount_factor":0.85},{"max_soc":60,"discount_factor":0.92},{"max_soc":80,"discount_factor":0.97},{"max_soc":100,"discount_factor":1.0}]}'
) ON CONFLICT DO NOTHING;

INSERT INTO users (username, password_hash, role)
VALUES ('admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'admin')
ON CONFLICT (username) DO NOTHING;

INSERT INTO vehicles (id, zone, soc, model) VALUES
    ('EV-10001','south-jakarta',85,'Electrum E1'),
    ('EV-10002','south-jakarta',35,'Electrum E1'),
    ('EV-10003','south-jakarta',62,'Electrum E2'),
    ('EV-10004','central-jakarta',78,'Electrum E1'),
    ('EV-10005','central-jakarta',22,'Electrum E2'),
    ('EV-10006','central-jakarta',91,'Electrum E1'),
    ('EV-10007','east-jakarta',45,'Electrum E2'),
    ('EV-10008','east-jakarta',15,'Electrum E1'),
    ('EV-10009','east-jakarta',73,'Electrum E1'),
    ('EV-10010','south-jakarta',55,'Electrum E2')
ON CONFLICT (id) DO NOTHING;

INSERT INTO zone_utilization (zone, utilization, name) VALUES
    ('south-jakarta',85,'South Jakarta'),
    ('central-jakarta',62,'Central Jakarta'),
    ('east-jakarta',45,'East Jakarta'),
    ('west-jakarta',78,'West Jakarta'),
    ('north-jakarta',30,'North Jakarta')
ON CONFLICT (zone) DO NOTHING;
