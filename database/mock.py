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


def random_language():
  candidates = ["en", "de"]
  idx = fake.random_int(0, len(candidates) - 1)
  return candidates[idx]


def random_subject():
  candidates = ["math", "bioinfo", "info", "mediainfo", 'computer science']
  idx = fake.random_int(0, len(candidates) - 1)
  return candidates[idx]


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
      # ('avatar_path', VAL.NULL),
      ('email', email),
      ('student_number', fake.random_int(1000, 2000)),
      ('semester', fake.random_int(2, 8)),
      ('subject', random_subject()),
      ('language', random_language()),
      ('encrypted_password', default_encrypted_password),
      ('reset_password_token', VAL.NULL),
      ('confirm_email_token', VAL.NULL),
      ('root', root),
  ])

  return data


def create_course(fake, name):
  data = OrderedDict([
      ('id', VAL.DEFAULT),
      ('created_at', VAL.TIMESTAMP),
      ('updated_at', VAL.TIMESTAMP),

      ('name', name),
      ('description', fake.text()),
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
      # ('ordering', k),
  ])

  return data


def create_task(k):
  data = OrderedDict([
      ('id', VAL.DEFAULT),
      ('created_at', VAL.TIMESTAMP),
      ('updated_at', VAL.TIMESTAMP),

      ('max_points', fake.random_int(10, 100)),
      ('name', "task %i" % k),
      # ('public_test_path', 'path/to/tests/public_dummy_test'),
      # ('private_test_path', 'path/to/tests/private_dummy_test'),

      ('public_docker_image', 'ImageCIRunnerJavaEnv'),
      ('private_docker_image', 'ImageCIRunnerJavaEnv'),
  ])

  return data


def create_task_sheet(task_id, sheet_id, k):
  data = OrderedDict([
      ('id', VAL.DEFAULT),
      ('task_id', task_id),
      ('sheet_id', sheet_id),
      # ('ordering', k),
  ])

  return data


def create_submission(user_id, task_id):
  data = OrderedDict([
      ('id', VAL.DEFAULT),
      ('created_at', VAL.TIMESTAMP),
      ('updated_at', VAL.TIMESTAMP),

      ('user_id', user_id),
      ('task_id', task_id),
  ])

  return data


def create_grade(submission_id, tutor_id, max_points):

  dummy_log = """[   OK   ] BinaryToStringValueTest:
[   OK   ] BinaryToStringClassStructureTest:
[   OK   ] AckermanClassStructureTest:
[ FAILED ] AckermannValueTest:
        Error 1/1
          - Tag: failure
          - Typ: junit.framework.AssertionFailedError
          - Msg: ackermann(2, 2) expected:<7> but was:<9>

[   OK   ] FibonacciValueTest:
[   OK   ] FibonacciClassStructureTest:
[ FAILED ] FactorialValueTest:
        Error 1/1
          - Tag: error
          - Typ: java.lang.ClassCastException
          - Msg: java.base/java.lang.Float cannot be cast to java.base/java.lang.Integer

[ FAILED ] FactorialClassStructureTest:
        Error 1/1
          - Tag: failure
          - Typ: junit.framework.AssertionFailedError
          - Msg: Method `public static float factorial (int )` in `class Factorial` found, but expected return type (`int`) is wrong. I just found `float`
"""
  graded = fake.random_int(0, 1)

  data = OrderedDict([
      ('id', VAL.DEFAULT),
      ('created_at', VAL.TIMESTAMP),
      ('updated_at', VAL.TIMESTAMP),

      ('public_execution_state', fake.random_int(0, 2)),
      ('private_execution_state', fake.random_int(0, 2)),

      ('acquired_points', fake.random_int(0, max_points) if graded else 0),
      ('public_test_log', dummy_log),
      ('private_test_log', dummy_log),

      ('public_test_status', fake.random_int(0, 1)),
      ('private_test_status', fake.random_int(0, 1)),

      ('feedback', 'Lorem Ipsum Feedback' if graded else ""),

      ('tutor_id', tutor_id),
      ('submission_id', submission_id),
  ])

  return data


def create_group(fake, tutor_id, course_id):
  data = OrderedDict([
      ('id', VAL.DEFAULT),
      ('created_at', VAL.TIMESTAMP),
      ('updated_at', VAL.TIMESTAMP),

      ('tutor_id', tutor_id),
      ('course_id', course_id),

      ('description', fake.text()),

  ])

  return data


def create_user_group(user_id, group_id):
  data = OrderedDict([
      ('id', VAL.DEFAULT),
      ('user_id', user_id),
      ('group_id', group_id),
  ])

  return data


def create_task_rating(student_id, task_id):
  data = OrderedDict([
      ('id', VAL.DEFAULT),
      ('user_id', student_id),
      ('task_id', task_id),
      ('rating', fake.random_int(1, 5)),
  ])

  return data


def create_group_bid(student_id, group_id):
  data = OrderedDict([
      ('id', VAL.DEFAULT),
      ('user_id', student_id),
      ('group_id', group_id),
      ('bid', fake.random_int(0, 10)),
  ])

  return data


