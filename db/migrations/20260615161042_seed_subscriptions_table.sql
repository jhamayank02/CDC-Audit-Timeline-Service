-- +goose Up
SELECT 'up SQL query';
INSERT INTO subscriptions (
    id,
    user_id,
    plan_name,
    status,
    start_date,
    end_date,
    auto_renew,
    created_by
)
SELECT
    seed.id,
    users.id,
    seed.plan_name,
    seed.status,
    seed.start_date,
    seed.end_date,
    seed.auto_renew,
    seed.created_by
FROM (
    VALUES
        ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1'::UUID, 'aarav.sharma@example.com', 'Basic', 'active', '2026-06-01 00:00:00'::TIMESTAMP, '2026-07-01 00:00:00'::TIMESTAMP, TRUE, '00000000-0000-0000-0000-000000000001'::UUID),
        ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa2'::UUID, 'isha.patel@example.com', 'Pro', 'active', '2026-06-05 00:00:00'::TIMESTAMP, '2026-07-05 00:00:00'::TIMESTAMP, TRUE, '00000000-0000-0000-0000-000000000001'::UUID),
        ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa3'::UUID, 'kabir.mehta@example.com', 'Business', 'active', '2026-06-10 00:00:00'::TIMESTAMP, '2026-07-10 00:00:00'::TIMESTAMP, TRUE, '00000000-0000-0000-0000-000000000001'::UUID),
        ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa4'::UUID, 'ananya.rao@example.com', 'Pro', 'expired', '2026-05-01 00:00:00'::TIMESTAMP, '2026-06-01 00:00:00'::TIMESTAMP, FALSE, '00000000-0000-0000-0000-000000000001'::UUID),
        ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa5'::UUID, 'rohan.verma@example.com', 'Basic', 'cancelled', '2026-05-15 00:00:00'::TIMESTAMP, '2026-06-15 00:00:00'::TIMESTAMP, FALSE, '00000000-0000-0000-0000-000000000001'::UUID)
) AS seed(id, email, plan_name, status, start_date, end_date, auto_renew, created_by)
JOIN users ON users.email = seed.email;

-- +goose Down
SELECT 'down SQL query';
DELETE FROM subscriptions
WHERE id IN (
    'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1',
    'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa2',
    'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa3',
    'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa4',
    'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa5'
);
