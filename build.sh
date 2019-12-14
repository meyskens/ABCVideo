#!/bin/bash
if [ $GOPATH == ""] 
then
    GOPATH=$HOME/go
fi
PATH=$GOPATH/bin/:$PATH
go get -u github.com/gobuffalo/packr/v2/...
go get -u github.com/gobuffalo/packr/v2/packr2

cd panel-frontend 
npm run-script build
cd ..

packr2
go build ./