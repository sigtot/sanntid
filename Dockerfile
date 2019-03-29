FROM ubuntu
RUN ["apt-get", "update"]
RUN ["apt-get", "install", "-y", "tmux", "ssh", "golang-go", "git"]
RUN mkdir go
RUN GOPATH=/root/go; GOROOT=/usr/lib/go; go get github.com/sirupsen/logrus go.etcd.io/bbolt github.com/sigtot/elevio
RUN ls /root/go
RUN yes pass | adduser elev
# Don't write dockerfile past midnight kids
RUN chmod -R 777 /root
RUN touch /orderwatcher.db && chmod 777 /orderwatcher.db
RUN adduser elev root
COPY entrypoint.sh /
COPY simelevserver /
WORKDIR /
RUN ["chmod", "777", "entrypoint.sh", "simelevserver"]
CMD /entrypoint.sh
