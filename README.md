# Hyperledger Fabric Golang application examples


* [Commercial paper](commercial-paper)


## How to create HyperLedger Fabric Golang application project

### Main points

* `Protobuf` message and service definitions allows to model chaincode in high level Interface Definition Language (IDL)
* [Code generation](https://blog.golang.org/generate) allows to automate development process 
   of creating API's, SDK, documentation for chainode


### Development steps

1. Define model using `.proto` file
2. Generate code and documentation with generators
3. Implement chaincode as service and tests 
4. Create chaincode binary
5. Create API

### Generators

Generators allows to automatically build lot of useful code and docs : Golang structures, 
validators, gRPC service and client, documentation in Markdown format and Swagger specification,
chaincode gateway for implementing API or SDK and mapper for embedding strong typed gRPC service
to chaincode implementation.

1. [Protobuf generator](https://github.com/golang/protobuf)

Go support for Protocol Buffers

2. [Validator generator](https://github.com/mwitkow/go-proto-validators)

A `protoc` plugin that generates `Validate() error` functions on Go proto structs based on field options inside
 `.proto` files.

3. [gRPC gateway generator](https://github.com/grpc-ecosystem/grpc-gateway)

Provides HTTP+JSON interface to gRPC service. A small amount of configuration in your service to attach HTTP semantics 
is all that's needed to generate a reverse-proxy with this library. Optionally emitting API definitions for
OpenAPI (Swagger) v2.

4. [Documentation generator](https://github.com/pseudomuto/protoc-gen-doc)

documentation generator plugin for the Google Protocol Buffers compiler (protoc). 
The plugin can generate HTML, JSON, DocBook and Markdown documentation from comments in your .proto files.

5. [Chaincode gateway generator]()

#### Install generators

```shell
# go get -u github.com/golang/protobuf/protoc-gen-go
# go get github.com/mwitkow/go-proto-validators/protoc-gen-govalidators
# go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
# go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
# go get -u github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc
# GO111MODULE=on go install github.com/s7techlab/cckit/gateway/protoc-gen-cc-gateway
```

This will place five binaries in your `$GOBIN`;

* `protoc-gen-go`
* `protoc-gen-govalidators`
* `protoc-gen-grpc-gateway`
* `protoc-gen-swagger`
* `protoc-gen-doc`

Make sure that your `$GOBIN` is in your `$PATH`.

### Dependencies

Golang 1.11+ supports [modules](https://blog.golang.org/using-go-modules), this is recommended
way to work with dependencies. Simple golang module usage example can be founded 
[here](https://github.com/golang/go/wiki/Modules#quick-start). Steps to create Golang project with modules:


## Tutorial based on Commercial Paper example

#### 1. Create a directory for project 

outside of your `$GOPATH`

#### 2. Initialize a new module inside this directory
 
```bash
# go mod init {put your module name here}  
```

(for example `github.com/s7techlab/cckit-sample`)

#### 3. Create `.proto` definitions 
 
gRPC technology stack natively supports a clean and powerful way to specify service contracts using the Interface
Definition Language (IDL):
* messages defines data structures of the input parameters and return types.
* services definition outlines methods signatures that can be invoked remotely

Chaincode [messages and service](commercial-paper/proto/commercial-paper.proto) allows to define chaincode interface and
data schema.



#### 4. Generate code 

Create [Makefile](commercial-paper/Makefile) for compiling `.proto` to `Golang` code

```Makefile
.: generate

generate:
	@protoc --version
	@echo "commercial paper schema proto generation"
	@protoc -I=./proto/ \
	-I=${GOPATH}/src \
	-I=${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
	--go_out=grpc:./proto/  \
	--govalidators_out=./proto/ \
	--grpc-gateway_out=logtostderr=true:./proto/ \
	--swagger_out=logtostderr=true:./proto/ \
	--doc_out=./proto/ --doc_opt=markdown,commercial-paper.md \
	--cc-gateway_out=logtostderr=true:./proto/ \
	./proto/*.proto
```

and run it. Following files will be generated:

* [commercial-paper.md](commercial-paper/proto/commercial-paper.md) -documentation
* [commercial-paper.pb.cc.go](commercial-paper/proto/commercial-paper.pb.cc.go) - gateway for chaincode
* [commercial-paper.pb.go](commercial-paper/proto/commercial-paper.pb.go) - Golang structs and gRPC service
* [commercial-paper.pb.gw.go](commercial-paper/proto/commercial-paper.pb.gw.go) - gRPC gateway
* [commercial-paper.swagger.json](commercial-paper/proto/commercial-paper.swagger.json) - Swagger specification
* [commercial-paper.validator.pb.go](commercial-paper/proto/commercial-paper.validator.pb.go) - validators

#### 5. Load dependencies

Standard commands like `go build` or `go test` will automatically add new dependencies as needed to 
satisfy imports (updating go.mod and downloading the new dependencies).

`# go mod vendor`

This command add dependencies to [go.mod](go.mod) file and download it to [vendor](vendor) directory
Go.mod file will contain:

```
module github.com/s7techlab/cckit-sample

go 1.12

require (
	github.com/golang/protobuf v1.3.2
	github.com/grpc-ecosystem/grpc-gateway v1.9.5
	github.com/mwitkow/go-proto-validators v0.0.0-20190709101305-c00cd28f239a
	github.com/pkg/errors v0.8.1
	github.com/s7techlab/cckit v0.6.1
	google.golang.org/genproto v0.0.0-20190801165951-fa694d86fc64
	google.golang.org/grpc v1.22.1
)
```

#### 6. Implement chaincode as service and tests

Implement [chaincode service and chaincode](commercial-paper/chaincode/chaincode.go) and 
[test](commercial-paper/chaincode/chaincode_test.go). After that don't forget to call `go mod vendor`
to download newly added dependencies  (Ginkgo and Gomega)

You can test chaincode service with command

`#chaincode go test -mod vendor`








