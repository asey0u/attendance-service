CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS employees (
    id SERIAL PRIMARY KEY,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    position TEXT
);

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    login TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    employee_id INT REFERENCES employees(id) ON DELETE CASCADE,
    role_id INT REFERENCES roles(id)
);

CREATE TABLE IF NOT EXISTS attendance (
    id SERIAL PRIMARY KEY,
    employee_id INT REFERENCES employees(id) ON DELETE CASCADE,
    check_in TIMESTAMP NOT NULL,
    check_out TIMESTAMP
);

INSERT INTO roles(name)
VALUES
    ('admin'),
    ('manager'),
    ('employee')
ON CONFLICT (name) DO NOTHING;

INSERT INTO employees(id, first_name, last_name, position)
VALUES
    (1, 'Иван', 'Петров', 'Administrator'),
    (2, 'Ольга', 'Смирнова', 'Manager'),
    (3, 'Сергей', 'Кузнецов', 'Developer'),
    (4, 'Мария', 'Иванова', 'Designer'),
    (5, 'Алексей', 'Соколов', 'QA Engineer'),
    (6, 'Екатерина', 'Новикова', 'Sales'),
    (7, 'Дмитрий', 'Морозов', 'Support')
ON CONFLICT (id) DO NOTHING;

-- admin → admin123
-- manager → manager123
-- user1 → employee1
-- user2 → employee2
-- user3 → employee3
-- user4 → employee4
-- user5 → employee5


INSERT INTO users(login, password, employee_id, role_id)
VALUES
    ('admin', '$2a$10$zXqfplrK4Wf7BwS.Dsk0aOyPdCKqTkfHT9O1kS74esTIQx8xK7xA2', 1, 1),
    ('manager', '$2a$10$WtJ2mCJnUck5CL50ZmpE.etkEq5/Gz1zq75xJRJpjDo9iBrBAgU6C', 2, 2),
    ('user1', '$2a$10$5omUgJTeaDY/rN5IcKIq3uvUlZGfUtyTHQCmVBqSF934hYbn8IRpK', 3, 3),
    ('user2', '$2a$10$vWXCIo/oHrOXFHbXHGwgqeKmDOwdRbyrR4YEStGO0I1qYaeC4r2Hi', 4, 3),
    ('user3', '$2a$10$hQqCg3jz60V9C7iiyR0KMeINemGxp1.yWHAKcfajhcFrtySKHtunG', 5, 3),
    ('user4', '$2a$10$R1NeBkopxEJBVgEwpx7hAuL9BqzZAbrcKiWtBNaKUetelzEf9.miq', 6, 3),
    ('user5', '$2a$10$LS1IUf849uPSkGtL6dZ1cu3AMK7v1ejzbYM28x4NuyXjYm6zseOqO', 7, 3)
ON CONFLICT (login) DO NOTHING;

INSERT INTO attendance(id, employee_id, check_in, check_out)
VALUES
    (1, 1, '2026-05-07 09:00:00', '2026-05-07 18:00:00'),
    (2, 2, '2026-05-07 09:30:00', '2026-05-07 18:30:00'),
    (3, 3, '2026-05-07 08:55:00', '2026-05-07 17:15:00'),
    (4, 4, '2026-05-07 09:10:00', '2026-05-07 17:50:00'),
    (5, 5, '2026-05-07 10:00:00', '2026-05-07 19:00:00'),
    (6, 6, '2026-05-07 09:05:00', NULL),
    (7, 7, '2026-05-07 09:20:00', '2026-05-07 17:40:00')
ON CONFLICT (id) DO NOTHING;

SELECT setval(pg_get_serial_sequence('employees','id'), COALESCE((SELECT MAX(id) FROM employees), 1));
SELECT setval(pg_get_serial_sequence('users','id'), COALESCE((SELECT MAX(id) FROM users), 1));
SELECT setval(pg_get_serial_sequence('attendance','id'), COALESCE((SELECT MAX(id) FROM attendance), 1));