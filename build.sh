#!/usr/bin/env bash

go build -v github.com/coffeehc/microserviceboot/serviceboot
go build -v github.com/coffeehc/microserviceboot/serviceclient
go build -v github.com/coffeehc/microserviceboot/base
go build -v github.com/coffeehc/microserviceboot/consultool
go build -v github.com/coffeehc/microserviceboot/integrated/databaseservice
go build -v github.com/coffeehc/microserviceboot/integrated/redisservice