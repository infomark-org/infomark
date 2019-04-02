FROM ubuntu:18.04

RUN apt-get update
RUN apt-get -y install git subversion gcc g++ make wget gfortran
RUN apt-get -y install golang

RUN cd /var
RUN svn checkout --non-interactive --trust-server-cert https://projects.coin-or.org/svn/SYMPHONY/releases/5.6.13 /var/symphony
RUN cd /var/symphony/ThirdParty/Glpk && ./get.Glpk
RUN cd /var/symphony && ./configure --with-gmpl &&  make -j4 && make install


ADD webapp /go/solver
ENTRYPOINT /go/solver/entry.sh

EXPOSE 9090
