cmake_minimum_required(VERSION 3.0)
project(hello)

add_executable(hello hello_test.cpp lib/divide.cpp)


target_compile_options(hello PUBLIC -std=c++1y -Wall -Wfloat-conversion)