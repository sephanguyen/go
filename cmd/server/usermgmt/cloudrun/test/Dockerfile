# Use the official Ubuntu image from Docker Hub as
# a base image
FROM ubuntu:22.04 AS runner

RUN apt-get update
RUN apt-get install sshpass -y

RUN mkdir /root/.ssh && chmod 0700 /root/.ssh
RUN chmod 700 /root/.ssh

RUN ssh-keygen -t rsa -b 2048 -C "phison.huynh@manabie.com" -f $HOME/.ssh/id_rsa

#RUN chmod 600 /root/.ssh/authorized_keys
RUN ssh-keyscan -t rsa github.com >> /root/.ssh/known_hosts

# Execute next commands in the directory /workspace
WORKDIR /workspace

# Copy over the script to the /workspace directory
COPY script.sh .

# Just in case the script doesn't have the executable bit set
RUN chmod +x ./script.sh

# Run the script when starting the container
CMD [ "./script.sh" ]
