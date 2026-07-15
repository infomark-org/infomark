#! /usr/bin/env python3
# -*- coding: utf-8 -*-

"""
Some example api calls. This requires python3!
"""

from infomark.infomark import Infomark


client = Infomark('http://localhost:2020')

# will not work (401 error as we are not logged in)
print(client.get("/api/v1/account"))

# login
client.login('test@uni-tuebingen.de', 'test')

# will print the information from the login identity
response = client.get("/api/v1/account")
print(response.status_code)
print(response.json())

query = 'iet'
users = client.get("/api/v1/users/find?query=%s" % query).json()
for user in users:
  print(user['first_name'], user['last_name'], user['email'])
