#!/bin/bash
set -e

WORK_DIR=/home/simple_ci_runner/python


cd /app/setup

mkdir -p ${WORK_DIR}/src
chown -R simple_ci_runner:simple_ci_runner ${WORK_DIR}
chmod o+rwx -R ${WORK_DIR}