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
