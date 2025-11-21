CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Các trường ban đầu
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(255) NOT NULL,
    from_date TIMESTAMP WITH TIME ZONE NOT NULL,
    to_date TIMESTAMP WITH TIME ZONE NOT NULL,
    
    -- Các trường bổ sung theo yêu cầu của bạn
    brand_name VARCHAR(255) NOT NULL, -- Tên thương hiệu của bạn
    competitor_names TEXT[], -- Mảng tên các đối thủ
    brand_keywords TEXT[] NOT NULL, -- Array các từ khóa của thương hiệu bạn
    competitor_keywords_map JSONB, -- Map/Từ điển lưu trữ từ khóa của từng đối thủ
    -- (Ví dụ: '{"Competitor A": ["kwA1", "kwA2"], "Competitor B": ["kwB1"]}')
    
    -- Các trường quan hệ và quản lý thời gian
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);