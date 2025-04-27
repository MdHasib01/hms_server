CREATE TABLE doctors (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    specialization VARCHAR(100),
    license_number VARCHAR(100) UNIQUE
);