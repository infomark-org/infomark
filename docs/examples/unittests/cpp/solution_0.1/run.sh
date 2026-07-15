mkdir build
cd build
cmake .. &> compile.log
make &>> compile.log
cat compile.log
./hello