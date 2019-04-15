-- http://localhost:8081/#
BEGIN;
drop table if exists material_course;
drop table if exists user_course;
drop table if exists user_group;
drop table if exists sheet_course;
drop table if exists task_sheet;
drop table if exists group_bids;
--  renamed to task_ratings
-- drop table if exists task_feedbacks;
drop table if exists task_ratings;

drop table if exists materials;
drop table if exists groups;
drop table if exists grades;
drop table if exists submissions;
drop table if exists tasks;
drop table if exists sheets;
drop table if exists courses;
drop table if exists users;

CREATE TABLE users (
  id SERIAL not null primary key,
  created_at TIMESTAMP not null DEFAULT current_timestamp,
  updated_at TIMESTAMP not null DEFAULT current_timestamp,

  first_name TEXT not null,
  last_name TEXT not null,
  -- we need to avatar_path (as it might be empty)
  avatar_url TEXT,
  email TEXT not null unique,
  student_number TEXT not null,
  semester INT not null,
  subject TEXT not null,

  language char(2) not null DEFAULT 'en',

  encrypted_password TEXT not null,
  reset_password_token TEXT,
  confirm_email_token TEXT,
  root BOOLEAN not null DEFAULT false
);

CREATE TABLE courses (
  id SERIAL not null primary key,
  created_at TIMESTAMP not null DEFAULT current_timestamp,
  updated_at TIMESTAMP not null DEFAULT current_timestamp,

  name TEXT not null,
  description TEXT not null,
  begins_at TIMESTAMP not null,
  ends_at TIMESTAMP not null,
  required_percentage INT  DEFAULT 0
);


CREATE TABLE user_course(
  id SERIAL not null primary key,
  user_id INT not null,
  course_id INT not null,
  -- 0: student, 1:tutor, 2:admin
  role INT DEFAULT 0,

  -- PRIMARY KEY (user_id, course_id),
  FOREIGN KEY (user_id)   REFERENCES users (id)    ON DELETE CASCADE,
  FOREIGN KEY (course_id) REFERENCES courses (id)  ON DELETE CASCADE
);

CREATE TABLE sheets(
  id serial not null primary key,
  created_at TIMESTAMP not null DEFAULT current_timestamp,
  updated_at TIMESTAMP not null DEFAULT current_timestamp,

  name TEXT not null,

  -- we us the canonical naming "sheet{ordering}.zip"
  -- file_path TEXT,
  publish_at TIMESTAMP not null,
  due_at TIMESTAMP not null
);


CREATE TABLE sheet_course(
  id SERIAL not null primary key,
  sheet_id INT not null,
  course_id INT not null,
  -- ordering INT not null,

  -- PRIMARY KEY (sheet_id, course_id),
  FOREIGN KEY (sheet_id)  REFERENCES sheets (id)    ON DELETE CASCADE,
  FOREIGN KEY (course_id) REFERENCES courses (id)   ON DELETE CASCADE
);


CREATE TABLE tasks(
  id serial not null primary key,
  created_at TIMESTAMP not null DEFAULT current_timestamp,
  updated_at TIMESTAMP not null DEFAULT current_timestamp,

  name TEXT not null,
  max_points INT DEFAULT 0,
  -- we keep both paths as they might be empty
  -- public_test_path TEXT,
  -- private_test_path TEXT,

  public_docker_image TEXT,
  private_docker_image TEXT
);

CREATE TABLE task_sheet(
  id SERIAL not null primary key,
  task_id INT not null,
  sheet_id INT not null,
  -- ordering INT ,

  -- PRIMARY KEY (task_id, sheet_id),
  FOREIGN KEY (task_id) REFERENCES tasks (id)     ON DELETE CASCADE,
  FOREIGN KEY (sheet_id) REFERENCES sheets (id)   ON DELETE CASCADE
);

CREATE TABLE submissions(
  id SERIAL not null primary key,
  created_at TIMESTAMP not null DEFAULT current_timestamp,
  updated_at TIMESTAMP not null DEFAULT current_timestamp,

  user_id INT not null,
  task_id INT not null,
  -- not necessary we use schema "{submission_id}.zip"
  -- file_path TEXT not null,

  -- PRIMARY KEY (user_id, task_id),
  FOREIGN KEY (user_id) REFERENCES users (id)   ON DELETE CASCADE,
  FOREIGN KEY (task_id) REFERENCES tasks (id)   ON DELETE CASCADE
);

