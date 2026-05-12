-- Migration: Create bathroom_reviews table
-- Date: 2026-05-12
-- Description: Create tables for ratings and reviews system

-- Tabela de ratings e reviews
CREATE TABLE IF NOT EXISTS bathroom_reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bathroom_id UUID NOT NULL REFERENCES bathrooms(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    title VARCHAR(100),
    comment TEXT CHECK (LENGTH(comment) <= 500),
    cleanliness_rating INTEGER CHECK (cleanliness_rating IS NULL OR (cleanliness_rating >= 1 AND cleanliness_rating <= 5)),
    accessibility_rating INTEGER CHECK (accessibility_rating IS NULL OR (accessibility_rating >= 1 AND accessibility_rating <= 5)),
    spaciousness_rating INTEGER CHECK (spaciousness_rating IS NULL OR (spaciousness_rating >= 1 AND spaciousness_rating <= 5)),
    helpful_count INTEGER DEFAULT 0,
    unhelpful_count INTEGER DEFAULT 0,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(bathroom_id, user_id)
);

-- Tabela de fotos de reviews
CREATE TABLE IF NOT EXISTS review_photos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    review_id UUID NOT NULL REFERENCES bathroom_reviews(id) ON DELETE CASCADE,
    photo_url VARCHAR(500) NOT NULL,
    uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabela de helpful votes
CREATE TABLE IF NOT EXISTS review_helpful_votes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    review_id UUID NOT NULL REFERENCES bathroom_reviews(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_helpful BOOLEAN NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(review_id, user_id)
);

-- View para estatísticas de ratings
CREATE OR REPLACE VIEW bathroom_rating_stats AS
SELECT 
    bathroom_id,
    COUNT(*) as total_reviews,
    ROUND(AVG(rating)::NUMERIC, 2) as average_rating,
    ROUND(AVG(COALESCE(cleanliness_rating, 0))::NUMERIC, 2) as avg_cleanliness,
    ROUND(AVG(COALESCE(accessibility_rating, 0))::NUMERIC, 2) as avg_accessibility,
    ROUND(AVG(COALESCE(spaciousness_rating, 0))::NUMERIC, 2) as avg_spaciousness
FROM bathroom_reviews
WHERE status = 'approved'
GROUP BY bathroom_id;

-- Índices para performance
CREATE INDEX IF NOT EXISTS idx_bathroom_reviews_bathroom_id ON bathroom_reviews(bathroom_id);
CREATE INDEX IF NOT EXISTS idx_bathroom_reviews_user_id ON bathroom_reviews(user_id);
CREATE INDEX IF NOT EXISTS idx_bathroom_reviews_created_at ON bathroom_reviews(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_bathroom_reviews_status ON bathroom_reviews(status);
CREATE INDEX IF NOT EXISTS idx_review_photos_review_id ON review_photos(review_id);
CREATE INDEX IF NOT EXISTS idx_review_helpful_votes_review_id ON review_helpful_votes(review_id);
CREATE INDEX IF NOT EXISTS idx_review_helpful_votes_user_id ON review_helpful_votes(user_id);
