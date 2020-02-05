#!/bin/bash
if [ $GOPATH == "" ]
then
    GOPATH=$HOME/go
fi
PATH=$GOPATH/bin/:$PATH
go get -u github.com/gobuffalo/packr/v2/...
go get -u github.com/gobuffalo/packr/v2/packr2

cd panel-frontend 
npm run-script build
cd ..

cd player
packr2
cd ..

cd panel
packr2
cd ..

go build -o ABCVideo-player ./player
go build -o ABCVideo-panel ./panel