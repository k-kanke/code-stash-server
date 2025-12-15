INSERT INTO oauth_clients (id, name, secret, grant_types)
VALUES (
  '7d8b1e7d-8c8d-4c7e-9f4a-2f0afc1a0f01',
  'CLI Device Client',
  'cli-device-secret',
  ARRAY['device_code']
)
ON CONFLICT (id) DO NOTHING;
