import unittest


class Testdivide(unittest.TestCase):

  def test_divide(self):
    import hello
    self.assertTrue(hasattr(hello, 'divide'))

    if hasattr(hello, 'divide'):
      self.assertEqual(hello.divide(14, 7), 2, "Should be 2")


if __name__ == '__main__':
  unittest.main()
