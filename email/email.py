import os


def sendMail():
  sendmail_location = "/usr/sbin/sendmail"  # sendmail location
  p = os.popen("%s -t" % sendmail_location, "w")
  p.write("From: %s\n" % "no-reply@info2.informatik.uni-tuebingen.de")
  p.write("To: %s\n" % "patrick.wieschollek@uni-tuebingen.de")
  p.write("Subject: TestSubject\n")
  p.write("\n")  # blank line separating headers from body
  p.write("body of the mail")
  status = p.close()
  if status != 0:
    print("Sendmail exit status", status)


sendMail()