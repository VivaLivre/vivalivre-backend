-- Insert test bathrooms with PostGIS locations
-- São Paulo coordinates as reference

INSERT INTO bathrooms (name, address, location, is_accessible, created_at) VALUES
  ('Banheiro Adaptado Centro', 'Av. Paulista, 1000 - São Paulo', ST_SetSRID(ST_MakePoint(-46.6558, -23.5615), 4326), true, NOW()),
  ('Banheiro Acessível Pinheiros', 'Rua Bandeira, 500 - São Paulo', ST_SetSRID(ST_MakePoint(-46.7038, -23.5505), 4326), true, NOW()),
  ('Banheiro Vila Mariana', 'Av. Imirim, 200 - São Paulo', ST_SetSRID(ST_MakePoint(-46.6200, -23.5800), 4326), true, NOW()),
  ('Banheiro Consolação', 'Rua Augusta, 1500 - São Paulo', ST_SetSRID(ST_MakePoint(-46.6650, -23.5450), 4326), false, NOW()),
  ('Banheiro Higienópolis', 'Av. Higienópolis, 800 - São Paulo', ST_SetSRID(ST_MakePoint(-46.6450, -23.5350), 4326), true, NOW()),
  ('Banheiro Liberdade', 'Rua Galvão Bueno, 300 - São Paulo', ST_SetSRID(ST_MakePoint(-46.6350, -23.5650), 4326), true, NOW()),
  ('Banheiro Bela Vista', 'Av. Paulista, 2000 - São Paulo', ST_SetSRID(ST_MakePoint(-46.6600, -23.5550), 4326), false, NOW()),
  ('Banheiro Jardins', 'Rua Oscar Freire, 500 - São Paulo', ST_SetSRID(ST_MakePoint(-46.6700, -23.5650), 4326), true, NOW());

-- Verify insertion
SELECT COUNT(*) as total_bathrooms FROM bathrooms;
SELECT id, name, address, ST_Y(location::geometry) as latitude, ST_X(location::geometry) as longitude, is_accessible FROM bathrooms;
