.PHONY: proto

proto:
	protoc -I ./internal/proto --go_out=./internal pb.proto