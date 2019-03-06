#!/usr/bin/env python3
from http.server import HTTPServer, SimpleHTTPRequestHandler, test
import sys
import os


class CORSRequestHandler (SimpleHTTPRequestHandler):
  def end_headers(self):
    self.send_header('Access-Control-Allow-Origin', '*')
    SimpleHTTPRequestHandler.end_headers(self)


if __name__ == '__main__':
  port = int(sys.argv[1]) if len(sys.argv) > 1 else 8000
  url_fmt = 'http://localhost:%i/swagger/index.html'
  print('')
  print('Enpoint at:')
  print(url_fmt % (port))
  print('')
  print('')
  test(CORSRequestHandler, HTTPServer, port=port)
