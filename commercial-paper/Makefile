.: generate

generate:
	@protoc --version
	@echo "commercial paper schema proto generation"
	@protoc -I=./proto/ \
	-I=${GOPATH}/src \
	-I=${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
	--go_out=plugins=grpc:./proto/  \
	--govalidators_out=./proto/ \
	--grpc-gateway_out=logtostderr=true:./proto/ \
	--swagger_out=logtostderr=true:./proto/ \
	--doc_out=./proto/ --doc_opt=markdown,commercial-paper.md \
	--cc-gateway_out=logtostderr=true:./proto/ \
	./proto/*.proto
