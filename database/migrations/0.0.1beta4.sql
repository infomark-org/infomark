CREATE TABLE exams (
  id SERIAL not null primary key,
  created_at TIMESTAMP not null DEFAULT current_timestamp,
  updated_at TIMESTAMP not null DEFAULT current_timestamp,

  name TEXT not null,
  description TEXT not null,
  exam_time TIMESTAMP not null,

  course_id INT not null,
  FOREIGN KEY (course_id) REFERENCES courses (id)  ON DELETE CASCADE
);


CREATE TABLE user_exam(
  id SERIAL not null primary key,
  user_id INT not null,
  exam_id INT not null,

  -- 0 unkown, 1 failed, 2 passed
  status INT null,
  mark TEXT null,

  -- PRIMARY KEY (user_id, exam_id),
  FOREIGN KEY (user_id)   REFERENCES users (id)    ON DELETE CASCADE,
  FOREIGN KEY (exam_id) REFERENCES exams (id)  ON DELETE CASCADE
);
