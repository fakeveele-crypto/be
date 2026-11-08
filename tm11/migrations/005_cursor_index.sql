CREATE INDEX IF NOT EXISTS users_created_at_id_desc_idx
    ON users (created_at DESC, id DESC);

INSERT INTO permissions (name, description) VALUES
    ('user:list', 'Melihat daftar seluruh user')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_name, permission_name) VALUES
    ('admin', 'user:list'),
    ('staff', 'user:list')
ON CONFLICT (role_name, permission_name) DO NOTHING;