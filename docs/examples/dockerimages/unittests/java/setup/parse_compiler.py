#!/usr/bin/python3
# -*- coding: utf-8 -*-
with open("/tmp/compile.log", "r") as f:
  for g in f:
    g = g.strip()

    ignores = ['-Xlint', '__unittest']

    ignore = False
    for i in ignores:
      if i in g:
        ignore = True
        break

    if not ignore:
      if g.startswith("[javac]"):
        print(g.replace('/home/simple_ci_runner/java', ''))
