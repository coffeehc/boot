package manage

// type Service interface {
//   GetHttpService() httpx.Service
//   GetEndpoint() string
//   Start(onShutdown func())
// }
//
// type serviceImpl struct {
//   httpService httpx.Service
//   endpoint    string
// }
//
// func (impl *serviceImpl) GetHttpService() httpx.Service {
//   return impl.httpService
// }
//
// func (impl *serviceImpl) GetEndpoint() string {
//   return impl.endpoint
// }
//
// func (impl *serviceImpl) Start(onShutdown func()) {
//   impl.httpService.Start(onShutdown)
// }
//
// func (impl *serviceImpl) registerManager() {
//   router := impl.httpService.GetGinEngine()
//   // router.Use(gin.BasicAuth(gin.Accounts{
//   // 	"root": "abc###123",
//   // }))
//   impl.registerServiceRuntimeInfoEndpoint(router)
//   impl.registerHealthEndpoint(router)
//   impl.registerMetricsEndpoint(router)
// }
//
// func newManageService() (Service, errors.Error) {
//   service := &serviceImpl{
//   }
//   if !viper.IsSet("manage.serverAddr"){
//     viper.SetDefault("manage.serverAddr", "0.0.0.0:0")
//   }
//   manageEndpoint := viper.GetString("manage.serverAddr")
//   manageEndpoint, err := utils.WarpServiceAddr(manageEndpoint)
//   if err != nil {
//     return nil, err
//   }
//   lis, err1 := net.Listen("tcp4", manageEndpoint)
//   if err1 != nil {
//     log.Error("启动ManageServer失败", zap.Error(err1))
//     return nil, errors.SystemError("启动ManageServer失败")
//   }
//   manageEndpoint = lis.Addr().String()
//   lis.Close()
//   service.endpoint = manageEndpoint
//   log.Debug("设置管理Endpoint", zap.String("endpoint", manageEndpoint))
//   service.httpService = httpx.NewService("manage", &httpx.Config{
//     ServerAddr: manageEndpoint,
//   }, log.GetLogger())
//   service.registerManager()
//   return service, nil
// }
