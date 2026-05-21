DO $$ BEGIN CREATE ROLE attendance_admin;    EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE ROLE attendance_manager;  EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE ROLE attendance_employee; EXCEPTION WHEN duplicate_object THEN NULL; END $$;

GRANT attendance_admin    TO CURRENT_USER;
GRANT attendance_manager  TO CURRENT_USER;
GRANT attendance_employee TO CURRENT_USER;

GRANT ALL PRIVILEGES ON ALL TABLES    IN SCHEMA public TO attendance_admin;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO attendance_admin;

GRANT SELECT ON ALL TABLES IN SCHEMA public TO attendance_manager;
GRANT INSERT, UPDATE ON attendance TO attendance_manager;
GRANT UPDATE ON tickets TO attendance_manager;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO attendance_manager;

GRANT SELECT ON employees, departments, settings TO attendance_employee;
GRANT SELECT, INSERT ON attendance               TO attendance_employee;
GRANT UPDATE (check_out) ON attendance           TO attendance_employee;
GRANT SELECT, INSERT, DELETE ON tickets          TO attendance_employee;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO attendance_employee;
