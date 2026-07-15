#include "lib/divide.h"

#define CATCH_CONFIG_MAIN  // This tells Catch to provide a main() - only do
                           // this in one cpp file
#include "catch.hpp"

TEST_CASE("Divide should be correct", "[divide]") {
  REQUIRE(divide(6, 3) == 2);
  REQUIRE(divide(14, 7) == 2);
}