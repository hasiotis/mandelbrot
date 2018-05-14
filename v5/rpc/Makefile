OUTPUT=rpc.pb.go

${OUTPUT}: rpc.proto
	# go get -u github.com/golang/protobuf/protoc-gen-go
	protoc -I . rpc.proto --go_out=plugins=grpc:.
