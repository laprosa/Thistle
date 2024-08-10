# Thistle

Thistle is simplistic based malware that offers a CLI interface over an already secure implentation: ssh

All communication between the client and server is done over encrypted websockets, with the option to download files. It's intended to be simple to keep the code small, and detections relatively low.

![Main display](https://i.imgur.com/Z7ASno5.png)

### Rough Detections as of 5/8/24

![enter image description here](https://i.imgur.com/WqQOnAq.png)

## Setup

To generate your private key for ssh:
- open a terminal in /Thistle/Server
- run the following: openssl genpkey -algorithm RSA -out private.key -pkeyopt rsa_keygen_bits:2048

To generate your self signed certificates:
- reuse the same terminal or open one again in the same server folder
- run the following commands: 
openssl genpkey -algorithm RSA -out server.key -pkeyopt rsa_keygen_bits:2048
openssl req -new -key private.key -out server.csr
openssl x509 -req -days 3650 -in server.csr -signkey server.key -out server.crt

modify the username and password inside users.json as needed.
to create a new bcrypt hashed password I just use : https://bcrypt-generator.com/

run the following command to keep the server running after you disconnect:
- screen -dms Thistle ./thistle_server
if you get permission denied error just run chmod +x thisle_server then retry

to connect to the server, run this command:
ssh -T username@serverip -p2222
^ That's for linux adapt as needed for putty or ssh on windows

Client setup:
- modify the address in main.go from localhost to whatever the server ip is, or your domain if you are using one.
- open a terminal in Thistle/Client/Windows
- run go build (This will produce a GUI for testing purposes)
- ensure connectivity, then rebuild without a gui:
- go build -ldflags="-H=windowsgui"
- if you are building from linux run it like this:
- GOOS=windows go build -ldflags="-H=windowsgui"



