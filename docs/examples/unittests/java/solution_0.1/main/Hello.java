package main;

public class Hello {
  public static void main(String[] args) {
    System.out.println("Hello World!");
    System.out.println("6 / 2 = " + divide(6, 2));
  }

  public static int divide(int a, int b) {
    if (b == 0) {
      return 0;
    } else {
      return a / b;
    }
  }
}
