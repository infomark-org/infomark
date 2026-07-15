#!/bin/bash
set -e

HOME_DIR="/home/simple_ci_runner"
RESULT_DIR="/results"
DATA_DIR="/data"

SUBMISSION_FILE="$DATA_DIR/submission.zip"
TEST_FILE="$DATA_DIR/unittest.zip"

# unpack submission
cd ${HOME_DIR}/python/src
unzip ${SUBMISSION_FILE}

# consider the existance of a __unittest folder as attack!
lines=$(find ${HOME_DIR}/python/src/ -type f -name "*_test.py" | wc -l)
if [[ ! $lines -eq 0 ]]; then
  echo "--- BEGIN --- INFOMARK -- WORKER"
  echo "Please remove the *_test.py files from your submission."
  echo "Writing unit-tests is our job not yours :-)"
  echo "We found the following directories:"
  find ${HOME_DIR}/python/src/ -type d -name "*_test.py" | sed -E 's/\/home\/simple_ci_runner\/python\/src\//  - /g'
  echo "Please remove them!"
  echo "--- END --- INFOMARK -- WORKER"
  exit 0
fi

# unzip tests
cd ${HOME_DIR}/python/src
unzip ${TEST_FILE} >/dev/null 2>&1

python3 -m unittest discover -s . --verbose -p '*_test.py' &> /tmp/test.log || true

echo "--- BEGIN --- INFOMARK -- WORKER"
python3 --version
cat /tmp/test.log | sed -E 's/\/home\/simple_ci_runner\/python\/src\///g'
echo "--- END --- INFOMARK -- WORKER"
