### First install and setup all the prerequesties that is required for our project setup

Run hyperledger fabric test network for project implementation 

## Admin application 
### Clone the integra-admin-backend repository from the particular github account

Go to the integra-admin-backend/integra-nock-sdk diretory:

`cd integra-admin-backend/integra-nock-sdk`

After that run the following command to download required go packages locally:

`go mod vendor`

After that we have a connection-org1.json file in which we have to include the network config path and required certificate path as per our network files path

**Note:-**  if we login with **user1** credentials we have to change organization details such as **ORG_NAME = "Org1", ORG_MSP = "Org1MSP", PEER_NAME = PEER1, CA_INSTANCE = "ca.org1.example.com", ORG_ADMIN = "org1admin", SECRET = "org1adminpw"** for user1 in integra-nock-sdk/config/config.go file.

After that run the following command to run the integra-admin-backend application from integra-admin-backend/integra-nock-sdk diretory:

`go run .`

## Client application
### Clone the integra-client-backend repository from the particular github account
Go to the integra-client-backend/integra-nock-sdk diretory:

`cd integra-client-backend/integra-nock-sdk`

After that run the following command to download required go packages locally:

`go mod vendor`

After that we have a connection-org1.json file in which we have to include the network config path and required certificate path as per our network files path

**Note:-**  if we login with **user1** credentials we have to change organization details such as **ORG_NAME = "Org1", PEER = "peer0.org1.example.com", PEER_NAME = PEER, CA_INSTANCE = "ca.org1.example.com", ORG_ADMIN = "org1admin", SECRET = "org1adminpw"** for user1 in integra-nock-sdk/config/config.go file.

After that run the following command to run the integra-client-backend application from integra-client-backend/integra-nock-sdk diretory:

`go run .`

## Frontend
### Clone the integra-nock-frontend repository from the particular github account

Go to the integra-nock-frontend diretory:

`cd integra-nock-frontend`

 After that run the following command to download required node packages:

`npm install`

After that run the following command to run the integra-nock-frontend application from integra-nock-frontend diretory:

`npm start`