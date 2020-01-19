## 01 Project structure

Project structure involves:

- Folder structure
- Content organisation
- Naming convention

### Folder structure

Packages definition

### Content Organisation

- Flat structure.
- Models have a dedicated files

### Naming convention

- If a file has the same name as the folder, it is considered as the entry point of
  the package, similarly to a `index.js` in JavaScript

## 02 Framework or vanilla

Gorilla because a lot of tutos with it

### CORS

https://www.moesif.com/blog/technical/cors/Authoritative-Guide-to-CORS-Cross-Origin-Resource-Sharing-for-REST-APIs/
https://stackoverflow.com/questions/46026409/what-are-proper-status-codes-for-cors-preflight-requests

## 03 Authentication middleware

Abstract of routing

## 04 `init()` call order

## 05 Logging

Based on Logback?

https://github.com/op/go-logging

## 06 JSON secret fields

Password for login and registration but not for user display or edit

https://stackoverflow.com/q/47256201/4906586
https://stackoverflow.com/q/47256201/4906586
https://blog.gopheracademy.com/advent-2016/advanced-encoding-decoding/

## 07 Easy MongoDB update

https://stackoverflow.com/q/55564562/4906586
https://stackoverflow.com/q/9611833/4906586
https://stackoverflow.com/a/58359989/4906586

```sql
> db.al_memos.deleteMany({"trackedentity.createdAt": {"$lt": new ISODate("2019-12-27T10:35:33")} })
{ "acknowledged" : true, "deletedCount" : 0 }
> db.al_memos.deleteMany({"trackedentity.createdOn": {"$lt": new ISODate("2019-12-27T10:35:33")} })
{ "acknowledged" : true, "deletedCount" : 1 }
> db.al_memos.deleteMany({"trackedentity.createdOn": null })
{ "acknowledged" : true, "deletedCount" : 4 }
```

## 08 Environment variables

https://github.com/joho/godotenv

```sh
go get github.com/joho/godotenv
```

Need to load per package