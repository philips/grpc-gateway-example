## gRPC + REST Gateway Play

Blog post: https://coreos.com/blog/gRPC-protobufs-swagger.html

To try it all out do this:

```
$ go get -u github.com/philips/grpc-gateway-example
$ grpc-gateway-example serve
$ grpc-gateway-example echo "my first rpc echo"
$ curl -X POST -k https://localhost:10000/v1/echo -H "Content-Type: text/plain" -d '{"value": "foo"}'
{"value":"my REST echo"}
```


Huge thanks to the hard work people have put into the [Go gRPC bindings][gogrpc] and [gRPC to JSON Gateway][grpcgateway]

[gogrpc]: https://github.com/grpc/grpc-go
[grpcgateway]: https://github.com/grpc-ecosystem/grpc-gateway
