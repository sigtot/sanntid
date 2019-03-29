FROM ubuntu
RUN ["apt-get", "update"]
RUN ["apt-get", "install", "-y", "tmux", "ssh", "sshpass", "nmap"]
COPY entrypoint.sh /
WORKDIR /
CMD /entrypoint.sh
