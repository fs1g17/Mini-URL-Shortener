-- password is "password" - $2a$14$lfwdV/JNa/O54I37RJx2iOqW6xrI0UONv49JLYFgmocbPTwv4GxH.

INSERT INTO users (username, password_hash) 
SELECT 'user_' || n, 
       '$2a$14$lfwdV/JNa/O54I37RJx2iOqW6xrI0UONv49JLYFgmocbPTwv4GxH.'
FROM generate_series(0, 9) AS series(n);

-- INSERT INTO links (slug , redirect_url , click_count , created_at , updated_at , click_limit , expires , owner_id)
-- VALUES 
--   ()