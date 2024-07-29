-- Create user table with various data types
CREATE TABLE user (
    id INT PRIMARY KEY,
    name VARCHAR(100),
    email VARCHAR(100),
    profile JSON,
    age INT,
    birth_date DATE,
    is_active BOOLEAN,
    balance DECIMAL(10, 2),
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

-- Insert sample data into user table
INSERT INTO user (id, name, email, profile, age, birth_date, is_active, balance, created_at, updated_at) VALUES
(1, 'Alice', 'alice@example.com', '{"age": 30, "address": {"city": "Wonderland", "street": "123 Rabbit Hole"}}', 30, '1991-04-12', TRUE, 1000.50, '2023-07-20 14:30:00', '2024-07-20 14:30:00'),
(2, 'Bob', 'bob@example.com', '{"age": 25, "address": {"city": "Builderland", "street": "456 Tool Ave"}}', 25, '1996-08-25', FALSE, 750.75, '2023-07-21 15:00:00', '2024-07-21 15:00:00'),
(3, 'Charlie', 'charlie@example.com', '{"age": 35, "address": {"city": "Chocolate Factory", "street": "789 Candy Lane"}}', 35, '1986-12-15', TRUE, 1500.00, '2023-07-22 16:45:00', '2024-07-22 16:45:00');

-- Create book table with additional columns for nested data
CREATE TABLE book (
    id INT PRIMARY KEY,
    title VARCHAR(200),
    user_id INT,
    published_date DATE,
    metadata JSON,
    FOREIGN KEY (user_id) REFERENCES user(id)
);

-- Insert sample data into book table
INSERT INTO book (id, title, user_id, published_date, metadata) VALUES
(1, 'Golang Programming', 1, '2021-01-01', '{"genre": "Technology", "pages": 350, "publisher": {"name": "Tech Books", "location": "USA"}}'),
(2, 'Advanced SQL', 1, '2020-05-15', '{"genre": "Database", "pages": 420, "publisher": {"name": "DB Masters", "location": "UK"}}'),
(3, 'Blockchain Basics', 2, '2019-11-11', '{"genre": "Finance", "pages": 310, "publisher": {"name": "Crypto Books", "location": "Canada"}}'),
(4, 'Python for Beginners', 3, '2022-07-07', '{"genre": "Programming", "pages": 270, "publisher": {"name": "Python Press", "location": "Australia"}}'),
(5, 'Machine Learning', 3, '2018-03-23', '{"genre": "Artificial Intelligence", "pages": 560, "publisher": {"name": "AI Books", "location": "India"}}');

-- Create a view to show users with their books, including JSON fields and other data types
CREATE VIEW user_books AS
SELECT
    u.id AS user_id,
    u.name AS user_name,
    u.email AS user_email,
    u.profile AS user_profile,
    u.age AS user_age,
    u.birth_date AS user_birth_date,
    u.is_active AS user_is_active,
    u.balance AS user_balance,
    u.created_at AS user_created_at,
    u.updated_at AS user_updated_at,
    b.id AS book_id,
    b.title AS book_title,
    b.published_date AS book_published_date,
    b.metadata AS book_metadata
FROM
    user u
LEFT JOIN
    book b ON u.id = b.user_id;

-- Query the view
SELECT * FROM user_books;

CREATE TABLE user_log(
        id INT PRIMARY KEY,
    name VARCHAR(100),
    email VARCHAR(100),
    profile JSON,
    age INT,
    birth_date DATE,
    is_active BOOLEAN,
    balance DECIMAL(10, 2),
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TRIGGER onInsert
AFTER INSERT ON user
FOR EACH ROW
BEGIN
    INSERT INTO user_log (
        id, name, email, profile, age, birth_date, is_active, balance, created_at, updated_at
    ) VALUES (
        NEW.id, NEW.name, NEW.email, NEW.profile, NEW.age, NEW.birth_date, NEW.is_active, NEW.balance, NEW.created_at, NEW.updated_at
    );
END;