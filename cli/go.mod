module github.com/way-platform/mapon-go/cli

go 1.26.0

require (
	github.com/spf13/cobra v1.10.2
	github.com/way-platform/mapon-go v0.0.0
	golang.org/x/term v0.41.0
	google.golang.org/protobuf v1.36.11
)

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.11-20260209202127-80ab13bee0bf.1 // indirect
	connectrpc.com/connect v1.19.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
	golang.org/x/sys v0.42.0 // indirect
)

replace github.com/way-platform/mapon-go => ../
