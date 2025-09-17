package flnwallet

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/flokiorg/flnd"
	"github.com/flokiorg/flnd/signal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	ErrDaemonNotRunning = errors.New("daemon is not running")
)

const (
	FlndEndpoint = "localhost:10005"
)

type daemon struct {
	config      *flnd.Config
	interceptor signal.Interceptor

	conn *grpc.ClientConn

	ctx    context.Context
	cancel context.CancelFunc
	closed bool
	mu     sync.Mutex
	wg     sync.WaitGroup
	client *Client
}

func newDaemon(pctx context.Context, config *flnd.Config) *daemon {

	ctx, cancel := context.WithCancel(pctx)

	interceptor, _ := signal.Intercept()

	return &daemon{
		config:      config,
		ctx:         ctx,
		cancel:      cancel,
		interceptor: interceptor,
	}
}

func (d *daemon) start() (c *Client, err error) {

	d.config, err = flnd.ValidateConfig(*d.config, d.interceptor, nil, nil)
	if err != nil {
		err = fmt.Errorf("failed to load config: %v", err)
		return
	}

	impl := d.config.ImplementationConfig(d.interceptor)
	defer func() {
		if err != nil {
			d.stop()
		}
	}()

	if err = d.exec(impl); err != nil {
		return
	}

	var creds credentials.TransportCredentials
	creds, err = tlsCreds(d.config.TLSCertPath)
	if err != nil {
		return nil, err
	}

	if len(d.config.RPCListeners) == 0 {
		return nil, fmt.Errorf("unable to open rpc connection, rpc listener is empty")
	}

	d.conn, err = grpc.NewClient(d.config.RPCListeners[0].String(), grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, err
	}

	d.client = NewClient(d.ctx, d.conn, d.config)
	c = d.client
	return
}

func (d *daemon) exec(impl *flnd.ImplementationCfg) error {

	errCh := make(chan error)
	flndStarted := make(chan struct{})

	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		if err := flnd.Main(d.config, flnd.ListenerCfg{}, impl, d.interceptor, flndStarted); err != nil {
			select {
			case errCh <- err:
			default:
				if d.client != nil {
					d.client.kill(err)
				}
			}
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-flndStarted:
		return nil
	}
}

func (d *daemon) waitForShutdown() {
	d.wg.Wait()
	d.closed = true
}

func (d *daemon) stop() {

	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return
	}

	if d.client != nil {
		d.client.close()
	}

	if d.conn != nil {
		d.conn.Close()
	}

	d.cancel()
	d.interceptor.RequestShutdown()
	<-d.interceptor.ShutdownChannel()

}

func tlsCreds(certPath string) (credentials.TransportCredentials, error) {
	pem, err := os.ReadFile(certPath)
	if err != nil {
		return nil, err
	}
	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(pem) {
		return nil, errors.New("failed to parse cert")
	}
	return credentials.NewClientTLSFromCert(cp, ""), nil
}
