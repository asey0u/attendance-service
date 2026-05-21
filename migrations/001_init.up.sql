CREATE TABLE IF NOT EXISTS roles (
    id   SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS departments (
    id         SERIAL PRIMARY KEY,
    name       TEXT UNIQUE NOT NULL,
    manager_id INT NULL
);

CREATE TABLE IF NOT EXISTS employees (
    id            SERIAL PRIMARY KEY,
    first_name    TEXT NOT NULL,
    last_name     TEXT NOT NULL,
    position      TEXT NOT NULL,
    department_id INT NULL REFERENCES departments(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS users (
    id          SERIAL PRIMARY KEY,
    login       TEXT UNIQUE NOT NULL,
    password    TEXT NOT NULL,
    employee_id INT UNIQUE REFERENCES employees(id) ON DELETE CASCADE,
    role_id     INT NOT NULL REFERENCES roles(id)
);

CREATE TABLE IF NOT EXISTS attendance (
    id          SERIAL PRIMARY KEY,
    employee_id INT NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    check_in    TIMESTAMP NOT NULL,
    check_out   TIMESTAMP NULL
);

CREATE TABLE IF NOT EXISTS tickets (
    id          SERIAL PRIMARY KEY,
    employee_id INT NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    ticket_date DATE NOT NULL,
    reason      TEXT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'pending'
                CHECK (status IN ('pending','approved','declined')),
    reviewed_by INT NULL REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMP NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT now()
);

ALTER TABLE departments DROP CONSTRAINT IF EXISTS fk_dept_manager;
ALTER TABLE departments ADD CONSTRAINT fk_dept_manager
    FOREIGN KEY (manager_id) REFERENCES users(id) ON DELETE SET NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_attendance_employee_day
    ON attendance(employee_id, ((check_in)::date));

CREATE INDEX IF NOT EXISTS idx_attendance_employee
    ON attendance(employee_id, check_in);

CREATE INDEX IF NOT EXISTS idx_tickets_status   ON tickets(status);
CREATE INDEX IF NOT EXISTS idx_tickets_employee ON tickets(employee_id);

CREATE INDEX IF NOT EXISTS idx_employees_department ON employees(department_id);
CREATE INDEX IF NOT EXISTS idx_users_role           ON users(role_id);

INSERT INTO roles(name) VALUES
    ('admin'), ('manager'), ('employee')
ON CONFLICT (name) DO NOTHING;

INSERT INTO departments(id, name) VALUES
    (1, 'Engineering'),
    (2, 'Sales')
ON CONFLICT (id) DO NOTHING;

INSERT INTO employees(id, first_name, last_name, position, department_id) VALUES
    (1, 'Иван',      'Петров',    'Administrator', 1),
    (2, 'Ольга',     'Смирнова',  'Manager',       1),
    (3, 'Сергей',    'Кузнецов',  'Developer',     1),
    (4, 'Мария',     'Иванова',   'Designer',      1),
    (5, 'Алексей',   'Соколов',   'QA Engineer',   1),
    (6, 'Екатерина', 'Новикова',  'Sales',         2),
    (7, 'Дмитрий',   'Морозов',   'Support',       2)
ON CONFLICT (id) DO NOTHING;

INSERT INTO users(id, login, password, employee_id, role_id) VALUES
    (1, 'admin',   '$2a$10$iIk00/FMc9dzxKr9.4KItuUSwl5mRZ1AAr5yMzjnv4zVnKukmTs9y', 1, 1),
    (2, 'manager', '$2a$10$WtJ2mCJnUck5CL50ZmpE.etkEq5/Gz1zq75xJRJpjDo9iBrBAgU6C', 2, 2),
    (3, 'user1',   '$2a$10$5omUgJTeaDY/rN5IcKIq3uvUlZGfUtyTHQCmVBqSF934hYbn8IRpK', 3, 3),
    (4, 'user2',   '$2a$10$vWXCIo/oHrOXFHbXHGwgqeKmDOwdRbyrR4YEStGO0I1qYaeC4r2Hi', 4, 3),
    (5, 'user3',   '$2a$10$hQqCg3jz60V9C7iiyR0KMeINemGxp1.yWHAKcfajhcFrtySKHtunG', 5, 3),
    (6, 'user4',   '$2a$10$R1NeBkopxEJBVgEwpx7hAuL9BqzZAbrcKiWtBNaKUetelzEf9.miq', 6, 3),
    (7, 'user5',   '$2a$10$LS1IUf849uPSkGtL6dZ1cu3AMK7v1ejzbYM28x4NuyXjYm6zseOqO', 7, 3)
ON CONFLICT (id) DO NOTHING;

UPDATE departments SET manager_id = 2 WHERE id = 1 AND manager_id IS NULL;

INSERT INTO attendance(id, employee_id, check_in, check_out) VALUES
    (1, 1, '2026-05-07 06:00:00', '2026-05-07 15:00:00'),
    (2, 2, '2026-05-07 06:30:00', '2026-05-07 15:30:00'),
    (3, 3, '2026-05-07 05:55:00', '2026-05-07 14:15:00'),
    (4, 4, '2026-05-07 06:10:00', '2026-05-07 14:50:00'),
    (5, 5, '2026-05-07 07:00:00', '2026-05-07 16:00:00'),
    (6, 6, '2026-05-07 06:05:00', '2026-05-07 14:30:00'),
    (7, 7, '2026-05-07 06:20:00', '2026-05-07 14:40:00')
ON CONFLICT (id) DO NOTHING;

SELECT setval(pg_get_serial_sequence('roles','id'),       COALESCE((SELECT MAX(id) FROM roles), 1));
SELECT setval(pg_get_serial_sequence('departments','id'), COALESCE((SELECT MAX(id) FROM departments), 1));
SELECT setval(pg_get_serial_sequence('employees','id'),   COALESCE((SELECT MAX(id) FROM employees), 1));
SELECT setval(pg_get_serial_sequence('users','id'),       COALESCE((SELECT MAX(id) FROM users), 1));
SELECT setval(pg_get_serial_sequence('attendance','id'),  COALESCE((SELECT MAX(id) FROM attendance), 1));
