// InfoMark - a platform for managing courses with
//            distributing exercise sheets and testing exercise submissions
// Copyright (C) 2019  ComputerGraphics Tuebingen
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package __unittest;

import static org.junit.Assert.fail;

import java.lang.reflect.Constructor;
import java.lang.reflect.Field;
import java.lang.reflect.Method;
import java.lang.reflect.Modifier;
import java.lang.reflect.Type;
import java.io.ByteArrayOutputStream;
import java.io.PrintStream;
import java.util.*;

public class Helper {

  public static class Pair<A, B> {
    public A a;
    public B b;

    public Pair(A a, B b) {
      this.a = a;
      this.b = b;
    }
  }

  public static class Triplet<A, B, C> {
    public A a;
    public B b;
    public C c;

    public Triplet(A a, B b, C c) {
      this.a = a;
      this.b = b;
      this.c = c;
    }
  }

  public static class Quadruple<A, B, C, D> {
    public A a;
    public B b;
    public C c;
    public D d;

    public Quadruple(A a, B b, C c, D d) {
      this.a = a;
      this.b = b;
      this.c = c;
      this.d = d;
    }
  }



  // return "Integer" from "java.class.integer"
  private static String getNameAfterDot(Object str) {
    String in = "" + str;
    String[] s = in.split("\\.");
    return clean_array_types(s[s.length - 1]);
  }

  public static String clean_array_types(String parameters){
      // https://docs.oracle.com/javase/9/docs/api/java/lang/Class.html#getName--
      parameters = parameters.replaceAll("class \\[Z", "boolean[]");
      parameters = parameters.replaceAll("class \\[B", "byte[]");
      parameters = parameters.replaceAll("class \\[S", "short[]");
      parameters = parameters.replaceAll("class \\[I", "int[]");
      parameters = parameters.replaceAll("class \\[J", "long[]");
      parameters = parameters.replaceAll("class \\[F", "float[]");
      parameters = parameters.replaceAll("class \\[D", "double[]");
      parameters = parameters.replaceAll("class \\[C", "char[]");
      return parameters;
    }

  // join a collection into a string with delimiter ","
  public static String join(Object[] s) {
    int l = s.length;
    if (l == 0) {
      return "";
    }
    String str = "" + Helper.getNameAfterDot(s[0]);
    for (int i = 1; i < l; i++) {
      str += ", " + Helper.getNameAfterDot(s[i]);
    }
    return clean_array_types(str);
  }

  // guess the type of the parameters
  protected static Class<?>[] types(Class<?>... values) {
    if (values == null) {
      return new Class[0];
    }

    Class<?>[] result = new Class[values.length];

    for (int i = 0; i < values.length; i++) {
      Object value = values[i];
      result[i] = (value == null) ? null : value.getClass();
    }

    return result;
  }

  // guess the type of the parameters
  protected static Class<?>[] types(Object... values) {
    if (values == null) {
      return new Class[0];
    }

    Class<?>[] result = new Class[values.length];

    for (int i = 0; i < values.length; i++) {
      Object value = values[i];
      result[i] = (value == null) ? null : value.getClass();
    }

    return result;
  }

  // replace primitives
  public static Class<?> wrapper(Class<?> type) {
    if (type == null) {
      return null;
    } else if (type.isPrimitive()) {
      if (boolean.class == type) {
        return Boolean.class;
      } else if (int.class == type) {
        return Integer.class;
      } else if (long.class == type) {
        return Long.class;
      } else if (short.class == type) {
        return Short.class;
      } else if (byte.class == type) {
        return Byte.class;
      } else if (double.class == type) {
        return Double.class;
      } else if (float.class == type) {
        return Float.class;
      } else if (char.class == type) {
        return Character.class;
      } else if (void.class == type) {
        return Void.class;
      }
    }

    return type;
  }

