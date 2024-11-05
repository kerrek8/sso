ADD env variables to the .env file or in your environment
look at the .env.example file for the variables needed

DB_CONTAINER is the name of the container that will be created for the database

before start container you need to create a network and add the container with db and container with the app to the network


then you need to apply the migrations to the database
```bash
go run ./cmd/migrator/main.go --storage-host= --storage-port= --storage-user= --storage-password= --migrations-path=./migrations
```

to build the app
```bash
go build -o ./bin/sso ./cmd/sso/main.go
```

to run the app
```bash
go run ./cmd/sso/main.go
```
to build the docker image
```bash
docker build -t sso_service .
```
to run the docker container
```bash
docker run --name=sso -p 44044:44044 -network sso_service_network sso_service
```
