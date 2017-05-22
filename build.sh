#!/usr/bin/env bash

OPTIONIS="-gcflags=-trimpath=$GOPATH -asmflags=-trimpath=$GOPATH -ldflags \"-w -s\""

go build  $OPTIONS github.com/coffeehc/microserviceboot/serviceboot/restboot
go build $OPTIONS github.com/coffeehc/microserviceboot/serviceboot/grpcboot
go build $OPTIONS github.com/coffeehc/microserviceboot/serviceclient/restclient
go build $OPTIONS github.com/coffeehc/microserviceboot/serviceclient/grpcclient
go build $OPTIONS github.com/coffeehc/microserviceboot/consultool