  private static boolean matchParameters(Class<?>[] actual_types, Class<?>[] expected_types) {
    if (actual_types.length == expected_types.length) {
      for (int i = 0; i < expected_types.length; i++) {
        if (expected_types[i] == null) {
          continue;
        }

        if (wrapper(actual_types[i]).isAssignableFrom(wrapper(expected_types[i]))) {
          continue;
        }

        return false;
      }

      return true;
    } else {
      return false;
    }
  }

  private static boolean matchSignature(
      Method candidate_method, String expected_name, Class<?>[] expected_params) {
    return (candidate_method.getName().equals(expected_name)
        && Helper.matchParameters(candidate_method.getParameterTypes(), expected_params));
  }

  private static boolean matchSignature(Method candidate_method, Method expected_method) {
    return (candidate_method.getName().equals(expected_method.getName())
        && Helper.matchParameters(
               candidate_method.getParameterTypes(), expected_method.getParameterTypes()));
  }

  // --------------------------------------------------------
  public static class MethodWrapper {
    // the atual method
    Method methodHnd;
    // link to class
    Helper.ClassWrapper classHnd;

    // return raw method
    public Method get() {
      return methodHnd;
    }



    // nice printout like "do_something(Integer, Integer)" (static)
    public static String toString(
        String method_name, int expected_modifier, Type expected_returnType, Class<?>... values) {
      String str = Modifier.toString(expected_modifier);
      str += " " + Helper.getNameAfterDot(expected_returnType);
      str += " " + method_name;
      str += " (";
      if (values.length > 0){
        str += Helper.join(values);
      }
      str += " )";

      return str;
    }

    public static String isMissingMessage(
        String method_name, int expected_modifier, Type expected_returnType, Class<?>... values) {
      return "We could not find the method `"
          + Helper.MethodWrapper.toString(
                method_name, expected_modifier, expected_returnType, values)
          + "`.";
    }

    // nice printout like "do_something(Integer, Integer)"
    public String toString() {
      String str = Modifier.toString(methodHnd.getModifiers());
      str += " " + Helper.getNameAfterDot(methodHnd.getGenericReturnType());
      str += " " + name();
      str += " (";

      str += Helper.join(methodHnd.getGenericParameterTypes());

      str += " )";

      return str;
    }

    public MethodWrapper(Helper.ClassWrapper cls, Method hnd) {
      methodHnd = hnd;
      classHnd = cls;
    }

    public MethodWrapper(Method hnd) {
      methodHnd = hnd;
      classHnd = new Helper.ClassWrapper(hnd.getClass().getName());
    }

    // checks if current method match signature (does NOT compare multiple methods)!
    public boolean checkSignature(
        int expected_modifier, Type expected_returnType, Type... expected_params) {
      if (!checkModifier(expected_modifier)) {
        fail("Method `" + toString() + "` in `class " + classHnd.name()
            + "` found, but expected modifier `" + Modifier.toString(expected_modifier)
            + "` is wrong. I just found `" + Modifier.toString(methodHnd.getModifiers()) + "`");
      }

      if (!checkReturnType(expected_returnType)) {
        fail("Method `" + toString() + "` in `class " + classHnd.name()
            + "` found, but expected return type (`" + expected_returnType
            + "`) is wrong. I just found `" + methodHnd.getGenericReturnType() + "`");
      }

      try {
        if (!checkParameters(expected_params)) {
          throw new Exception("");
        }

      } catch (Exception e) {
        fail("Method `" + toString() + "` in `class " + classHnd.name()
            + "` found, but parameters do not match. We expect " + expected_params.length
            + " parameter(s) {`" + Helper.join(expected_params) + "`"
            + "} but found " + methodHnd.getGenericParameterTypes().length + " parameter(s) {`"
            + Helper.join(methodHnd.getGenericParameterTypes()) + "`}.");
      }

      return true;
    }

    // checks if current return type match expected return type
    public boolean checkReturnType(Type expected) {
      Type actual = methodHnd.getGenericReturnType();

      return expected.equals(actual);
    }

    // check if current modifers match expected modifiers
    public boolean checkModifier(int expected) {
      int actual = methodHnd.getModifiers();
      return (actual == expected);
    }

