
### 协议支持

|消息类型|编码协议|传输协议|
|--|--|--|
|PingPong|Thrift/Protobuf|TTHeader/HTTP2(gRPC)|
|Oneway|Thrift|TTHeader|
|Streaming|Protobuf|HTTP2(gRPC)|
- `PingPong`: 客户端发起一个请求后会等待一个响应才可以进行下一次请求
- `Oneway`: 客户端发起一个请求后不会等待响应，也不会阻塞后续请求
- `Streaming`: 客户端发起一个或多个请求后，等待一个或多个请求

TODO： 
- 目前代码生成只支持Protobuf,要考虑Streaming的支持
- 当前代码生成能否和kitex的代码生成进行合并，避免功能缺失？
