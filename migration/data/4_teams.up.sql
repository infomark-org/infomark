BEGIN;

ALTER TABLE submissions ADD COLUMN team_id INT;
ALTER TABLE submissions ADD CONSTRAINT FK_SubmissionTeam
FOREIGN KEY (team_id) REFERENCES teams(id);

ALTER TABLE grades ADD COLUMN user_id INT;
ALTER TABLE grades ADD CONSTRAINT FK_GradesUser
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE grades ADD COLUMN team_id INT;
ALTER TABLE grades ADD CONSTRAINT FK_GradesTeam
FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE;

WITH uids AS
(
  SELECT nextval('tasks_id_seq') AS id, user_id
  FROM   user_course
  WHERE  team_id IS NULL
),
tids AS
(
  INSERT INTO teams
  SELECT id
  FROM uids
),
course AS
(
  UPDATE user_course
    SET team_id = uids.id,
        team_confirmed = true
  FROM uids
  WHERE user_course.user_id = uids.user_id
  RETURNING *
)
SELECT COUNT(*) FROM course;

UPDATE submissions s
  SET team_id = uc.team_id
  FROM user_course uc
  WHERE uc.user_id = s.user_id;

DELETE FROM submissions WHERE team_id IS NULL;

ALTER TABLE submissions ALTER COLUMN team_ID SET NOT NULL;


-- Move column user_id from table submissions to table grades
UPDATE grades AS g1
  SET  user_id = s.user_id,
       team_id = s.team_id
FROM submissions s
WHERE s.id = g1.submission_id;

ALTER TABLE grades ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE grades ALTER COLUMN team_id SET NOT NULL;

ALTER TABLE submissions RENAME COLUMN user_id TO upload_user_id;

COMMIT;
