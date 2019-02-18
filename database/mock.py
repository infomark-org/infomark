from faker import Factory
from faker import Faker
from enum import Enum
import bcrypt
from collections import OrderedDict
import datetime

NUM_STUDENTS = 300
NUM_TUTORS = 10
NUM_ADMINS = 1


class VAL(Enum):
  DEFAULT = 1
  NULL = 2
  TIMESTAMP = 3
  TRUE = 4
  FALSE = 5


salt = bcrypt.gensalt()
default_encrypted_password = (bcrypt.hashpw(
    'test'.encode('utf-8'), salt)).decode("utf-8")


def time_stamp(time):
  # 2019-02-13T13:26:54.952595Z
  # datetime.datetime(2013, 9, 30, 7, 6, 5)
  return time.strftime('%Y-%m-%dT%H:%M:%S.%fZ')


def create_user(fake, role='student'):
  first_name = fake.first_name()
  last_name = fake.last_name()
  if role == 'admin':
    email = "%s.%s@.uni-tuebingen.de" % (first_name, last_name)
  elif role == 'tutor':
    email = "%s.%s@tutor.uni-tuebingen.de" % (first_name, last_name)
  else:
    email = "%s.%s@student.uni-tuebingen.de" % (first_name, last_name)
  email = email.lower()

  root = VAL.TRUE if (role == 'admin') else VAL.FALSE

  data = OrderedDict([
      ('id', VAL.DEFAULT),
      ('created_at', VAL.TIMESTAMP),
      ('updated_at', VAL.TIMESTAMP),
      ('first_name', first_name),
      ('last_name', last_name),
      ('avatar_path', VAL.NULL),
      ('email', email),
      ('student_number', fake.random_int(1000, 2000)),
      ('semester', fake.random_int(2, 8)),
      ('subject', 'computer science'),
      ('language', 'de'),
      ('encrypted_password', default_encrypted_password),
      ('reset_password_token', VAL.NULL),
      ('confirm_email_token', VAL.NULL),
      ('root', root),
  ])

  return data


def create_course(name):
  data = OrderedDict([
      ('id', VAL.DEFAULT),
      ('created_at', VAL.TIMESTAMP),
      ('updated_at', VAL.TIMESTAMP),

      ('name', name),
      ('description', 'Lorem Ipsum'),
      ('begins_at', time_stamp(datetime.datetime(2019, 2, 1, 1, 2, 3))),
      ('ends_at', time_stamp(datetime.datetime(2019, 7, 30, 23, 59, 59))),
      ('required_percentage', fake.random_int(10, 100)),
  ])

  return data


def create_user_course(user_id, course_id, role):
  data = OrderedDict([
      ('id', VAL.DEFAULT),
      ('user_id', user_id),
      ('course_id', course_id),
      ('role', role),
  ])

  return data


def create_sheet(name):
  data = OrderedDict([
      ('id', VAL.DEFAULT),
      ('created_at', VAL.TIMESTAMP),
      ('updated_at', VAL.TIMESTAMP),

      ('name', name),

      ('publish_at', time_stamp(datetime.datetime(2019, 2, 1, 1, 2, 3))),
      ('due_at', time_stamp(datetime.datetime(2019, 7, 30, 23, 59, 59))),
  ])

  return data


def create_sheet_course(sheet_id, course_id, k):
  data = OrderedDict([
      ('id', VAL.DEFAULT),
      ('sheet_id', sheet_id),
      ('course_id', course_id),
      ('ordering', k),
  ])

  return data


def create_task():
  data = OrderedDict([
      ('id', VAL.DEFAULT),
      ('created_at', VAL.TIMESTAMP),
      ('updated_at', VAL.TIMESTAMP),

      ('max_points', fake.random_int(10, 100)),
      ('public_test_path', 'path/to/tests/public_dummy_test'),
      ('private_test_path', 'path/to/tests/private_dummy_test'),

      ('public_docker_image', 'ImageCIRunnerJavaEnv'),
      ('private_docker_image', 'ImageCIRunnerJavaEnv'),
  ])

  return data