def create_material(fake):
  data = OrderedDict([
      ('id', VAL.DEFAULT),
      ('created_at', VAL.TIMESTAMP),
      ('updated_at', VAL.TIMESTAMP),

      ('name', fake.text()),
      # ('filename', "path2"),
      ('kind', fake.random_int(0, 1)),
      ('required_role', fake.random_int(0, 2)),
      ('publish_at', VAL.TIMESTAMP),
      ('lecture_at', VAL.TIMESTAMP),
  ])

  return data


def create_material_course(material_id, course_id):
  data = OrderedDict([
      ('id', VAL.DEFAULT),
      ('material_id', material_id),
      ('course_id', course_id),
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
  # fake = Factory.create()
  fake = Faker('de_DE')
  fake.seed_instance(0)

  # print(time_stamp())

  # generate admins+tutors+students
  # ----------------------------------------------------------------------------
  admins = [create_user(fake, role='admin') for _ in range(NUM_ADMINS)]
  tutors = [create_user(fake, role='tutor') for _ in range(NUM_TUTORS)]
  students = [create_user(fake, role='student') for _ in range(NUM_STUDENTS)]

  admins[0]['email'] = 'test@uni-tuebingen.de'

  courses = [create_course(fake, 'Info2'), create_course(fake, 'Info3')]

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
      tasks.append(create_task(k))
      task_sheet.append(create_task_sheet(taskCounter, sheetCounter, k))
      taskCounter += 1
    sheetCounter += 1

  submissions = []
  grades = []

  for student_idx in range(1, NUM_STUDENTS + 1):
    student_id = student_idx + NUM_TUTORS + NUM_ADMINS

    for task_idx in range(1, len(tasks) + 1):

      max_pts = tasks[task_idx - 1]['max_points']

      # number_of_submissions = fake.random_int(0, 10)
      # We discussed to delete all but the newest submission
      number_of_submissions = 1
      for i in range(number_of_submissions):
        s = create_submission(student_id, task_idx)

        submissions.append(s)
        tid = fake.random_int(1, len(tutors))
        grades.append(create_grade(len(submissions), tid, max_pts))

  # we just do groups for course with id 1
  groups = []
  for i in range(len(tutors)):
    groups.append(create_group(fake, tutor_id=i + 1, course_id=1))

  # enroll students in random groups
  group_enrollments = []
  for i in range(len(students)):
    student_id = NUM_ADMINS + NUM_TUTORS + i + 1
    group_id = fake.random_int(1, len(groups))

  task_ratings = []
  for i in range(len(students)):
    student_id = NUM_ADMINS + NUM_TUTORS + i + 1
    for j in range(len(tasks)):
      task_id = j + 1
      task_ratings.append(create_task_rating(student_id, task_id))

  group_bids = []
  for i in range(len(students)):
    student_id = NUM_ADMINS + NUM_TUTORS + i + 1
    for j in range(len(groups)):
      group_id = j + 1
      group_bids.append(create_group_bid(student_id, group_id))

  materials = []
  material_course = []
  for i in range(10):
    course_id = 1
    materials.append(create_material(fake))
    material_course.append(create_material_course(i + 1, 1))

  with open('mock.sql', 'w') as f:
    f.write('BEGIN;')
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

    for cid in [1, 2]:
      for k, admin in enumerate(admins):
        k += 1
        f.write(to_statement('user_course', create_user_course(k, cid, 2)))

      for k, tutor in enumerate(tutors):
        k += 1
        k += len(admins)
        f.write(to_statement('user_course', create_user_course(k, cid, 1)))

      # enrollments
      for k, student in enumerate(students):
        k += 1
        k += len(admins)
        k += len(tutors)
        f.write(to_statement('user_course', create_user_course(k, cid, 0)))

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

    # tasks
    f.write('ALTER SEQUENCE task_sheet_id_seq RESTART WITH 1;\n')
    for el in task_sheet:
      f.write(to_statement('task_sheet', el))

    # submissions
    f.write('ALTER SEQUENCE submissions_id_seq RESTART WITH 1;\n')
    for el in submissions:
      f.write(to_statement('submissions', el))

    # grades
    f.write('ALTER SEQUENCE grades_id_seq RESTART WITH 1;\n')
    for el in grades:
      f.write(to_statement('grades', el))

    # groups
    f.write('ALTER SEQUENCE groups_id_seq RESTART WITH 1;\n')
    for el in groups:
      f.write(to_statement('groups', el))

    # groups-enrollments
    f.write('ALTER SEQUENCE user_group_id_seq RESTART WITH 1;\n')
    for k, admin in enumerate(students):
      group_id = fake.random_int(1, len(groups))
      user_id = k + 1
      f.write(to_statement('user_group', create_user_group(user_id, group_id)))

    f.write('ALTER SEQUENCE task_ratings_id_seq RESTART WITH 1;\n')
    for t in task_ratings:
      f.write(to_statement('task_ratings', t))

    f.write('ALTER SEQUENCE group_bids_id_seq RESTART WITH 1;\n')
    for t in group_bids:
      f.write(to_statement('group_bids', t))

    f.write('ALTER SEQUENCE materials_id_seq RESTART WITH 1;\n')
    for t in materials:
      f.write(to_statement('materials', t))

    f.write('ALTER SEQUENCE material_course_id_seq RESTART WITH 1;\n')
    for t in material_course:
      f.write(to_statement('material_course', t))

    f.write('COMMIT;')