CREATE TABLE grades(
  id SERIAL not null primary key,
  created_at TIMESTAMP not null DEFAULT current_timestamp,
  updated_at TIMESTAMP not null DEFAULT current_timestamp,

  -- 0: pending, 1: running, 2: finished
  public_execution_state INT DEFAULT 0,
  private_execution_state INT DEFAULT 0,

  public_test_log TEXT,
  private_test_log TEXT,

  -- 0 means ok, 1 failed (just like return codes)
  public_test_status INT  DEFAULT 0,
  private_test_status INT  DEFAULT 0,

  acquired_points INT  DEFAULT 0,

  feedback TEXT,

  tutor_id INT not null,
  submission_id INT not null,

  -- PRIMARY KEY (tutor_id, submission_id),
  FOREIGN KEY (tutor_id) REFERENCES users (id)              ON DELETE CASCADE,
  FOREIGN KEY (submission_id) REFERENCES submissions (id)   ON DELETE CASCADE
);

-- exercise groups
CREATE TABLE groups(
  id SERIAL not null primary key,
  created_at TIMESTAMP not null DEFAULT current_timestamp,
  updated_at TIMESTAMP not null DEFAULT current_timestamp,

  tutor_id INT not null,
  course_id INT not null,
  description TEXT not null,

  -- PRIMARY KEY (tutor_id, course_id),
  FOREIGN KEY (tutor_id) REFERENCES users (id)    ON DELETE CASCADE,
  FOREIGN KEY (course_id) REFERENCES courses (id) ON DELETE CASCADE
);

CREATE TABLE user_group(
  id SERIAL not null primary key,
  user_id INT not null,
  group_id INT not null,

  -- PRIMARY KEY (user_id, group_id),
  FOREIGN KEY (user_id)   REFERENCES users (id)    ON DELETE CASCADE,
  FOREIGN KEY (group_id) REFERENCES groups (id)  ON DELETE CASCADE
);

-- ratings of task from students
CREATE TABLE task_ratings(
  id SERIAL not null primary key,

  user_id INT not null,
  task_id INT not null,
  rating INT DEFAULT 0,

  -- PRIMARY KEY (user_id, task_id),
  FOREIGN KEY (user_id) REFERENCES users (id)  ON DELETE CASCADE,
  FOREIGN KEY (task_id) REFERENCES tasks (id)  ON DELETE CASCADE
);

-- bids of students to get into a specific exercise group
CREATE TABLE group_bids(
  id SERIAL not null primary key,

  user_id INT not null,
  group_id INT not null,
  bid INT DEFAULT 0,

  -- PRIMARY KEY (user_id, group_id),
  FOREIGN KEY (user_id) REFERENCES users (id)   ON DELETE CASCADE,
  FOREIGN KEY (group_id) REFERENCES groups (id) ON DELETE CASCADE
);

CREATE TABLE materials(
  id SERIAL not null primary key,
  created_at TIMESTAMP not null DEFAULT current_timestamp,
  updated_at TIMESTAMP not null DEFAULT current_timestamp,

  name TEXT not null,
  -- should be unique in one course (keep original filename as it has some meaning)
  -- filename TEXT not null,
  -- 0: slide, 1: supplementary
  kind INT DEFAULT 0,
  -- required_role INT DEFAULT 0,
  publish_at TIMESTAMP not null,
  lecture_at TIMESTAMP not null
);

CREATE TABLE material_course(
  id SERIAL not null primary key,
  material_id INT not null,
  course_id INT not null,

  -- PRIMARY KEY (material_id, course_id),
  FOREIGN KEY (material_id) REFERENCES materials (id)   ON DELETE CASCADE,
  FOREIGN KEY (course_id) REFERENCES courses (id)       ON DELETE CASCADE
);

-- add_index "users", ["confirmation_token"], name: "index_users_on_confirmation_token", unique: true, using: :btree
-- add_index "users", ["email"], name: "index_users_on_email", unique: true, using: :btree

COMMIT;