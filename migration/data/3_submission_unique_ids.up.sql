BEGIN;
-- Do not allow duplicate submissions from the same user
ALTER TABLE submissions ADD UNIQUE(task_id, user_id);
COMMIT;