    // check if current parameters match expected parameters
    public boolean checkParameters(Type[] parameters) throws Exception {
      Type[] gpType = methodHnd.getGenericParameterTypes();

      if (gpType.length != parameters.length) {
        throw new Exception("");
      }

      boolean cont = true;
      for (int i = 0; i < gpType.length && cont; i++) {
        cont = gpType[i].getClass().getName().equals(parameters[i].getClass().getName());
      }

      if (!cont)
        throw new Exception("");
      return cont;
    }

    public String name() {
      return methodHnd.getName();
    }
  }

  // --------------------------------------------------------
  // represents an instantiated class
  public static class InstanceWrapper {
    Object obj;
    ClassWrapper cls;

    public Object get() {
      return obj;
    }

    public boolean implementsInterface(ClassWrapper cls) {
      return (cls.get().isInstance(obj));
    }

    public InstanceWrapper(ClassWrapper cls, Object obj) {
      this.cls = cls;
      this.obj = obj;
    }

    // accesses a field
    public <T> T getField(String fieldName) {
      try {
        Field f = cls.classHnd.getDeclaredField(fieldName);
        f.setAccessible(true);
        return (T) f.get(obj);
      } catch (NoSuchFieldException e) {
        fail("no such field:"
            + "there exists no field `" + fieldName + "` in `class " + cls.name() + "`");
      } catch (SecurityException e) {
        //
        fail("An error occured.");
      } catch (IllegalArgumentException e) {
        //
        fail("An error occured.");
      } catch (IllegalAccessException e) {
        //
        fail("An error occured.");
      }
      return null;
    }

    // excute method of object
    public <T> T execute(MethodWrapper m, Object... parameters) {
      try {
        return (T) m.get().invoke(obj, parameters);
        // check the state change of obj at the calling site
      } catch (Exception e) {
        // fail(message);
        return null;
      }
    }

    public <T> T execute(String str, Object... parameters) {
      try {
        return (T) cls.getMethod(str).get().invoke(obj, parameters);
        // check the state change of obj at the calling site
      } catch (Exception e) {
        // fail(message);
        return null;
      }
    }

    public <T> T execute(MethodWrapper m) {
      try {
        return (T) m.get().invoke(obj);
        // check the state change of obj at the calling site
      } catch (Exception e) {
        // fail(message);
        return null;
      }
    }

    public <T> T execute(String str) {
      try {
        return (T) cls.getMethod(str).get().invoke(obj);
        // check the state change of obj at the calling site
      } catch (Exception e) {
        // fail(message);
        return null;
      }
    }
  }
  // --------------------------------------------------------
  // represents a CLASS not an OBJECT instantiated by this class
  public static class ClassWrapper {
    Class<?> classHnd;

    public Type type() {
      return classHnd;
    }

    public Class<?> get() {
      return classHnd;
    }

    public Helper.InstanceWrapper create() {
      Constructor<?> ctor = classHnd.getDeclaredConstructors()[0];
      try {
        return new Helper.InstanceWrapper(this, ctor.newInstance());
      } catch (Exception e) {
        fail("An Exception occured on invoking the constructor of `class " + name()
            + "`: " + e.getMessage());
        return null;
      }
    }

    public Helper.InstanceWrapper create(Object... args) {
      Constructor<?> ctor = null;
      try {
        ctor = constructor(Modifier.PUBLIC, Helper.types(args));
        try {
          if (args.length == 0) {
            return new Helper.InstanceWrapper(this, ctor.newInstance());
          } else {
            return new Helper.InstanceWrapper(this, ctor.newInstance(args));
          }
        } catch (Exception e) {
          fail("An Exception occured on invoking the constructor of `class " + name()
              + "` with parameters (`" + Helper.join(args) + "`) of type (`"
              + Helper.join(Helper.types(args)) + "`).");
          return null;
        }
      } catch (Exception e) {
        fail("We did not find a constructor for `class " + name() + "` with parameters (`"
            + Helper.join(args) + "`) of type (`" + Helper.join(Helper.types(args)) + "`).");
      }

      return null;
    }

