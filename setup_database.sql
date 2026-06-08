-- Ativar a extensão PostGIS se ainda não estiver ativa
CREATE EXTENSION IF NOT EXISTS postgis;

-- Criar a tabela de utilizadores (caso não exista, já deve existir mas previne erros)
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255),
    avatar_url TEXT,
    height INTEGER,
    weight DOUBLE PRECISION,
    birth_date DATE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Tabela de Banheiros
CREATE TABLE IF NOT EXISTS bathrooms (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    address VARCHAR(255) NOT NULL,
    location GEOMETRY(Point, 4326) NOT NULL,
    is_accessible BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Criar Índice Espacial (crucial para otimização de busca Radius no PostGIS)
CREATE INDEX IF NOT EXISTS idx_bathrooms_location ON bathrooms USING GIST (location);

-- Tabela de Registos de Saúde
CREATE TABLE IF NOT EXISTS health_entries (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    severity VARCHAR(50) NOT NULL DEFAULT 'Leve',
    description TEXT,
    symptoms TEXT[] DEFAULT '{}',
    entry_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Tabela de Reviews de Banheiros
CREATE TABLE IF NOT EXISTS bathroom_reviews (
    id SERIAL PRIMARY KEY,
    bathroom_id INTEGER NOT NULL REFERENCES bathrooms(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    title VARCHAR(100),
    comment TEXT,
    cleanliness_rating INTEGER CHECK (cleanliness_rating >= 1 AND cleanliness_rating <= 5),
    accessibility_rating INTEGER CHECK (accessibility_rating >= 1 AND accessibility_rating <= 5),
    spaciousness_rating INTEGER CHECK (spaciousness_rating >= 1 AND spaciousness_rating <= 5),
    helpful_count INTEGER DEFAULT 0,
    unhelpful_count INTEGER DEFAULT 0,
    status VARCHAR(20) DEFAULT 'approved',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(bathroom_id, user_id)
);

-- Tabela de Votos de Utilidade
CREATE TABLE IF NOT EXISTS review_helpful_votes (
    id SERIAL PRIMARY KEY,
    review_id INTEGER NOT NULL REFERENCES bathroom_reviews(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_helpful BOOLEAN NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(review_id, user_id)
);

-- Índices para performance
CREATE INDEX IF NOT EXISTS idx_reviews_bathroom ON bathroom_reviews(bathroom_id);
CREATE INDEX IF NOT EXISTS idx_reviews_user ON bathroom_reviews(user_id);
CREATE INDEX IF NOT EXISTS idx_reviews_created ON bathroom_reviews(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_helpful_votes_review ON review_helpful_votes(review_id);

-- Inserir dados Mock (Banheiros na região de SP - Próximo das coordenadas -23.66, -46.43)
INSERT INTO bathrooms (name, address, location, is_accessible) VALUES
('Banheiro Público Praça', 'Praça Central, São Paulo', ST_SetSRID(ST_MakePoint(-46.430891, -23.660704), 4326), true),
('Banheiro Shopping ABCD', 'Av. das Américas, São Paulo', ST_SetSRID(ST_MakePoint(-46.425000, -23.665000), 4326), true),
('Banheiro Posto 24h', 'Rodovia Principal, São Paulo', ST_SetSRID(ST_MakePoint(-46.435000, -23.655000), 4326), false),
('Banheiro Parque Municipal', 'Parque da Cidade, São Paulo', ST_SetSRID(ST_MakePoint(-46.438000, -23.668000), 4326), true)
ON CONFLICT DO NOTHING;
