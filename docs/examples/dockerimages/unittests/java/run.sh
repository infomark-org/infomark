#!/bin/bash
set -e

HOME_DIR="/home/simple_ci_runner"
RESULT_DIR="/results"
DATA_DIR="/data"

SUBMISSION_FILE="$DATA_DIR/submission.zip"
TEST_FILE="$DATA_DIR/unittest.zip"

# unpack submission
cd ${HOME_DIR}/java/src
unzip -o ${SUBMISSION_FILE} || true

# consider the existance of a __unittest folder as attack!
lines=$(find ${HOME_DIR}/java/src/ -type d -name "__unittest" | wc -l)
if [[ ! $lines -eq 0 ]]; then
  echo "--- BEGIN --- INFOMARK -- WORKER"
  echo "Please remove the __unittest folder(s) from your submission."
  echo "Writing unit-tests is our job not yours :-)"
  echo "We found the following directories:"
  find ${HOME_DIR}/java/src/ -type d -name "__unittest" | sed -E 's/\/home\/simple_ci_runner\/java\/src\//  - /g'
  echo "Please remove them!"
  echo "--- END --- INFOMARK -- WORKER"
  exit 0
  # TODO(): discuss if we simply remove these directories
  # rm "${HOME_DIR}/java/src/__unittest"
fi

# remove black listes files
find ${HOME_DIR}/java/src/ -name "module-info.java" -type f -delete


# unzip tests
cd ${HOME_DIR}/java
unzip -o ${TEST_FILE} >/dev/null 2>&1
chmod -R 555 ${HOME_DIR}/java/src/__unittest

# run unit tests
ant junit > /tmp/compile.log || true
cat /tmp/compile.log

# run report translator
chmod o+r /tmp/compile.log

echo "--- BEGIN --- INFOMARK -- WORKER"
java --version
javac --version
PYTHONIOENCODING=UTF-8 python3 parse_compiler.py
PYTHONIOENCODING=UTF-8 python3 junit_xml_report_translator.py
echo "--- END --- INFOMARK -- WORKER"