    public String name() {
      return Helper.getNameAfterDot(classHnd.getName());
    }

    public ClassWrapper(String class_with_package_name) {
      try {
        classHnd = Class.forName(class_with_package_name);
      } catch (ClassNotFoundException e) {
        fail("class not found:`" + class_with_package_name + "`");
        classHnd = null;
      }
    }
    public ClassWrapper(Class<?> cls) {
      classHnd = cls;
    }

    public void shouldMatch(String clazz) {
      shouldMatch(new ClassWrapper(clazz));
    }
    public void shouldMatch(ClassWrapper expected) {
      // get all methods from expected
      try {
        Method[] allExpectedMethods = expected.get().getDeclaredMethods();
        Method[] allActualMethods = get().getDeclaredMethods();

        String msg = "";

        for (Method expected_method : allExpectedMethods) {
          boolean exists = false;
          // find exact match in current class
          for (Method candidate_method : allActualMethods) {
            if (!candidate_method.getName().equals(expected_method.getName())) {
              continue;
            } else {
              // check signature
              if (!Helper.matchSignature(expected_method, candidate_method)) {
                continue;
              }
              exists = true;
            }
          }
          if (!exists) {
            msg += "                  -  " + (new Helper.MethodWrapper(this, expected_method)).toString() + "\n";
            // msg += "you implementation in `class " + name() + "`. But it is not there.";
            // "Could not retrieve method `" + (new Helper.MethodWrapper(this,
            // expected_method)).toString() + "` in `class " + name() + "`."

          }
        }

        if(msg.length() > 0){
          fail("We expected to find in class `"+ name()  +"` the following method(s):\n"+ msg+
            "                 But we could not find them.");
        }

      } catch (Exception e) {
        fail("Could not retrieve methods  in `class " + name() + "`.");
      }
    }

    public Object field(Object instance, String fieldName) {
      try {
        Field f = classHnd.getDeclaredField(fieldName);
        f.setAccessible(true);
        return f.get(instance);
      } catch (NoSuchFieldException e) {
        fail("no such field:"
            + "there exists no field `" + fieldName + "` in `class " + name() + "`");
      } catch (SecurityException e) {
        //
        fail("An error occured.");
      } catch (IllegalArgumentException e) {
        //
        fail("An error occured.");
      } catch (IllegalAccessException e) {
        //
        fail("An error occured.");
      }
      return null;
    }

    public Constructor<?> constructor(int modifier, Class<?>[] parameters) throws Exception {
      try {
        Constructor<?> constructor = classHnd.getDeclaredConstructor(parameters);
        return constructor;
      }
      // ein Versuch war es wert!
      catch (NoSuchMethodException e) {
        // iterate over all construtors
        for (Constructor<?> constructor : classHnd.getDeclaredConstructors()) {
          if (Helper.matchParameters(constructor.getParameterTypes(), parameters)) {
            return constructor;
          }
        }
        throw new Exception("");
      }
    }

    /**
     * @param name of method
     * @return returns whether a method with a specified name exists
     */
    public boolean mustHasMethod(String method_name) {
      return (getMethod(method_name) != null);
    }

    /**
     * @param name of method
     * @param expected modifiers
     * @param expected return type
     * @param expected arguments
     * @return returns whether such a method exists
     */
    public boolean mustHasMethod(
        String method_name, int expected_modifier, Type expected_returnType, Class<?>... values) {
      String msg = "In your class `class " + name() + "`: ";
      msg += Helper.MethodWrapper.isMissingMessage(
          method_name, expected_modifier, expected_returnType, values);
      msg += " But we expected to see it there.";

      try {
        Method[] allMethods = classHnd.getDeclaredMethods();
        for (Method m : allMethods) {
          // name match?
          if (!m.getName().equals(method_name)) {
            continue;
          }
          if (values.length == 0) {
            // we expect no arguments
            if (m.getGenericParameterTypes().length != 0) {
              continue;
            }
          }

          MethodWrapper method = new MethodWrapper(this, m);

          if (!method.checkSignature(expected_modifier, expected_returnType, values))
            continue;

          return true;
        }
        fail(msg);
        return false;
      } catch (Exception e) {
        fail(msg);
        return false;
      }
    }

