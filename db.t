CREATE TYPE user_role AS ENUM ('doctor', 'patient', 'receptionist');

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    full_name VARCHAR(255),
    role user_role NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CUTTENT TIMESTAMP
);



-- Doctor table
CREATE TABLE doctors (
    user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    specialization VARCHAR(100),
    license_number VARCHAR(100) UNIQUE
);

-- Patient table
CREATE TABLE patients (
    user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    date_of_birth DATE,
    gender VARCHAR(10)
);

-- Receptionist table
CREATE TABLE receptionists (
    user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    shift_time VARCHAR(100)
);


CREATE OR REPLACE FUNCTION check_single_role()
RETURNS TRIGGER AS $$
DECLARE
    role_count INT;
BEGIN
    SELECT 
        (SELECT COUNT(*) FROM doctors WHERE user_id = NEW.user_id) +
        (SELECT COUNT(*) FROM patients WHERE user_id = NEW.user_id) +
        (SELECT COUNT(*) FROM receptionists WHERE user_id = NEW.user_id)
    INTO role_count;

    IF role_count > 1 THEN
        RAISE EXCEPTION 'A user can only have one role';
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


CREATE TRIGGER doctor_role_check
AFTER INSERT OR UPDATE ON doctors
FOR EACH ROW EXECUTE FUNCTION check_single_role();

CREATE TRIGGER patient_role_check
AFTER INSERT OR UPDATE ON patients
FOR EACH ROW EXECUTE FUNCTION check_single_role();

CREATE TRIGGER receptionist_role_check
AFTER INSERT OR UPDATE ON receptionists
FOR EACH ROW EXECUTE FUNCTION check_single_role();



-- Insert into users table
INSERT INTO users (email, password_hash, full_name, role)
VALUES ('dr.john@example.com', 'hashed_password_123', 'Dr. John Doe', 'doctor')
RETURNING id;
-- Insert into doctors table
INSERT INTO doctors (user_id, specialization, license_number)
VALUES (1, 'Cardiology', 'DOC123456');

-- Insert into users table
INSERT INTO users (email, password_hash, full_name, role)
VALUES ('jane.doe@example.com', 'hashed_password_456', 'Jane Doe', 'patient')
RETURNING id;

-- Insert into patients table
INSERT INTO patients (user_id, date_of_birth, gender)
VALUES (2, '1990-05-14', 'Female');

-- Insert into users table
INSERT INTO users (email, password_hash, full_name, role)
VALUES ('mark.smith@example.com', 'hashed_password_789', 'Mark Smith', 'receptionist')
RETURNING id;

-- Insert into receptionists table
INSERT INTO receptionists (user_id, shift_time)
VALUES (3, 'Morning Shift: 8 AM - 2 PM');

CREATE TABLE doctor_availability (
    id SERIAL PRIMARY KEY,
    doctor_id INTEGER REFERENCES doctors(user_id) ON DELETE CASCADE,
    day VARCHAR(20) NOT NULL, -- e.g. 'Monday'
    start_time TIME NOT NULL,
    end_time TIME NOT NULL
);


-- Lets say doctor_id = 1

INSERT INTO doctor_availability (doctor_id, day, start_time, end_time)
VALUES
(1, 'Monday', '09:00', '12:00'),
(1, 'Wednesday', '14:00', '18:00'),
(1, 'Friday', '10:00', '13:00');


SELECT 
    d.user_id,
    u.full_name,
 	u.email,
    d.specialization,
    ARRAY_AGG(
        json_build_object(
            'day', da.day,
            'start_time', da.start_time,
            'end_time', da.end_time
        )
    ) AS availability
FROM doctors d
JOIN users u ON d.user_id = u.id
LEFT JOIN doctor_availability da ON da.doctor_id = d.user_id
GROUP BY d.user_id, u.full_name, u.email;
