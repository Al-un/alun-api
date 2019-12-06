# Al-un API

- [Resources](#Resources)

API for al-un.fr

## Resources

- [Project layout](https://github.com/golang-standards/project-layout)

### MongoDB

- [MongoDB driver](https://github.com/mongodb/mongo-go-driver)
- [Quick Start: Golang and MongoDB](https://www.mongodb.com/blog/post/quick-start-golang--mongodb--starting-and-setup)
- [How to find a MongoDB document by its BSON ObjectID](https://kb.objectrocket.com/mongo-db/how-to-find-a-mongodb-document-by-its-bson-objectid-using-golang-452)


## Heroku

```sh
heroku accounts:set al-un

heroku apps:create --region=eu --buildpack=heroku/go alun-api
heroku addons:create mongolab:sandbox --app alun-api
```