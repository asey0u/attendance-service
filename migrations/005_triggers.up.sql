CREATE OR REPLACE FUNCTION fn_ticket_set_reviewed_at()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status <> OLD.status AND NEW.status IN ('approved', 'declined') THEN
        NEW.reviewed_at = now();
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_ticket_set_reviewed_at ON tickets;
CREATE TRIGGER trg_ticket_set_reviewed_at
    BEFORE UPDATE ON tickets
    FOR EACH ROW
    EXECUTE FUNCTION fn_ticket_set_reviewed_at();


CREATE OR REPLACE FUNCTION fn_attendance_no_duplicate_open()
RETURNS TRIGGER AS $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM attendance
        WHERE employee_id = NEW.employee_id
          AND check_out IS NULL
    ) THEN
        RAISE EXCEPTION 'employee % already has an open check-in', NEW.employee_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_attendance_no_duplicate_open ON attendance;
CREATE TRIGGER trg_attendance_no_duplicate_open
    BEFORE INSERT ON attendance
    FOR EACH ROW
    EXECUTE FUNCTION fn_attendance_no_duplicate_open();


CREATE OR REPLACE FUNCTION fn_ticket_status_immutable()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.status <> 'pending' AND NEW.status <> OLD.status THEN
        RAISE EXCEPTION 'cannot change status of a % ticket', OLD.status;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_ticket_status_immutable ON tickets;
CREATE TRIGGER trg_ticket_status_immutable
    BEFORE UPDATE ON tickets
    FOR EACH ROW
    EXECUTE FUNCTION fn_ticket_status_immutable();
