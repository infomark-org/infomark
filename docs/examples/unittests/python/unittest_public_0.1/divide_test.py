import unittest


class Testdivide(unittest.TestCase):

  def test_divide(self):
    import hello
    self.assertTrue(hasattr(hello, 'divide'))

    if hasattr(hello, 'divide'):
      self.assertEqual(hello.divide(6, 2), 3, "Should be 3")


if __name__ == '__main__':
  unittest.main()
