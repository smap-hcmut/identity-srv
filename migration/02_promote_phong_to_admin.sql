-- Promote user to ADMIN role
-- User: phong.dang2212548@hcmut.edu.vn
-- Date: 2026-06-06

SET search_path TO identity;

UPDATE identity.users
SET role_hash = '0zSAZYTmciCygbBTf9Oy3uv4DOZ5ofqp1cIUNJ20uuc=',
    updated_at = CURRENT_TIMESTAMP
WHERE email = 'phong.dang2212548@hcmut.edu.vn';
