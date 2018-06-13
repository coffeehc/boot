package boot

import "context"

const (
	//	ctx_Key_serverAddr  = "boot.serverAddr"
	//	ctx_Key_serviceInfo = "boot.serviceInfo"
	ctx_Key_serviceName = "boot.serviceName"

//	ctx_Key_grpcServer  = "boot.grpcServer"
//	ctx_Key_etcdClient  = "boot.etcdClient"
)

//
//func SetGRPCServer(ctx context.Context, server *grpc.Server) context.Context {
//	return context.WithValue(ctx, ctx_Key_grpcServer, server)
//}
//
//func GetGRPCServer(ctx context.Context) *grpc.Server {
//	return ctx.Value(ctx_Key_grpcServer).(*grpc.Server)
//}
//
//func SetEtcdClient(ctx context.Context, client *clientv3.Client) context.Context {
//	return context.WithValue(ctx, ctx_Key_etcdClient, client)
//}
//
//func GetEtcdClient(ctx context.Context) *clientv3.Client {
//	return ctx.Value(ctx_Key_etcdClient).(*clientv3.Client)
//}
//
//func SetServerAddr(ctx context.Context, serverAddr string) context.Context {
//	return context.WithValue(ctx, ctx_Key_serverAddr, serverAddr)
//}
//
//func GetServerAddr(ctx context.Context) string {
//	return ctx.Value(ctx_Key_serverAddr).(string)
//}
//
func SetServiceName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, ctx_Key_serviceName, name)
}

func GetServiceName(ctx context.Context) string {
	return ctx.Value(ctx_Key_serviceName).(string)
}

//func SetServiceInfo(ctx context.Context, info ServiceInfo) context.Context {
//	return context.WithValue(ctx, ctx_Key_serviceInfo, &info)
//}
//
//func GetServiceInfo(ctx context.Context) ServiceInfo {
//	return *ctx.Value(ctx_Key_serviceInfo).(*ServiceInfo)
//}
