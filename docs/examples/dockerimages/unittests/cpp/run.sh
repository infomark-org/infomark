#!/usr/bin/env sh

#
cd /src
unzip /data/submission.zip
unzip  -o /data/unittest.zip
chmod x+r /src/run.sh
echo "--- BEGIN --- INFOMARK -- WORKER"
clang++ --version
/src/run.sh
echo "--- END --- INFOMARK -- WORKER"