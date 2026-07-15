# Tutor's guide

# Auto-tests

An auto-test (unit test) consists of

* a submission `<STUDENT_UPLOAD.ZIP>` made by a student
* a testing framework `<TASK_SEPCIFIC_TEST.ZIP>` written by one of the tutors
  or instructors
* the name of a Docker image `<YOUR_DOCKER_IMAGE>`

Each InfoMark worker fetches a submission from a queue and executes a command
similar to:

```bash
docker run --rm -it --net="none" \
  -v <STUDENT_UPLOAD.ZIP>:/data/submission.zip:ro \
  -v <TASK_SEPCIFIC_TEST.ZIP>:/data/unittest.zip:ro  \
  <YOUR_DOCKER_IMAGE>
```

The output is stored in the database and displayed to the student and to the
tutor who grades the solution.

> In fact, the command is slightly more complicated, as the tests run with a
> limited amount of memory (without swap) and a limited number of cores to
> avoid overloading the system (infinite loops, memory leaks in submissions).
> These settings can be defined in the InfoMark configuration file. We also act
> directly as a Docker client instead of calling a plain shell command.


The most minimal and simple Dockerfile that handles uploaded solutions is:

```docker
# Dockerfile
FROM ubuntu:18.04
ADD scripts/run.sh /app/run.sh
ENTRYPOINT ["/app/run.sh"]
```

which uses the script

```bash
# run.sh
DATA_DIR="/data"
SUBMISSION_FILE="$DATA_DIR/submission.zip"
TEST_FILE="$DATA_DIR/unittest.zip"

echo "this line will be ignored"
echo "--- BEGIN --- INFOMARK -- WORKER"
echo ${SUBMISSION_FILE}
echo ${TEST_FILE}
echo "--- END --- INFOMARK -- WORKER"
echo "this line will be ignored"
```

as an entry point. Please either use
[one of our predefined test examples](../examples/unittests)
or create your own. By design, we assume that:

* The testing framework, for example JUnit, ensures that all stdout from the
  uploaded user code is suppressed.
* The tests are fast enough to finish before the timeout from your
  configuration file is reached.

## Overview

There are a few ways to ease the task of writing unit tests. A clear directory
structure and a Makefile to automatically pack all necessary archives and/or
run the unit tests locally can dramatically speed up the entire process and
avoid debugging on the server.

The *Makefile* should be able to clean up temporary files, create zip files,
and simulate the test result locally using the correct Docker image.
Furthermore, specifying the Docker image in the *Makefile* helps you set up the
task in InfoMark, since you need to specify it there when creating a new
exercise task. An
[example Makefile](../examples/unittests/python/makefile)
is provided in our repository.

InfoMark is language-agnostic. The system only records the Docker output. All
post-processing of runs (processing JUnit output) must be done *within* the
Docker container.

We provide several testing templates and examples:

