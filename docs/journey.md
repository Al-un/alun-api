- [Project structure](#project-structure)
  - [Folder structure](#folder-structure)
  - [Content organisation](#content-organisation)
  - [Naming convention](#naming-convention)
    - [Package entry point](#package-entry-point)
    - [API structure](#api-structure)
- [Framework or vanilla](#framework-or-vanilla)
  - [CORS](#cors)
- [Authentication middleware](#authentication-middleware)
- [`init()` call order](#init-call-order)
- [Logging](#logging)
- [JSON secret fields](#json-secret-fields)
- [Easy MongoDB update](#easy-mongodb-update)
- [08 Environment variables](#08-environment-variables)


## Project structure

When starting the project from an empty folder, without boilerplate, I did not know where to start: which folder should I create? while file name should I use? All ready-to-use stuff such as Rails or Vue CLI are opening the highway for me, I just had to follow the road! Now I am in the middle of the jungle, time to get my compass and find my way out.

> As a side-project, I do not extensively use framework such as Gin, until I really feel the need to, to explore the "manual way" as much as possible without, with the best I can, sacrificing code quality and best practices. Check [Framework or vanilla?](#framework-or-vanilla)

### Folder structure

Because, most of the time and I feel especially in Go, nothing beats the standard, let's start with the [Standard project layout](https://github.com/golang-standards/project-layout). At first, I confess it scared me as I felt being back in a huge Java project with all the `com.somewhere.somewhereelse...`. This is a side-project so I want a structure which would be:

- Scalable: Future-proofing structure, and hopefully my code as well, is deadly important to me
- Simple: I now want to avoid dig into deeply nested folder structure
- Professional: I want this project to be as "real-world" as possible

All my code is not in `pkg/` folder:

- `pkg/` is reserved for _library-like_ development which can potentially be re-used in other project or, why-not, by other users
- `alun/` is reserved for _al-un.fr_ back-end package
  - `internal/` package is not used so far as I do not expect my code to be re-used. Also, I might have another application along the al-un.fr back-end
  - > TODO: Move `al-un.fr` content to `internal/`?

Folder structure, as-of January 2020, looks like

```sh
/cmd
    /alun-api/          # Al-un.fr API executable
    {other executables}
/pkg
    /communication      # Communication libraries such as email or Slack integration
    /logger             # Logging stuff
/{some al-un.fr name}   # Code for al-un.fr back-end
    /core               # al-un.fr core code
    /{mini-app-1} 
    /{mini-app-2} 
    /{...} 
    /utils              # Always helpful to have some "utils/" 
.env                    # dotenv files are at the root of the project
```

### Content organisation

For the sake of exploration, I am building multiple mini-applications relying on a core package. Such mini-application would be an independant package / folder. This most likely calls for some code redundancy but such isolation will allow some experiments in a given mini-application without breaking the other mini-application.

### Naming convention

I have not digged for a strict naming convention so I will start with

#### Package entry point

If a file has the same name as the package, such as `pkg/core/core.go`, then it is the package entry point - Package documentation must be written in this file - "Various" initialisation such as package logger has to be done in this file

#### API structure

API structure main revolves around three files:

- Entity layer: `models.go` or `{feature}_models.go`:
  - Define all data models
  - Define business logic when it can be defined for a single instance. E.g. _Mark an order as completed for a given order_ would be a function on the `Order` struct
- Database persistence layer: `dao.go` or `{feature}_dao.go`:
  - Methods should **not** return a database specific type but a structure defined in some `(xxx_)models.go` or some standard Go structure
  - Methods should, whenever possible, not include business logic except to guarantee database consistency
- Service layer: `handlers.go` or `{feature}_handlers.go`:
  - Define all endpoint handlers
  - Define business logic when it is request specific. E.g. _Update user language based on HTTP request_ would be a function checking something in the incoming request and update the request body, namely the user, accordingly.
- Routing layer: `api.go` or `{feature}_api.go`:
  - Define routes and map route endpoints to appropriate handlers
  - Can also define route guards

## Framework or vanilla

After reading some articles such as [Why I don't use go web frameworks](https://medium.com/code-zen/why-i-don-t-use-go-web-frameworks-1087e1facfa4), comparing Revel, Gin or Gorilla, I ended up with Gorilla. Why? Because most of the tutorials are based on Gorilla :) 

Also, I try to be as-vanilla as possible, mainly for the sake of learning. As-of January 2020, I only use Gorilla for routing but I might ask for more Gorilla help later, hello Websockets.

### CORS

To test my very first requests, I used `curl`. That's nice but moving from the stone age to the bronze age, aka some front-end development, makes me meet a friend of many of us: CORS.

Without further ado. Let's start with the approach: Having some middleware was the most common approach I found with the Adapter pattern ([Writing middleware in #golang and how Go makes it so much fun.](https://medium.com/@matryer/writing-middleware-in-golang-and-how-go-makes-it-so-much-fun-4375c1246e81)). But each request would have the same CORS config which I wanted to avoid

It might be called, or derived from, [principle of least privilege](https://en.wikipedia.org/wiki/Principle_of_least_privilege) but if an endpoint only needs `GET, POST`, I want to return `GET,POST`, not `GET,POST,PUT,DELETE` nor `*`.

I would then need a custom adapter hence the signature

```go
type CorsConfig struct {
	Hosts   string
	Methods string
	Headers string
}

func AddCorsHeaders(next http.Handler, corsConfig CorsConfig) http.Handler {}
```
This is nice for individual endpoints configuration but some endpoints might share the same URL pattern with different methods. Let's consider three endpoints:

- `GET`: `/aaa`
- `POST`: `/aaa`
- `GET`: `/bbb`

When doing a preflight request on `/aaa`, the server actually has to return `GET,POST` but only `GET` when doing the preflight on `/bbb`.

Some references:
- [Authoritative guide to CORS (Cross-Origin Resource Sharing) for REST APIs](https://www.moesif.com/blog/technical/cors/Authoritative-Guide-to-CORS-Cross-Origin-Resource-Sharing-for-REST-APIs/)
- [StackOverflow: What are proper status codes for CORS preflight requests?](https://stackoverflow.com/questions/46026409/what-are-proper-status-codes-for-cors-preflight-requests)

## Authentication middleware

Abstract of routing

## `init()` call order

## Logging

Based on Logback?

https://github.com/op/go-logging

## JSON secret fields

Password for login and registration but not for user display or edit

https://stackoverflow.com/q/47256201/4906586
https://stackoverflow.com/q/47256201/4906586
https://blog.gopheracademy.com/advent-2016/advanced-encoding-decoding/

## Easy MongoDB update

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
