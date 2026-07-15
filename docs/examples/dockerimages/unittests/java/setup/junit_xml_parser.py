#!/usr/bin/python3
# -*- coding: utf-8 -*-
import xml.etree.ElementTree as ET
import sys


def indent(elem, level=0):
  i = "\n" + level * "  "
  j = "\n" + (level - 1) * "  "
  if len(elem):
    if not elem.text or not elem.text.strip():
      elem.text = i + "  "
    if not elem.tail or not elem.tail.strip():
      elem.tail = i
    for subelem in elem:
      indent(subelem, level + 1)
    if not elem.tail or not elem.tail.strip():
      elem.tail = j
  else:
    if level and (not elem.tail or not elem.tail.strip()):
      elem.tail = j
  return elem


if len(sys.argv) < 2:
  exit(-1)

# print(sys.argv[1])

tree = ET.parse(sys.argv[1])
root = tree.getroot()
root_out = ET.Element(root.tag, root.attrib)

for cases in root.iter('testcase'):
  # print(cases.attrib['name'])
  elem = ET.SubElement(root_out, 'test')
  elem.set('name', cases.attrib['name'])
  if len(cases) == 0:
    elem.set('testpassed', 'true')
  else:
    elem.set('testpassed', 'false')
    elem.set('type', cases[0].tag)
    elem.set('exectype', cases[0].attrib['type'])
    elem.set('message', cases[0].attrib['message'])

    # print(dir(cases[0]))
    # print(cases[0].attrib)
    #
    # for child in cases:
    #     print(child.attrib)


indent(root_out)

out_tree = ET.ElementTree(root_out)
out_tree.write('file.xml')
