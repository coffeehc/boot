//go:build linux
// +build linux

package engine

// func daemon(serviceInfo configuration.ServiceInfo) error {
//   conn, err := dbus.ConnectSessionBus()
//   if err != nil {
//     fmt.Fprintln(os.Stderr, "Failed to connect to session bus:", err)
//     os.Exit(1)
//   }
//   defer conn.Close()
//   serviceInfo := configuration.GetServiceInfo()
//   names := conn.Names()
//   for _, name := range names {
//     if name == serviceInfo.ServiceName {
//
//     }
//   }
// }