def create_task_sheet(task_id, sheet_id, k):
  data = OrderedDict([
      ('id', VAL.DEFAULT),
      ('task_id', task_id),
      ('sheet_id', sheet_id),
      ('ordering', k),
  ])

  return data


# """"""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""
# """"""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""
# """"""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

def to_statement(table, arr):
  cols = []
  values = []

  for k in arr:
    cols.append(k)
    if arr[k] == VAL.DEFAULT:
      values.append('DEFAULT')
    elif arr[k] == VAL.NULL:
      values.append('NULL')
    elif arr[k] == VAL.TIMESTAMP:
      values.append('current_timestamp')
    elif isinstance(arr[k], int):
      values.append("%i" % arr[k])
    elif arr[k] == VAL.FALSE:
      values.append('FALSE')
    elif arr[k] == VAL.TRUE:
      values.append('TRUE')
    else:
      values.append("'%s'" % arr[k])
  cols = ','.join(cols)
  values = ','.join(values)
  stmt = "INSERT INTO %s (%s) VALUES (%s);\n" % (table, cols, values)
  return stmt


if __name__ == "__main__":
  fake = Factory.create()
  fake = Faker('de_DE')
  fake.seed_instance(0)

  # print(time_stamp())

  # generate admins+tutors+students
  # ----------------------------------------------------------------------------
  admins = [create_user(fake, role='admin') for _ in range(NUM_ADMINS)]
  tutors = [create_user(fake, role='tutor') for _ in range(NUM_TUTORS)]
  students = [create_user(fake, role='student') for _ in range(NUM_STUDENTS)]

  admins[0]['email'] = 'test@uni-tuebingen.de'

  courses = [create_course('Info2'), create_course('Info3')]

  course_id = 1

  taskCounter = 1
  sheetCounter = 1

  sheet_course = []
  sheets = []
  tasks = []
  task_sheet = []
  for i in range(10):
    sheets.append(create_sheet('Blatt %i' % i))
    sheet_course.append(create_sheet_course(
        sheetCounter, course_id, sheetCounter))
    for k in range(3):
      tasks.append(create_task())
      task_sheet.append(create_task_sheet(taskCounter, sheetCounter, k))
      taskCounter += 1
    sheetCounter += 1

  with open('mock.sql', 'w') as f:
    # users
    f.write('DELETE FROM users;\n')
    f.write('ALTER SEQUENCE users_id_seq RESTART WITH 1;\n')
    for admin in admins:
      f.write(to_statement('users', admin))
    for tutor in tutors:
      f.write(to_statement('users', tutor))
    for student in students:
      f.write(to_statement('users', student))

    # courses
    f.write('DELETE FROM courses;\n')
    f.write('ALTER SEQUENCE courses_id_seq RESTART WITH 1;\n')
    for course in courses:
      f.write(to_statement('courses', course))

    # enrollments
    f.write('DELETE FROM user_course;\n')
    f.write('ALTER SEQUENCE user_course_id_seq RESTART WITH 1;\n')

    for k, admin in enumerate(admins):
      k += 1
      f.write(to_statement('user_course', create_user_course(k, course_id, 2)))

    for k, tutor in enumerate(tutors):
      k += 1
      k += len(admins)
      f.write(to_statement('user_course', create_user_course(k, course_id, 1)))

    # enrollments
    for k, student in enumerate(students):
      k += 1
      k += len(admins)
      k += len(tutors)
      f.write(to_statement('user_course', create_user_course(k, course_id, 0)))

    # sheets
    f.write('DELETE FROM sheets;\n')
    f.write('ALTER SEQUENCE sheets_id_seq RESTART WITH 1;\n')
    for sheet in sheets:
      f.write(to_statement('sheets', sheet))

    for el in sheet_course:
      f.write(to_statement('sheet_course', el))

    f.write('DELETE FROM tasks;\n')
    f.write('ALTER SEQUENCE tasks_id_seq RESTART WITH 1;\n')
    for task in tasks:
      f.write(to_statement('tasks', task))

    for el in task_sheet:
      f.write(to_statement('task_sheet', el))

