drop table if exists users;

drop table if exists users;

create table users (
  id serial ,
  created_at timestamp not null,
  updated_at timestamp not null,

  first_name text not null,
  last_name text not null,
  avatar_url text,
  email text not null,
  student_number text not null,
  semester text not null,
  subject text not null,

  encrypted_password text not null,
  reset_password_token text not null,
  confirm_email_token text not null,
  root boolean not null
);


add_index "users", ["confirmation_token"], name: "index_users_on_confirmation_token", unique: true, using: :btree
add_index "users", ["email"], name: "index_users_on_email", unique: true, using: :btree