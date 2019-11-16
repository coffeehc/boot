1. 前言
grpc-go本身已经支持安全通信，该文是举例介绍下双向认证的安全通信，客户端和服务端是如何实现的。

2. 使用openssl生成密钥和证书
  简单介绍下双向认证的原理，客户端和服务端在进行双向认证前会交换彼此的证书，如何信任对方的证书呢？这就需要有个权威的第三方CA（认证中心）为双方“背书”，由CA为双方签发证书，这样客户端和服务端可以用CA的根证书来验证对方的证书取得信任。
  下面简单介绍下用openssl生成CA根证书以及客户端和服务端的密钥及证书。下面涉及的文件类型有三种：.key一般作为私钥；.pem一般作为证书；.csr是Cerificate Signing Request的缩写，用来请求签发证书的。
  我们生成密钥的文件结构如下：
  
  ```
    key
    ├── ca.key(根密钥)
    ├── ca.pem(根证书)
    ├── client
    │   ├── client.csr(客户端证书签发请求)
    │   ├── client.key(客户端密钥)
    │   └── client.pem(客户端证书)
    └── server
        ├── server.csr(服务端证书签发请求)
        ├── server.key(服务端密钥)
        └── server.pem(服务端证书)
  ```
  openssl的安装不再赘述。
  
  生成CA的密钥和根证书
  生成2048位密钥：
  ```bash
  [root@VM_4_242_centos /usr1/key]# openssl genrsa -out ca.key 2048
  Generating RSA private key, 2048 bit long modulus
  ....+++
  ......................................................+++
  e is 65537 (0x10001)
  ```
  
  生成根证书，这里根证书没有设置日期，所以默认是永不过期的：
  
  ```bash
  [root@VM_4_242_centos /usr1/key]# openssl req -new -x509 -key ca.key -out ca.pem       
  You are about to be asked to enter information that will be incorporated
  into your certificate request.
  What you are about to enter is what is called a Distinguished Name or a DN.
  There are quite a few fields but you can leave some blank
  For some fields there will be a default value,
  If you enter '.', the field will be left blank.
  -----
  Country Name (2 letter code) [XX]:
  State or Province Name (full name) []:
  Locality Name (eg, city) [Default City]:
  Organization Name (eg, company) [Default Company Ltd]:
  Organizational Unit Name (eg, section) []:
  Common Name (eg, your name or your server's hostname) []:demo
  Email Address []:
  ```
  
  生成服务端的密钥和证书
  生成2048位密钥：
  
  ```bash
  [root@VM_4_242_centos /usr1/key]# mkdir server
  [root@VM_4_242_centos /usr1/key]# openssl genrsa -out server/server.key 2048
  Generating RSA private key, 2048 bit long modulus
  ....................................................+++
  ..........................+++
  e is 65537 (0x10001)
  ```
  
  生成证书签发请求，注意这里面会要求填一系列内容，除了Common Name外都可以不填，Common Name对grpc的双向认证很重要：
  
  ```bash
  [root@VM_4_242_centos /usr1/key]# openssl req -new -key server/server.key -out server/server.csr
  You are about to be asked to enter information that will be incorporated
  into your certificate request.
  What you are about to enter is what is called a Distinguished Name or a DN.
  There are quite a few fields but you can leave some blank
  For some fields there will be a default value,
  If you enter '.', the field will be left blank.
  -----
  Country Name (2 letter code) [XX]:
  State or Province Name (full name) []:
  Locality Name (eg, city) [Default City]:
  Organization Name (eg, company) [Default Company Ltd]:
  Organizational Unit Name (eg, section) []:
  Common Name (eg, your name or your server's hostname) []:demo
  Email Address []:
  
  Please enter the following 'extra' attributes
  to be sent with your certificate request
  A challenge password []:
  An optional company name []:
  ```
  
  使用根密钥和根证书为服务端签发证书，这里设置了证书的有效期为1年：
  
  ```bash
  [root@VM_4_242_centos /usr1/key]# openssl x509 -req -sha256 -CA ca.pem -CAkey ca.key -CAcreateserial -days 365 -in server/server.csr -out server/server.pem
  Signature ok
  subject=/C=XX/L=Default City/O=Default Company Ltd/CN=demo
  Getting CA Private Key
  ```
  
  生成客户端的密钥和证书
  生成2048位密钥：
  
  ```bash
  [root@VM_4_242_centos /usr1/key]# mkdir client
  [root@VM_4_242_centos /usr1/key]# openssl genrsa -out client/client.key 2048
  Generating RSA private key, 2048 bit long modulus
  .....................+++
  ..............................................................................................+++
  e is 65537 (0x10001)
  ```
  
  生成证书签发请求，如果只是作为客户端，Common Name可以随便命名：
  
  ```bash
  [root@VM_4_242_centos /usr1/key]# openssl req -new -key client/client.key -out client/client.csr               
  You are about to be asked to enter information that will be incorporated
  into your certificate request.
  What you are about to enter is what is called a Distinguished Name or a DN.
  There are quite a few fields but you can leave some blank
  For some fields there will be a default value,
  If you enter '.', the field will be left blank.
  -----
  Country Name (2 letter code) [XX]:
  State or Province Name (full name) []:
  Locality Name (eg, city) [Default City]:
  Organization Name (eg, company) [Default Company Ltd]:
  Organizational Unit Name (eg, section) []:
  Common Name (eg, your name or your server's hostname) []:demo
  Email Address []:
  
  Please enter the following 'extra' attributes
  to be sent with your certificate request
  A challenge password []:
  An optional company name []:
  ```
  
  使用根密钥和根证书为客户端签发证书，这里设置了证书的有效期为1年：
  
  ```bash
  [root@VM_4_242_centos /usr1/key]# openssl x509 -req -sha256 -CA ca.pem -CAkey ca.key -CAcreateserial -days 365 -in client/client.csr -out client/client.pem               
  Signature ok
  subject=/C=XX/L=Default City/O=Default Company Ltd/CN=demo
  Getting CA Private Key
  ```

