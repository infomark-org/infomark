-- from https://github.com/cgtuebingen/InfoMark/blob/396cd6077c5540bcb8fb3b90b07138387e2c736e/server/db/schema.rb
-- RUN: psql < schema.sql
-- RUN: go generate  # uses sqlboiler to generate all necessary models


drop table if exists users_courses;
drop table if exists biddings;
drop table if exists manuscripts;
drop table if exists groups;
drop table if exists gradings;
drop table if exists submissions;
drop table if exists tasks;
drop table if exists sheets;
drop table if exists courses;
drop table if exists users;

create table users (
  email text not null,
  name text not null,

  id serial not null primary key,
  created_at timestamp not null,
  updated_at timestamp not null
);

-- represent the entire course, like "Info 2"
create table courses (
  title text not null,
  subtitle text,
  description text,
  begin_at timestamp not null,
  ends_at timestamp not null,

  id serial not null primary key,
  created_at timestamp not null,
  updated_at timestamp not null,

  author_id int not null,
  foreign key (user_id) references users (id)
);

-- enrollments
create table users_courses (
  -- role: 0 student, 1 tutor, 2 ..., 99 admin
  role int default 0,

  created_at timestamp not null,
  updated_at timestamp not null,

  user_id int not null,
  course_id int not null,

  primary key (user_id, course_id),
  foreign key (user_id) references users (id),
  foreign key (course_id) references courses (id)
);

-- exercise sheets
create table sheets(
  ordering int not null default 0,
  description text,
  publish_at timestamp not null,
  deadline_at timestamp not null,
  filename text not null,

  id serial not null primary key,
  created_at timestamp not null,
  updated_at timestamp not null,

  course_id int not null,
  foreign key (course_id) references courses (id)

);

-- exercise tasks
create table tasks(
  ordering int not null default 0,
  max_points int not null default 0,
  test_public text,
  test_private text,

  id serial not null primary key,
  created_at timestamp not null,
  updated_at timestamp not null,

  sheet_id int not null,
  foreign key (sheet_id) references sheets (id)
);

-- submitted files
create table submissions (
  -- queued, processing, done
  status int not null,
  filename text not null,

  id serial not null primary key,
  created_at timestamp not null,
  updated_at timestamp not null,

  user_id int not null,
  exercise_task_id int not null,
  foreign key (user_id) references users (id),
  foreign key (exercise_task_id) references tasks (id)
);

-- gradings given from users with role>0 for a given course
create table gradings(
  feedback text,
  submission_id int not null,
  user_id int not null,
  points int not null,

  created_at timestamp not null,
  updated_at timestamp not null,

  primary key (submission_id),
  foreign key (submission_id) references submissions (id),
  foreign key (user_id) references users (id)
);

-- exercise groups
create table groups(
  info text,


  id serial not null primary key,
  created_at timestamp not null,
  updated_at timestamp not null,

  user_id int not null,
  course_id int not null,
  foreign key (course_id) references courses (id),
  foreign key (user_id) references users (id)
);

-- biddigns for specific exercise groups
create table biddings(
  bid int not null,

  user_id int not null,
  group_id int not null,
  foreign key (group_id) references groups (id),
  foreign key (user_id) references users (id)
);

-- lecture material: slides, extra infos, ...
create table manuscripts(
  filename text not null,

  id serial not null primary key,
  created_at timestamp not null,
  updated_at timestamp not null,

  course_id int not null,
  foreign key (course_id) references courses (id)
);