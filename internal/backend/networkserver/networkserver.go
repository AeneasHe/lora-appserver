package networkserver

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"sync"
	"time"

	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/brocaar/lora-app-server/internal/config"
	"github.com/brocaar/loraserver/api/ns"
)

var p Pool

// Pool defines the network-server client pool.
type Pool interface {
	Get(hostname string, caCert, tlsCert, tlsKey []byte) (ns.NetworkServerServiceClient, error)
}

type client struct {
	client     ns.NetworkServerServiceClient //lora network的grpc服务
	clientConn *grpc.ClientConn              //lora network的grpc连接
	caCert     []byte
	tlsCert    []byte
	tlsKey     []byte
}

// Setup configures the networkserver package.
func Setup(conf config.Config) error {
	//pool是服务器客户端连接池
	p = &pool{
		clients: make(map[string]client),
	}
	return nil
}

// GetPool returns the networkserver pool.
func GetPool() Pool {
	return p
}

// SetPool sets the network-server pool.
func SetPool(pp Pool) {
	p = pp
}

type pool struct {
	sync.RWMutex                   //读写锁，写锁定时是阻塞的不能进行其他读写，没有写锁时可以进行多次读
	clients      map[string]client //客户端表
}

// Get returns a NetworkServerClient for the given server (hostname:ip).
func (p *pool) Get(hostname string, caCert, tlsCert, tlsKey []byte) (ns.NetworkServerServiceClient, error) {
	defer p.Unlock()
	p.Lock()

	var connect bool
	c, ok := p.clients[hostname]
	if !ok {
		connect = true
	}

	// if the connection exists in the map, but when the certificates changed
	// try to close the connection and re-connect
	if ok && (!bytes.Equal(c.caCert, caCert) || !bytes.Equal(c.tlsCert, tlsCert) || !bytes.Equal(c.tlsKey, tlsKey)) {
		c.clientConn.Close()
		delete(p.clients, hostname)
		connect = true
	}

	//如果客户端在连接池中
	if connect {
		//建立该客户端的连接
		clientConn, nsClient, err := p.createClient(hostname, caCert, tlsCert, tlsKey)
		if err != nil {
			return nil, errors.Wrap(err, "create network-server api client error")
		}
		c = client{
			client:     nsClient,
			clientConn: clientConn,
			caCert:     caCert,
			tlsCert:    tlsCert,
			tlsKey:     tlsKey,
		}
		//更新客户端连接
		p.clients[hostname] = c
	}

	return c.client, nil
}

func (p *pool) createClient(hostname string, caCert, tlsCert, tlsKey []byte) (*grpc.ClientConn, ns.NetworkServerServiceClient, error) {
	logrusEntry := log.NewEntry(log.StandardLogger())
	logrusOpts := []grpc_logrus.Option{
		grpc_logrus.WithLevels(grpc_logrus.DefaultCodeToLevel),
	}

	//grpc调用配置切片
	nsOpts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithUnaryInterceptor(
			grpc_logrus.UnaryClientInterceptor(logrusEntry, logrusOpts...),
		),
		grpc.WithStreamInterceptor(
			grpc_logrus.StreamClientInterceptor(logrusEntry, logrusOpts...),
		),
	}

	if len(caCert) == 0 && len(tlsCert) == 0 && len(tlsKey) == 0 {
		//如果没有启用tls认证，给出警告
		nsOpts = append(nsOpts, grpc.WithInsecure())
		log.WithField("server", hostname).Warning("creating insecure network-server client")
	} else {
		log.WithField("server", hostname).Info("creating network-server client")
		cert, err := tls.X509KeyPair(tlsCert, tlsKey)
		if err != nil {
			return nil, nil, errors.Wrap(err, "load x509 keypair error")
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, nil, errors.Wrap(err, "append ca cert to pool error")
		}
		//将认证配置写入grpc配置切片中
		nsOpts = append(nsOpts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      caCertPool,
		})))
	}

	//设置请求上下文
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	//grpc连接
	nsClient, err := grpc.DialContext(ctx, hostname, nsOpts...)
	if err != nil {
		return nil, nil, errors.Wrap(err, "dial network-server api error")
	}

	//返回grpc请求及grpc服务
	return nsClient, ns.NewNetworkServerServiceClient(nsClient), nil
}
