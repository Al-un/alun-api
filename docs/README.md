# Al-un API

- [Resources](#Resources)

API for al-un.fr

## Resources

- [Project layout](https://github.com/golang-standards/project-layout)

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