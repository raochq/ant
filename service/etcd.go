package service

import (
	"context"
	"log/slog"
	"time"

	clientV3 "go.etcd.io/etcd/client/v3"
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
	resp, err := s.client.Grant(context.TODO(), 10)
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
	slog.Debug("etcd client revoke", "key", s.key)
}

func (s *ETCDClient) Update(value string) {
	if s == nil {
		return
	}
	_, err := s.client.Put(context.TODO(), s.key, value, clientV3.WithLease(s.leaseId))
	if err != nil {
		slog.Error("etcd put failed", "key", s.key, "value", value, "error", err)
	}
	slog.Debug(value)
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
			s.revoke()
			return
		case ka, ok := <-ch:
			if !ok {
				slog.Info("keep alive channel closed", "key", s.key)
				s.revoke()
				return
			}

			slog.Debug("Recv reply from service", "key", s.key, "ttl", ka.TTL)
		}
	}
}
func (s *ETCDClient) Close() {
	s.client.Close()
}
