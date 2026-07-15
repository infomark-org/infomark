package __unittest;

import static org.junit.Assert.*;

import java.lang.reflect.Modifier;

import org.junit.Test;
import org.junit.rules.Timeout;
import __unittest.Helper;
import org.junit.Rule;



public class HelloTest {

  // We set this timeout for each single test-case.
  @Rule
  public Timeout globalTimeout = Timeout.seconds(5);

  // first we always check if the uploads of the students have all necessary
  // methods. We need to do this in a two fold process:
  // 1. check static methods manually
  // 2. check the rest by relying on some reflection stuff and testing the submission against
  //    the class in "Expected" (above).
  @Test
  public void HelloClassStructureTest() {
    // 1. check static methods
    Helper.ClassWrapper clazz = new Helper.ClassWrapper("main.Hello");
    clazz.mustHasMethod("divide", Modifier.PUBLIC|Modifier.STATIC, int.class, int.class, int.class);
  }


  @Test
  public void DivideValueTest() {
    // we cannot guarantee this exists (hence we use our helper class)
    // this would fail otherwise.
    Helper.InstanceWrapper actual_class = new Helper.ClassWrapper("main.Hello").create();

    assertEquals("divide(6, 2)", 3, (int) actual_class.execute("divide", 6, 2));
    assertEquals("divide(12, 3)", 4, (int) actual_class.execute("divide", 12, 3));

  }


}
