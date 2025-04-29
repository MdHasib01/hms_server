CREATE TABLE availability (
    doctor_id UUID REFERENCES doctors(user_id),
    available_day VARCHAR(10) NOT NULL, -- like 'Monday'
    starts_at TIME NOT NULL,
    ends_at TIME NOT NULL,
    PRIMARY KEY (doctor_id, available_day, starts_at)
);