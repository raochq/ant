package service

import (
	"context"
	"time"

	clientV3 "go.etcd.io/etcd/client/v3"

	"github.com/raochq/ant/util/logger"
)

type ETCDClient struct {
	client  *clientV3.Client
	leaseId clientV3.LeaseID
	key     string

	ctx context.Context
}

func NewETCDClient(ctx context.Context, key string, endpoints []string) (*ETCDClient, error) {
	cli, err := clientV3.New(clientV3.Config{
		Context:     ctx,
		Endpoints:   endpoints,
		DialTimeout: 2 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	svr := &ETCDClient{
		ctx:    ctx,
		client: cli,
		key:    key,
	}
	return svr, nil
}

func (s *ETCDClient) keepAlive(value string) (<-chan *clientV3.LeaseKeepAliveResponse, error) {
	resp, err := s.client.Grant(context.TODO(), 5)
	if err != nil {
		return nil, err
	}

	_, err = s.client.Put(context.TODO(), s.key, value, clientV3.WithLease(resp.ID))
	if err != nil {
		return nil, err
	}

	s.leaseId = resp.ID
	return s.client.KeepAlive(s.ctx, s.leaseId)
}

func (s *ETCDClient) revoke() {
	s.client.Revoke(s.ctx, s.leaseId)
	logger.WithField("key", s.key).Debug("etcd client revoke")
}

func (s *ETCDClient) Update(value string) {
	if s == nil {
		return
	}
	_, err := s.client.Put(context.TODO(), s.key, value, clientV3.WithLease(s.leaseId))
	if err != nil {
		logger.WithError(err).WithField("key", s.key).Error("etcd put failed")
	}
	logger.Debug(value)
}

func (s *ETCDClient) Start(svrInfo string, callback func()) error {
	if s == nil {
		return nil
	}
	ch, err := s.keepAlive(svrInfo)
	if err != nil {
		return err
	}
	go s.loop(ch, callback)
	return nil
}

func (s *ETCDClient) loop(ch <-chan *clientV3.LeaseKeepAliveResponse, callback func()) {
	defer callback()
	defer s.Close()
	for {
		select {
		case <-s.ctx.Done():
			logger.Info(s.ctx.Err())
			s.revoke()
			return
		case ka, ok := <-ch:
			if !ok {
				logger.WithField("key", s.key).Info("keep alive channel closed")
				s.revoke()
				return
			} else {
				logger.WithField("key", s.key).WithField("ttl", ka.TTL).Debug("Recv reply from service")
			}
		}
	}
}
func (s *ETCDClient) Close() {
	s.client.Close()
}
