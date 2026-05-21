CREATE OR REPLACE VIEW v_attendance_today AS
SELECT
    a.id,
    a.employee_id,
    a.check_in,
    a.check_out,
    e.first_name,
    e.last_name,
    e.position,
    e.department_id
FROM attendance a
JOIN employees e ON e.id = a.employee_id
WHERE a.check_in::date = CURRENT_DATE;


CREATE OR REPLACE VIEW v_pending_tickets AS
SELECT
    t.id,
    t.employee_id,
    t.ticket_date,
    t.reason,
    t.status,
    t.created_at,
    e.first_name,
    e.last_name,
    e.department_id
FROM tickets t
JOIN employees e ON e.id = t.employee_id
WHERE t.status = 'pending';


CREATE OR REPLACE FUNCTION fn_employee_stats(
    p_employee_id INT,
    p_from        TIMESTAMP DEFAULT NULL,
    p_to          TIMESTAMP DEFAULT NULL
)
RETURNS TABLE(
    days_present  INT,
    days_late     INT,
    total_hours   NUMERIC,
    avg_hours     NUMERIC
)
LANGUAGE plpgsql AS $$
DECLARE
    v_threshold TIME;
    v_threshold_str TEXT;
BEGIN
    SELECT value INTO v_threshold_str FROM settings WHERE key = 'late_threshold';
    v_threshold := COALESCE(v_threshold_str, '09:30')::TIME;

    RETURN QUERY
    SELECT
        COUNT(*)::INT,
        COUNT(*) FILTER (WHERE (check_in AT TIME ZONE 'UTC')::TIME > v_threshold)::INT,
        ROUND(COALESCE(SUM(
            EXTRACT(EPOCH FROM (check_out - check_in)) / 3600
        ) FILTER (WHERE check_out IS NOT NULL), 0)::NUMERIC, 2),
        ROUND(COALESCE(SUM(
            EXTRACT(EPOCH FROM (check_out - check_in)) / 3600
        ) FILTER (WHERE check_out IS NOT NULL) / NULLIF(COUNT(*), 0), 0)::NUMERIC, 2)
    FROM attendance
    WHERE employee_id = p_employee_id
      AND (p_from IS NULL OR check_in >= p_from)
      AND (p_to   IS NULL OR check_in <= p_to);
END;
$$;


CREATE OR REPLACE PROCEDURE pr_review_ticket(
    p_ticket_id      INT,
    p_reviewer_id    INT,
    p_status         TEXT
)
LANGUAGE plpgsql AS $$
BEGIN
    UPDATE tickets
    SET status = p_status, reviewed_by = p_reviewer_id
    WHERE id = p_ticket_id AND status = 'pending';

    IF NOT FOUND THEN
        RAISE EXCEPTION 'ticket % not found or already reviewed', p_ticket_id
            USING ERRCODE = 'P0404';
    END IF;
END;
$$;
