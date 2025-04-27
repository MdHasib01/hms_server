CREATE TABLE IF NOT EXISTS roles (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL UNIQUE,
  level int NOT NULL DEFAULT 0,
  description TEXT
);

INSERT INTO
  roles (name, description, level)
VALUES
  (
    'patient',
    'A patient can only make appointments',
    1
  );

INSERT INTO
  roles (name, description, level)
VALUES
  (
    'doctor',
    'A doctor can see the appointments of other patients',
    2
  );

INSERT INTO
  roles (name, description, level)
VALUES
  (
    'receptionist',
    'A receptionist can see the appointments of all patients',
    2
  );
INSERT INTO
  roles (name, description, level)
VALUES
  (
    'admin',
    'An admin can see the appointments of all patients',
    3
  );