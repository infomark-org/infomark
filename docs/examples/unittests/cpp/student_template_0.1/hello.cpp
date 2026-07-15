#include <stdio.h>
#include "lib/divide.h"

int main(int argc, char const *argv[]) {
  printf("%d / %d = %d\n", 6, 3, divide(6, 3));
  return 0;
}
