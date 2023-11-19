\c authdb

CREATE SEQUENCE user_id_seq
    START WITH 578
    INCREMENT BY 21;

CREATE TABLE countries (
    code VARCHAR(2) PRIMARY KEY,
    country VARCHAR(60),
    dialcode VARCHAR(4)
);

CREATE TABLE users (
    id INTEGER PRIMARY KEY DEFAULT nextval('user_id_seq'),
    username VARCHAR(50) UNIQUE NOT NULL,
    passwordHash VARCHAR(300) NOT NULL,
    first_name VARCHAR(60),
    last_name VARCHAR(60),
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(16), -- E.164 Format i.e. +15551231234
        -- use countries table to get phone code
    country VARCHAR DEFAULT 'XX', -- Foreign Key
    is_superuser BOOLEAN DEFAULT FALSE NOT NULL,
    is_staff BOOLEAN DEFAULT FALSE NOT NULL,
    is_active BOOLEAN DEFAULT TRUE NOT NULL,
    date_joined TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    last_login TIMESTAMP WITH TIME ZONE,
    session_id VARCHAR(255),
    
    CONSTRAINT phone_requires_country CHECK (
        (phone IS NOT NULL AND country IS NOT NULL) OR
        (phone IS NULL)
    ),
    FOREIGN KEY (country) REFERENCES countries(code)
);

INSERT INTO countries ( code, country, dialcode ) VALUES
('XX', 'No Country Specified', ''),
('US', 'United States', '+1'),
('GB', 'United Kingdom', '+44'),
('CA', 'Canada', '+1'),
('AU', 'Austrailia', '+61'),
('NZ', 'New Zealand', '+64'),
('JP', 'Japan', '+81'),
('DE', 'Germany', '49'),
('FR', 'France', '+33'),
('UA', 'Ukraine', '+380'),
('MX', 'Mexico', '+52'),
('RU', 'Russia', '+7');


