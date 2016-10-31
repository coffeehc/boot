#!/usr/bin/env bash

go build -v github.com/coffeehc/microserviceboot/serviceboot/restboot
go build -v github.com/coffeehc/microserviceboot/serviceboot/grpcboot
go build -v github.com/coffeehc/microserviceboot/serviceclient/restclient
go build -v github.com/coffeehc/microserviceboot/serviceclient/grpcclient
go build -v github.com/coffeehc/microserviceboot/consultool
go build -v github.com/coffeehc/microserviceboot/integrated/databaseservice
go build -v github.com/coffeehc/microserviceboot/integrated/redisservice
