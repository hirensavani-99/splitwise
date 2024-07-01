CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL
);


CREATE TABLE IF NOT EXISTS groups (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT
);


CREATE TABLE IF NOT EXISTS group_member (
    id SERIAL PRIMARY KEY,
    group_id INT REFERENCES groups(id),
    user_id INT REFERENCES users(id)
);


CREATE TABLE IF NOT EXIST expense(
    id SERIAL PRIMARY KEY,
    description TEXT,
    amount DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    category VARCHAR(255),
    added_at TIMESTAMP NOT NULL,
    is_recurring BOOLEAN NOT NULL,
    recurring_period VARCHAR(255),
    notes TEXT,
    group_id INT REFERENCES groups(id),
    AddedBY INT REFERENCES users(id)
)