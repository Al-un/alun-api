[![CircleCI](https://circleci.com/gh/Al-un/alun-api/tree/master.svg?style=svg)](https://circleci.com/gh/Al-un/alun-api/tree/master)
[![Go Report Card](https://goreportcard.com/badge/github.com/Al-un/alun-api)](https://goreportcard.com/report/github.com/Al-un/alun-api)
[![Docker Badge](https://img.shields.io/docker/cloud/build/alunsng/alun-api.svg)](https://hub.docker.com/r/alunsng/alun-api)

# Al-un API <!-- omit in toc -->

- [Docker](#docker)
  - [Monolithic build](#monolithic-build)
  - [Microservices build](#microservices-build)
- [Resources](#resources)
  - [MongoDB](#mongodb)
    - [Queries](#queries)
- [Heroku](#heroku)
- [Notes](#notes)

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

## Notes

**Testing**:

- https://blog.alexellis.io/golang-writing-unit-tests/
- https://stackoverflow.com/questions/47045445/idiomatic-way-to-pass-variables-to-test-cases-in-golang
- https://blog.questionable.services/article/testing-http-handlers-go/
- https://medium.com/@matryer/5-simple-tips-and-tricks-for-writing-unit-tests-in-golang-619653f90742
- https://stackoverflow.com/questions/23729790/how-can-i-do-test-setup-using-the-testing-package-in-go
- http://cs-guy.com/blog/2015/01/test-main/
- https://www.toptal.com/go/your-introductory-course-to-testing-with-go
- https://lanre.wtf/blog/2017/04/08/testing-http-handlers-go/
- https://stackoverflow.com/questions/44325232/are-tests-executed-in-parallel-in-go-or-one-by-one/44326377#44326377
- https://blog.golang.org/subtests
- Testing tips: https://medium.com/@povilasve/go-advanced-tips-tricks-a872503ac859
- Helpers: https://github.com/benbjohnson/testing

**CircleCI**

- Golang modules: https://circleci.com/blog/go-v1.11-modules-and-circleci/
- Language guide: https://circleci.com/docs/2.0/language-go/
- https://itnext.io/go-modules-and-circleci-c0d6fac0b000
- dependencies location: https://stackoverflow.com/questions/52082783/how-do-i-find-the-go-module-source-cache
- artifacts: https://circleci.com/docs/2.0/artifacts/

**MongoDB: lookup**

- https://www.mongodb.com/blog/post/quick-start-golang--mongodb--data-aggregation-pipeline
- https://www.sitepoint.com/using-joins-in-mongodb-nosql-databases/
- https://www.mongodb.com/blog/post/6-rules-of-thumb-for-mongodb-schema-design-part-3