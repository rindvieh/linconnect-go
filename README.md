# LinConnect-go

This is a quick and dirty golang port of https://github.com/hauckwill/linconnect-server

I made this to use LinConnect on MacOS X, however the code should work on Linux/Windows(growl) (not tested)

## Installation
   go get github.com/rindvieh/linconnect-go

## Configuration/First start

   ./linconnect-go -init

Will create a default config.json in the current folder and start the service.

   ./linconnect-go -init -conf myConfig.json

Will create a default config at -conf path and start the service using the new config file.

## Usage

  ./linconnect-go
  ./linconnect-go -h
  ./LinConnect-go -conf myConfig.json
