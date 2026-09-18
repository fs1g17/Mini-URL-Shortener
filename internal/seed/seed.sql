-- password is "password" - $2a$14$lfwdV/JNa/O54I37RJx2iOqW6xrI0UONv49JLYFgmocbPTwv4GxH.
TRUNCATE links, users, click_events RESTART IDENTITY;

INSERT INTO users (username, password_hash)
SELECT 'user_' || n,
       '$2a$14$lfwdV/JNa/O54I37RJx2iOqW6xrI0UONv49JLYFgmocbPTwv4GxH.'
FROM generate_series(0, 9) AS series(n);

-- INSERT INTO links (slug , redirect_url , click_count , created_at , updated_at , click_limit , expires , owner_id)
-- i want: 
-- - click limit on 10% of links (can do if modulo is 1 for example)
-- - expiry on 20% of links (same as above) - one in 2 days, one in a week

INSERT INTO links (
              slug,
              redirect_url,
              created_at,
              click_limit,
              expires,
              owner_id
       )
SELECT lpad(n::text, 6, '0'),
       CONCAT('https://example', n, '.com'),
       now() - (10 - MOD(n, 10)) * interval '1 day',
       CASE
              WHEN MOD(n, 10) = 0 THEN 1000
              ELSE NULL
       END,
       CASE
              WHEN MOD(n, 10) = 1 THEN now() + interval '2 days'
              WHEN MOD(n, 10) = 2 THEN now() + interval '9 days'
              ELSE NULL
       END,
       DIV(n, 10) + 1
FROM generate_series(0, 99) AS series(n);

INSERT INTO click_events (
       link_id, 
       clicked_at
)
SELECT
       id,
       now() - INTERVAL '1 month' + random() * INTERVAL '1 month'
FROM generate_series(0, 999) as series(n)
CROSS JOIN (SELECT id FROM links);
