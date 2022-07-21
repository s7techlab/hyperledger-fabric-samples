PROTO_PACKAGES_CC := samples

proto: clean
	@for pkg in $(PROTO_PACKAGES_CC) ;do echo $$pkg && buf generate --template buf.gen.cc.yaml $$pkg -o ./$$(echo $$pkg | cut -d "/" -f1); done
clean:
	@for pkg in $(PROTO_PACKAGES_CC); do find $$pkg \( -name '*.pb.go' -or -name '*.pb.cc.go' -or -name '*.pb.gw.go' -or -name '*.swagger.json' -or -name '*.pb.md' \) -delete;done