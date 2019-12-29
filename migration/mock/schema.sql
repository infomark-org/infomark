-- http://localhost:8081/#
BEGIN;
DROP TABLE IF EXISTS material_course;
DROP TABLE IF EXISTS user_exam;
DROP TABLE IF EXISTS user_course;
DROP TABLE IF EXISTS user_group;
DROP TABLE IF EXISTS sheet_course;
DROP TABLE IF EXISTS task_sheet;
DROP TABLE IF EXISTS group_bids;
--  renamed to task_ratings
-- DROP TABLE IF EXISTS task_feedbacks;
DROP TABLE IF EXISTS task_ratings;

DROP TABLE IF EXISTS materials;
DROP TABLE IF EXISTS groups;
DROP TABLE IF EXISTS grades;
DROP TABLE IF EXISTS exams;
DROP TABLE IF EXISTS submissions;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS sheets;
DROP TABLE IF EXISTS courses;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS schema_migrations;

COMMIT;