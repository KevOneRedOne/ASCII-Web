# The base go-image
FROM golang:1.15.6
 
# Create a directory for the app
RUN mkdir /ascii-web-docker
 
# Copy all files from the current directory to the app directory
COPY . /ascii-web-docker
 
# Set working directory
WORKDIR /ascii-web-docker
 
# Run command as described:
# go build will build an executable file named server in the current directory
RUN go build -o server . 
 
# Run the server executable
CMD [ "/ascii-web-docker/server" ]

#Run the docker in the command prompt:
#1
#Build the program
#docker build -t application-tag .

#2
#Run the docker
#docker run -it --rm -p 8081:8080 application-tag