    public Helper.MethodWrapper getMethod(String method_name) {
      try {
        Method[] allMethods = classHnd.getDeclaredMethods();
        for (Method m : allMethods) {
          if (!m.getName().equals(method_name)) {
            continue;
          }
          return new Helper.MethodWrapper(this, m);
        }
      } catch (Exception e) {
        fail("Could not retrieve method `" + method_name + "`.");
      }

      fail("Could not find method `" + method_name + "`.");
      return null;
    }

    public Helper.MethodWrapper getMethod(
        String method_name, int expected_modifier, Type expected_returnType, Class<?>... values) {
      String msg = Helper.MethodWrapper.isMissingMessage(
          method_name, expected_modifier, expected_returnType, values);
      try {
        Method[] allMethods = classHnd.getDeclaredMethods();
        for (Method m : allMethods) {
          // name match?
          if (!m.getName().equals(method_name)) {
            continue;
          }
          if (values.length == 0) {
            // we expect no arguments
            if (m.getGenericParameterTypes().length != 0) {
              continue;
            }
          }
          // parameters match?
          if (!Helper.matchSignature(m, m.getName(), values)) {
            continue;
          }
          return new Helper.MethodWrapper(this, m);
        }

        fail(msg);
        return null;
      } catch (Exception e) {
        fail(msg);
        return null;
      }
    }
  }

  public static class PackageWrapper {
    public static void shouldMatch(String checkPackage, String layoutClass) {
      Helper.ClassWrapper superClass = new Helper.ClassWrapper(layoutClass);
      System.err.println(superClass.name());
      System.err.println(superClass.name());
      System.err.println(superClass.name());

      Class<?>[] d = superClass.get().getDeclaredClasses();

      System.err.println(checkPackage);
      System.err.println(layoutClass);
      System.err.println(d.length);
      for (Class<?> cls : d) {
        String their_class_name = checkPackage + "." + cls.getName().replace(layoutClass + "$", "");
        System.err.println(their_class_name);

        Helper.ClassWrapper classToCheck = new Helper.ClassWrapper(their_class_name);
        Helper.ClassWrapper correctLayout = new Helper.ClassWrapper(cls);
        classToCheck.shouldMatch(correctLayout);
      }
    }
  }

  public static class StreamRecorder {

    private final ByteArrayOutputStream outContent = new ByteArrayOutputStream();
    private final PrintStream originalOut = System.out;

    public void start() {
      outContent.reset();
      System.setOut(new PrintStream(outContent));
    }

    public String stop() {
      String captured = outContent.toString();
      System.setOut(originalOut);
      return captured.trim();
      // return "kk";
    }
  }

  public static class StreamTester {

    public static void checkOutput(String label, String actual, String expected) {
      String[] actual_lines = actual.split("[\\r\\n]+");
      String[] expected_lines = expected.split("[\\r\\n]+");

      List<String> errorMsgs = new ArrayList<String>();

      for (int i = 0; i < expected_lines.length; ++i) {
        String expected_line = expected_lines[i].trim();
        String actual_line = "";

        if (i < actual_lines.length ){
          actual_line = actual_lines[i].trim();
        }

        if(!expected_line.equals(actual_line)){
          String errorMsg = "Output-line "+ String.valueOf(i) + " of `"+label+"` is wrong. We expected `"+expected_line+"` but got `"+actual_line+"`.";
          errorMsgs.add(errorMsg);
        }


      }

      int max_shown = 10;
      int k = errorMsgs.size();
      if ( k > max_shown ){
        errorMsgs.subList(max_shown, k).clear();
        errorMsgs.add("... ");
      }

      if(errorMsgs.size() > 0){
        fail("The output does not match our expectation\n" + String.join("\n", errorMsgs));
      }
    }
  }
}
