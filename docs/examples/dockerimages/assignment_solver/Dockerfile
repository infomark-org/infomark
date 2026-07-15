FROM ubuntu:18.04

MAINTAINER Patrick Wieschollek <mail@patwie.com>

RUN apt-get update && \
  apt-get install -y --no-install-recommends \
  git  \
  subversion  \
  gcc  \
  g++  \
  make  \
  wget  \
  gfortran \
  && rm -rf /var/lib/apt/lists/*


RUN cd /var
RUN svn checkout --non-interactive --trust-server-cert https://projects.coin-or.org/svn/SYMPHONY/releases/5.6.13 /var/symphony
RUN cd /var/symphony/ThirdParty/Glpk
RUN cd /var/symphony/ThirdParty/Glpk && ./get.Glpk
RUN cd /var/symphony && ./configure --with-gmpl &&  make -j4 && make install
