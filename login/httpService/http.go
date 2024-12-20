package httpService

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
)

func StartHttpService(addr string, zoneID uint32) (*http.Server, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("LoginHttpService listen %s fail: %w", addr, err)
	}
	httpServeMux := registerHttpHandler(zoneID)
	srv := &http.Server{
		Addr:    l.Addr().String(),
		Handler: httpServeMux,
	}
	go func() {
		if err := srv.Serve(l); err != nil {
			slog.Error("http.ListenAndServe failed", "addr", addr, "error", err)
		}
	}()
	srv.SetKeepAlivesEnabled(true)
	slog.Info("Start http service", "addr", l.Addr())
	return srv, nil
}

// 注册http handler
func registerHttpHandler(zoneID uint32) *http.ServeMux {
	httpServeMux := http.NewServeMux()
	srv := NewLoginService(zoneID)
	httpServeMux.HandleFunc("/x/4/account/login", srv.UserLogin)
	httpServeMux.HandleFunc("/x/4/account/register", srv.UserRegister)
	httpServeMux.HandleFunc("/x/4/account/platformlogin", srv.UserLoginOrRegister)
	//httpServeMux.HandleFunc("/x/4/dummyrequest", httpService.DummyRequest)
	//httpServeMux.HandleFunc("/x/4/account/version", httpService.ServerVer)
	//httpServeMux.HandleFunc("/x/4/account/loginspec", httpService.UserLoginSpec)
	//httpServeMux.HandleFunc("/CheckColdUpdate", httpService.CheckColdUpdate)
	//httpServeMux.HandleFunc("/CheckHotfix", httpService.CheckHotfix)
	httpServeMux.HandleFunc("/", srv.NullRequest)
	return httpServeMux
}