3. grpc server双向认证实现
服务端需要3个文件：根证书（ca.pem），服务端私钥（server.key），服务端证书（server.pem）
```go
// 加载服务端私钥和证书
cert, err := tls.LoadX509KeyPair("server.pem", "server.key")
if err != nil {
  panic(err)
}

// 生成证书池，将根证书加入证书池
certPool := x509.NewCertPool()
rootBuf, err := ioutil.ReadFile("ca.pem")
if err != nil {
  panic(err)
}
if !certPool.AppendCertsFromPEM(rootBuf) {
  panic("Fail to append ca")
}

// 初始化TLSConfig
// ClientAuth有5种类型，如果要进行双向认证必须是RequireAndVerifyClientCert
tlsConf := &tls.Config{
  ClientAuth:   tls.RequireAndVerifyClientCert,
  Certificates: []tls.Certificate{cert},
  ClientCAs:    certPool,
}

// 开启服务端监听
listener, err := net.Listen("tcp", "127.0.0.1:8000")
if err != nil {
  panic(err)
}
defer listener.Close()

// 创建grpc server
server := grpc.NewServer(grpc.Creds(credentials.NewTLS(tlsConf)))

// 注册gprc server，这部分代码仅供参考
demo.RegisterDemoServer(server, &DemoServerImpl{})

// 启动grpc服务
server.Serv(listener)
```
	
4. grpc client双向认证实现
客户端需要3个文件：根证书（ca.pem），客户端私钥（client.key），客户端证书（client.pem）
```go


	// 加载客户端私钥和证书
	cert, err := tls.LoadX509KeyPair("client.pem", "client.key")
	if err != nil {
		panic(err)
	}

	// 将根证书加入证书池
	certPool := x509.NewCertPool()
	rootBuf, err := ioutil.ReadFile("ca.pem")
	if err != nil {
		panic(err)
	}
	if !certPool.AppendCertsFromPEM(rootBuf) {
		panic("Fail to append ca")
	}

	// 新建凭证
	// 注意ServerName需要与服务器证书内的Common Name一致
    // 客户端是根据根证书和ServerName对服务端进行验证的
	creds := credentials.NewTLS(&tls.Config{
		ServerName:   "demo",
		Certificates: []tls.Certificate{cert},
		RootCAs:      certPool,
	})

	// 不使用认证建立连接
	conn, err := grpc.Dial("127.0.0.1:8000", grpc.WithTransportCredentials(creds))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// 创建 gRPC 客户端实例，这部分代码仅供参考
	grpcClient := demo.NewDemoClient(conn)
```
