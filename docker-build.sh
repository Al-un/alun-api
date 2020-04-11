#!/bin/sh

docker build --file api-memo.Dockerfile --tag alunsng/alun:api-memo .
docker build --file api-user.Dockerfile --tag alunsng/alun:api-user .
docker build --file api-monolith.Dockerfile --tag alunsng/alun:api-monolith .

docker push alunsng/alun:api-memo
docker push alunsng/alun:api-user
docker push alunsng/alun:api-monolith