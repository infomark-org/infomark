BEGIN;

CREATE TABLE IF NOT EXISTS teams (
	id SERIAL not null primary key
);

-- Add new team data to user
ALTER TABLE user_course ADD COLUMN team_confirmed BOOLEAN DEFAULT FALSE;
ALTER TABLE user_course ADD COLUMN team_id INT DEFAULT NULL;
ALTER TABLE user_course ADD CONSTRAINT FK_UserCourseTeam
FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE SET DEFAULT;

-- Add team size to course
ALTER TABLE courses ADD COLUMN max_team_size INT DEFAULT 1;

COMMIT;
