# Hyperledger Fabric Golang application examples


* [Commercial paper](samples/cpaper)


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

## Prerequisites

### Generators

Generators allows to automatically build lot of useful code and docs : Golang structures, 
validators, gRPC service and client, documentation in Markdown format and Swagger specification,
chaincode gateway for implementing API or SDK and mapper for embedding strong typed gRPC service
to chaincode implementation.

![img](samples/cpaper/docs/img/cc-code-gen.png)

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

5. [Chaincode gateway generator](https://github.com/s7techlab/cckit/tree/master/gateway)

#### Install generators

`cd geterators && ./install.sh`

This will place five binaries in [generators/bin](generators/bin);

* `protoc-gen-go`
* `protoc-gen-govalidators`
* `protoc-gen-grpc-gateway`
* `protoc-gen-swagger`
* `protoc-gen-doc`


## Step by step

#### 1. Create a directory for project 

outside of your `$GOPATH`

#### 2. Initialize a new module inside this directory
 
```bash
# go mod init {put your module name here}  
```

(for example `github.com/s7techlab/hyperledger-fabric-samples)

#### 3. Create `.proto` definitions 
 
gRPC technology stack natively supports a clean and powerful way to specify service contracts using the Interface
Definition Language (IDL):
* messages defines data structures of the input parameters and return types.
* services definition outlines methods signatures that can be invoked remotely

Chaincode [messages and service](samples/cpaper/cpaper.proto) allows to define chaincode interface and
data schema.


#### 4. Generate code 

Create [proto/Makefile](samples/cpaper/proto/Makefile) for compiling `.proto` to `Golang` code

```Makefile
.: generate

generate:
	@protoc --version
	@echo "commercial paper schema proto generation"
	@protoc -I=./ \
	-I=${GOPATH}/src \
	-I=${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
	--go_out=plugins=grpc:./ \
	--govalidators_out=./ \
	--grpc-gateway_out=logtostderr=true:./ \
	--swagger_out=logtostderr=true:./ \
	--doc_out=./ --doc_opt=markdown,commercial-paper.md \
	--cc-gateway_out=logtostderr=true:./ \
	./*.proto
```

and run it. Following files will be generated:

* [commercial-paper.md](samples/cpaper/proto/commercial-paper.md) -documentation
* [commercial-paper.pb.cc.go](samples/cpaper/proto/commercial-paper.pb.cc.go) - gateway for chaincode
* [commercial-paper.pb.go](samples/cpaper/proto/commercial-paper.pb.go) - Golang structs and gRPC service
* [commercial-paper.pb.gw.go](samples/cpaper/proto/commercial-paper.pb.gw.go) - gRPC gateway
* [commercial-paper.swagger.json](samples/cpaper/proto/commercial-paper.swagger.json) - Swagger specification
* [commercial-paper.validator.pb.go](samples/cpaper/proto/commercial-paper.validator.pb.go) - validators

#### 5. Load dependencies

Standard commands like `go build` or `go test` will automatically add new dependencies as needed to 
satisfy imports (updating go.mod and downloading the new dependencies).

`# go mod vendor`

This command add dependencies to [go.mod](go.mod) file and download it to [vendor](vendor) directory
Go.mod file will contain:

```
module github.com/s7techlab/hyperledger-fabric-samples

go 1.12

require (
	github.com/golang/protobuf v1.3.2
	github.com/grpc-ecosystem/grpc-gateway v1.9.5
	github.com/hyperledger/fabric v1.4.4
	github.com/mwitkow/go-proto-validators v0.0.0-20190709101305-c00cd28f239a
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/pkg/errors v0.8.1
	github.com/s7techlab/cckit v0.6.9
	google.golang.org/genproto v0.0.0-20190801165951-fa694d86fc64
	google.golang.org/grpc v1.22.1
)

```

#### 6. Implement chaincode as service and tests

Implement [chaincode service and chaincode](samples/cpaper/chaincode/chaincode.go) and 
[test](samples/cpaper/chaincode/chaincode_test.go). After that don't forget to call `go mod vendor`
to download newly added dependencies  (Ginkgo and Gomega)

You can test chaincode service with command

`#chaincode go test -mod vendor`

Check test code coverage: chaincode logic must be covered by test to maximum level. Code coverage of 70-80% is a 
reasonable goal for system test of most projects with most coverage metrics

`#chaincode go test -mod vendor -coverage`


#### 7. Implement off-chain application to communicate with chaincode (API)

With [CCKit gateway](https://github.com/s7techlab/cckit/tree/master/gateway) and generated 
[gRPC service server](samples/cpaper/proto/commercial-paper.pb.go) and [gRPC gateway](samples/cpaper/proto/commercial-paper.pb.gw.go)
quite ease to implement [API](samples/cpaper/api) for chaincode.

You can run provided mocked example using command
```
# cd commercial-paper/api/mock
# go run main.go
```

![start](samples/cpaper/docs/img/gateway-mocked-start.png)

Then you can use API usage examples and sample payloads:

![example](samples/cpaper/docs/img/gateway-mocked-usage.png)

`grpc-gateway` will automatically converts http request to gRPC call, input JSON payloads to protobuf, invokes chaincode 
service and then converts returned value from protobuf to JSON. You can also use this service as pure gRPC service. 
Chaincode methods can be called with [generated gRPC client](samples/cpaper/proto/commercial-paper.pb.go). 

Swagger [specification](samples/cpaper/proto/commercial-paper.swagger.json), service and schema documentation are also 
[auto-generated](samples/cpaper/proto/commercial-paper.md).