| Language   |      Dockerimage  (hub.docker.com)     |  Test Example | Dockerfile |
|----------|:-------------|:-------:|:------:|
| Java 11 |  [patwie/test_java_submission:latest](https://hub.docker.com/r/patwie/test_java_submission) | [yes](../examples/unittests/java) | [yes](../examples/dockerimages/unittests/java) |
| Python3 |  [patwie/test_python3_submission:latest](https://hub.docker.com/r/patwie/test_python3_submission) | [yes](../examples/unittests/python) | [yes](../examples/dockerimages/unittests/python) |
| C++ |  [patwie/test_cpp_submission:latest](https://hub.docker.com/r/patwie/test_cpp_submission) | [yes](../examples/unittests/cpp) | [yes](../examples/dockerimages/unittests/cpp) |


## Java 11

We suggest using our
[Docker image](../examples/dockerimages/unittests)
and following the guide below.

We use the following directory structure to reduce the effort of writing
reusable unit tests.

```bash
exercises
  exercise<a>
    makefile
    tasks                             # any LaTeX related content describing the exercise
      sheet.tex
    solution                          # sample solution
      main
        FileA.java
        FileB.java
    student_template_[<a>.<b>]        # student-template for each task
      main
        FileA.java
        FileB.java
    unittest_public_[<a>.<b>]         # all tests without output visible to students
      src
        __unittest
          FileAtest.java
          FileBtest.java
      build.xml
    unittest_private[<a>.<b>]        # all tests without output visible only to tutors/TAs
      src
        __unittest
          FileAtest.java
          FileBtest.java
      build.xml
```

where `[a.b]` represents the exercise-task number. A working example can be
found in the
[InfoMark repository](../examples/unittests).

### Student template

The student template is usually a zip file with the exercise tasks as a PDF
and a code template. A very basic example might be:

```java
package main;

public class Hello {

  public static void main(String[] args) {
    System.out.println("6 / 2 = " + divide(6, 2));
  }

  public static int divide(int a, int b) {
    return 1;
  }
}

```

We suggest creating a code template that can be uploaded to the system as is,
so that there are no compilation errors. Testing the code template above
against the unit tests gives:

```
[javac] Compiling 3 source files to /build/classes
[   OK   ] HelloClassStructureTest:
[ FAILED ] DivideValueTest:
        Error 1/1
          - Tag: failure
          - Typ: junit.framework.AssertionFailedError
          - Msg: divide(6, 2) expected:<3> but was:<1>
```


### Tests

Public tests convey their results back to the student in our system. These are
tests that help a student make sure their solution to a programming assignment
is roughly correct. They have to be:

- verbose
- deterministic and reproducible

Private test logs are not visible to the student. Only tutors can access this
information, which should test whether the submitted solution is completely
correct by checking several corner cases. Tutors see both test results (public
and private test output).

We suggest having two subtypes of tests:

- structure tests
- value tests

We have put together a
[Helper.java](../examples/unittests/java/unittest_private_0.1/src/__unittest/Helper.java)
file to ease the work with reflection when checking the solutions to the
exercise tasks.

The overall structure should be:

```java
package __unittest;

import static org.junit.Assert.*;

import java.lang.reflect.Modifier;
import org.junit.Test;
import org.junit.rules.Timeout;
import __unittest.Helper;
import org.junit.Rule;



public class SomeNameTest {

  // We set this timeout for each single test-case.
  @Rule
  public Timeout globalTimeout = Timeout.seconds(5);

  @Test
  // ...

}

```

We *will* stop the Docker process from the worker. You need to make sure to
use a large enough timeout *inside* the InfoMark configuration file.

#### Structure tests

A structure test uses reflection to query all expected methods and check their
signatures and names. All code uploaded by any user of our system should be
accessed only via reflection, as there is no guarantee that the expected
methods exist. Rather than getting a compilation error, we would like to check
these at runtime so we can give more verbose information about what we expect.


```java
//  unittest_public[a.b]/src/__unittest/FileAtest.java
  @Test
  public void HelloClassStructureTest() {
    // no guarantee here that the class `Hello` exists --> we use Reflections
    Helper.ClassWrapper clazz = new Helper.ClassWrapper("main.Hello");
    // no guarantee here that the method `divide` exists --> we tests existence here
    // this will fail if the method does not exists (run-time error with verbose message)
    //                   name       expect (public static)          return     param 1    param 2
    clazz.mustHasMethod("divide", Modifier.PUBLIC|Modifier.STATIC, int.class, int.class, int.class);
  }
```

Running structure tests only makes sense in the public tests, so that students
get feedback if we cannot test their code because of missing methods or wrong
signatures.

If the return type of `divide` were `double` in the uploaded solution (while we
expect it to be an `int`), the output would be:

```
[ FAILED ] HelloClassStructureTest:
        Error 1/1
          - Tag: failure
          - Typ: junit.framework.AssertionFailedError
          - Msg: Method `public static double divide (int, int )` in `class Hello` found, but expected return type (`int`) is wrong. I just found `double`

```

#### Value tests

Assuming the structure is valid, we check the solution itself. Again, any wrong
method signature would be a compilation error, so we use reflection here as
well.

##### Public test

A good candidate for a public test would be:

```java
//  unittest_public[a.b]/src/__unittest/FileAtest.java
  @Test
  public void DivideValueTest() {
    Helper.InstanceWrapper actual_class = new Helper.ClassWrapper("main.Hello").create();

    assertEquals("divide(6, 2)", 3, (int) actual_class.execute("divide", 6, 2));
    assertEquals("divide(12, 3)", 4, (int) actual_class.execute("divide", 12, 3));
    // ...
  }
```

##### Private test

Corner cases should be part of the private tests, for example division by zero.
Furthermore, since the output of the public tests is visible, solutions can
overfit to those tests. A solution that passes the public test is:

```java
public static int divide(int a, int b) {
    if((a == 6) && (b==2)){
      return 3;
    }

    if((a == 12) && (b==3)){
      return 4;
    }

    return 42;
  }
```

We want to avoid this kind of solution. One approach would be to test random
numbers, but in some cases this is not possible, and we want reproducible
tests. Otherwise, learning from test output is hardly possible. The better way
to test for overfitting is to test several additional input combinations. A
good candidate for a private test would be:

```java
//  unittest_private[a.b]/src/__unittest/FileAtest.java
  @Test
  public void DivideValueTest() {
    Helper.InstanceWrapper actual_class = new Helper.ClassWrapper("main.Hello").create();

    // to avoid "overfitting"
    // this test would fail when hard-coding a solution against the public test
    assertEquals("divide(99, 3)", 33, (int) actual_class.execute("divide", 99, 3));

    // we might write in the exercise task, that x/0 should be 0
    // we should test this corner-case here as well
    assertEquals("divide(6, 0)", 0, (int) actual_class.execute("divide", 6, 0));
    assertEquals("divide(-12, 3)", -4, (int) actual_class.execute("divide", -12, 3));
    // ...
  }
```

## Python 3

We provide a very basic but working test set for checking Python programming
assignment solutions in our
[git repository](../examples/unittests/python).

The basic idea is that each upload is unzipped into a directory together with
the unit test:

```
/src
  <name>.py        # from student submission
  <name>_test.py   # from testing-framework
```

Here, again, the submission upload is extracted first, and the test framework
is extracted afterwards so that it cannot be overwritten. Tests are executed
via:

```bash
python3 -m unittest discover -s . --verbose -p '*_test.py'
```

As Python code can be written without any type annotations, you need to test
whether the methods with the correct signature exist.

For a given upload with the content:

```python
# hello.py
def divide(a, b):
  return a + b  # here is a mistake
```

A test might look like:

```python
# hello_test.py
import unittest


class Testdivide(unittest.TestCase):

  def test_divide(self):
    import hello
    self.assertTrue(hasattr(hello, 'divide'))

    if hasattr(hello, 'divide'):
      self.assertEqual(hello.divide(14, 7), 2, "Should be 2")
```

This would produce the output:

```
Python 3.7.3
test_divide (divide_test.Testdivide) ... FAIL

======================================================================
FAIL: test_divide (divide_test.Testdivide)
----------------------------------------------------------------------
Traceback (most recent call last):
  File "divide_test.py", line 11, in test_divide
    self.assertEqual(hello.divide(14, 7), 2, "Should be 2")
AssertionError: 21 != 2 : Should be 2

----------------------------------------------------------------------
Ran 1 test in 0.001s

FAILED (failures=1)
```

## C++

Testing in C++ is a bit tricky, as reflection is difficult. A basic example is
provided in our
[git repository](../examples/unittests/cpp).
The final directory structure *inside* the Docker container is:

```
/src
  lib/             # from the upload/student_template
    divide.cpp     # from the upload/student_template
    divide.hpp     # from the upload/student_template
  hello.cpp        # from the upload/student_template
  hello_test.cpp   # from the test
  catch.hpp        # from the test
  CMakeLists.txt   # from the test
  run.sh           # from the test
```

Any file from the testing zip overrides a file from the uploaded submission if
such a file exists. Furthermore, we automatically remove any `*.sh` file from
the submission in our Docker setup.

A submission consists of a main file:

```cpp
// hello.cpp
#include <stdio.h>
#include "lib/divide.h"

int main(int argc, char const *argv[]) {
  printf("%d / %d = %d\n", 6, 3, divide(6, 3));
  return 0;
}
```

and an implementation in `lib`:

```cpp
// lib/divide.cpp
#include "divide.h"

int divide(int a, int b) { return a + b; }
```

with a forward declaration:

```cpp
// lib/divide.h
#ifndef LIB_DIVIDE_H_
#define LIB_DIVIDE_H_

int divide(int a, int b);

#endif  // LIB_DIVIDE_H_
```


A simple way to test C++ implementations uses the header-only library Catch:

```cpp
#include "lib/divide.h"

#define CATCH_CONFIG_MAIN  // This tells Catch to provide a main() - only do
                           // this in one cpp file
#include "catch.hpp"

TEST_CASE("Divide should be correct", "[divide]") {
  REQUIRE(divide(6, 3) == 2);
}
```

As the implementation of `divide` is not correct, the output will be:

    Alpine clang version 5.0.1 (tags/RELEASE_501/final) (based on LLVM 5.0.1)
    Target: x86_64-alpine-linux-musl
    Thread model: posix
    InstalledDir: /usr/bin
    -- The C compiler identification is GNU 6.4.0
    -- The CXX compiler identification is GNU 6.4.0
    -- Check for working C compiler: /usr/bin/cc
    -- Check for working C compiler: /usr/bin/cc -- works
    -- Detecting C compiler ABI info
    -- Detecting C compiler ABI info - done
    -- Detecting C compile features
    -- Detecting C compile features - done
    -- Check for working CXX compiler: /usr/bin/c++
    -- Check for working CXX compiler: /usr/bin/c++ -- works
    -- Detecting CXX compiler ABI info
    -- Detecting CXX compiler ABI info - done
    -- Detecting CXX compile features
    -- Detecting CXX compile features - done
    -- Configuring done
    -- Generating done
    -- Build files have been written to: /src/build
    Scanning dependencies of target hello
    [ 33%] Building CXX object CMakeFiles/hello.dir/hello_test.cpp.o
    [ 66%] Building CXX object CMakeFiles/hello.dir/lib/divide.cpp.o
    [100%] Linking CXX executable hello
    [100%] Built target hello
    ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
    hello is a Catch v2.7.2 host application.
    Run with -? for options
    -------------------------------------------------------------------------------
    Divide should be correct
    -------------------------------------------------------------------------------
    /src/hello_test.cpp:7
    ...............................................................................
    /src/hello_test.cpp:8: FAILED:
      REQUIRE( divide(6, 3) == 2 )
    with expansion:
      9 == 2
    ===============================================================================
    test cases: 1 | 1 failed
    assertions: 1 | 1 failed

The output would also contain any linking issues (a wrongly named function),
such as:

    CMakeFiles/hello.dir/hello_test.cpp.o: In function `____C_A_T_C_H____T_E_S_T____0()':
    hello_test.cpp:(.text+0x2699e): undefined reference to `divide(double, double)'
    hello_test.cpp:(.text+0x26afd): undefined reference to `divide(double, double)'
    collect2: error: ld returned 1 exit status
    make[2]: *** [CMakeFiles/hello.dir/build.make:99: hello] Error 1
    make[1]: *** [CMakeFiles/Makefile2:68: CMakeFiles/hello.dir/all] Error 2
    make: *** [Makefile:84: all] Error 2
    /src/run.sh: line 8: ./hello: not found

> Currently, there is no way to skip non-existing methods. If a method does not
> exist or has the wrong signature, the output will contain the linking error
> without any test results from the run itself. This is caused by the nature of
> C++.
