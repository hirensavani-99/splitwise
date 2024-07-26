CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL
);


CREATE TABLE IF NOT EXISTS groups (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    simplify_debt BOOLEAN DEFAULT FALSE
);


CREATE TABLE IF NOT EXISTS group_member (
    id SERIAL PRIMARY KEY,
    group_id INT REFERENCES groups(id),
    user_id INT REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_group_member_userid ON group_member(user_id);

CREATE TABLE IF NOT EXISTS expense(
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
    added_by INT REFERENCES users(id),
    add_to json

);
CREATE INDEX IF NOT EXISTS idx_expense_group_id ON expense(group_id);

CREATE TABLE IF NOT EXISTS Wallets (
    user_id INT PRIMARY KEY,
    Balance DECIMAL(10, 2) NOT NULL,
    Currency VARCHAR(10) NOT NULL,
    CreatedAt TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UpdatedAt TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES Users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS Balances(
            id SERIAL PRIMARY KEY,
            from_user_id INT NOT NULL,
            to_user_id INT NOT NULL,
            group_id INT NOT NULL,
            amount FLOAT NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (from_user_id) REFERENCES users(id),
            FOREIGN kEY (to_user_id) REFERENCES Users(id),
            FOREIGN KEY (group_id) REFERENCES groups(id)

);

CREATE TABLE IF NOT EXISTS comments(
    id SERIAL PRIMARY KEY,
    expense_id int NOT NULL,
    user_id int NOT NULL,
    content text NOT null,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (expense_id) REFERENCES expense(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);