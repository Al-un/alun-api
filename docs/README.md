# Al-un API <!-- omit in toc -->

- [Docker](#docker)
  - [Monolithic build](#monolithic-build)
  - [Microservices build](#microservices-build)
- [Resources](#resources)
  - [MongoDB](#mongodb)
    - [Queries](#queries)
- [Heroku](#heroku)

API for al-un.fr

## Docker

Images are available on [Docker hub](https://hub.docker.com/repository/docker/alunsng/alun).

### Monolithic build

```sh
sudo docker-compose --file api-monolith-compose.yml build --no-cache
sudo docker-compose --file api-monolith-compose.yml up
```

Endpoints are all in `https://localhost:8000`

### Microservices build

```sh
sudo docker-compose --file api-microservice-compose.yml build --no-cache
sudo docker-compose --file api-microservice-compose.yml up
```

Endpoints are:

- _user app_: `http://localhost:8001`
- _memo app_: `http://localhost:8002`

## Resources

- [Project layout](https://github.com/golang-standards/project-layout)
  - `pkg/`: Code like libraries
  - `alun/`: al-un.fr back-end

### MongoDB

- [MongoDB driver](https://github.com/mongodb/mongo-go-driver)
- [Quick Start: Golang and MongoDB](https://www.mongodb.com/blog/post/quick-start-golang--mongodb--starting-and-setup)
- [How to find a MongoDB document by its BSON ObjectID](https://kb.objectrocket.com/mongo-db/how-to-find-a-mongodb-document-by-its-bson-objectid-using-golang-452)

#### Queries

```sql
db.al_users_login.deleteMany({ "timestamp": {"$lte": new Date() } )
db.al_users_login.deleteMany( { "timestamp": { "$lte": new Date(2999,1,1) } } )
```

https://stackoverflow.com/a/30772989/4906586
https://stackoverflow.com/a/47170066/4906586

## Heroku

```sh
heroku accounts:set al-un

heroku apps:create --region=eu --buildpack=heroku/go alun-api
heroku addons:create mongolab:sandbox --app alun-api
```
