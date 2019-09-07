package serviceboot

// // ServiceConfig 服务配置
// type ServiceConfig struct {
//
// }
//
// func (s *ServiceConfig) GetServiceEndpoint() string {
// 	return viper.GetString("serviceEndpoint")
// }
//
// // LoadConfig 加载ServiceConfiguration的配置
// func loadServiceConfig(ctx context.Context, errorService errors.Service, logger *zap.Logger) (*ServiceConfig, string, errors.Error) {
// 	config := &ServiceConfig{}
// 	serviceEndpoint, ok := os.LookupEnv("ENV_SERVICE_ENDPOINT")
// 	if !ok {
// 		serviceEndpoint = *serviceEndpointFlag
// 	}
// 	serviceEndpoint, err := bootutils.WarpServerAddr(serviceEndpoint, errorService)
// 	if err != nil {
// 		return nil, "", err
// 	}
// 	config.serviceEndpoint = serviceEndpoint
// 	if *singleService {
// 		config.SingleService = *singleService
// 	}
// 	if config.SingleService {
// 		logger.Info("启动单体模式")
// 	}
// 	return config, *configPath, nil
// }
