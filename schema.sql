drop table if exists users;

create table users(
  id serial not null primary key,
  email text not null,
  name text not null
);