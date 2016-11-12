#!/usr/bin/env bash

go build github.com/coffeehc/microserviceboot/serviceboot/restboot
go build github.com/coffeehc/microserviceboot/serviceboot/grpcboot
go build github.com/coffeehc/microserviceboot/serviceclient/restclient
go build github.com/coffeehc/microserviceboot/serviceclient/grpcclient
go build github.com/coffeehc/microserviceboot/consultool
go build github.com/coffeehc/microserviceboot/integrated/databaseservice
go build github.com/coffeehc/microserviceboot/integrated/redisservice
