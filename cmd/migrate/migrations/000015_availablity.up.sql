
CREATE TABLE availability (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    doctor_id UUID NOT NULL,
    available_day VARCHAR(10) NOT NULL,
    starts_at TIME NOT NULL,
    ends_at TIME NOT NULL,
    FOREIGN KEY (doctor_id) REFERENCES doctors(user_id) ON DELETE CASCADE
);