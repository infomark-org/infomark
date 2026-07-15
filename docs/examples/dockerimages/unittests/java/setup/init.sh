#!/bin/bash
set -e

WORK_DIR=/home/simple_ci_runner/java


cd /app/setup

mkdir -p ${WORK_DIR}/src
mkdir -p ${WORK_DIR}/lib

cp -R junit-4.12.jar ${WORK_DIR}/lib
cp -R hamcrest-core-1.3.jar ${WORK_DIR}/lib
cp -R junit_xml_report_translator.py ${WORK_DIR}
cp -R parse_compiler.py ${WORK_DIR}

chown -R simple_ci_runner:simple_ci_runner ${WORK_DIR}
chmod o+rwx -R ${WORK_DIR}
