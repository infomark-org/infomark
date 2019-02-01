-- http://localhost:8081/#
drop table if exists users;

create table users (
  id serial ,
  created_at timestamp not null,
  updated_at timestamp not null,

  first_name text not null,
  last_name text not null,
  avatar_url text,
  email text not null unique,
  student_number text not null,
  semester int not null,
  subject text not null,

  encrypted_password text not null,
  reset_password_token text,
  confirm_email_token text,
  root boolean not null default false
);

-- add_index "users", ["confirmation_token"], name: "index_users_on_confirmation_token", unique: true, using: :btree
-- add_index "users", ["email"], name: "index_users_on_email", unique: true, using: :btree