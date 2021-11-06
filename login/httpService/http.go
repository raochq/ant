package httpService

import (
	"fmt"
	"github.com/raochq/ant/engine/logger"
	"net"
	"net/http"
)

func StartHttpService(port, zoneID uint32) (*http.Server, error) {
	addr := fmt.Sprintf("0.0.0.0:%d", port)
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
			logger.Error("http.ListenAndServe(\"%s\") failed (%v)", addr, err)
		}
	}()
	srv.SetKeepAlivesEnabled(true)
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
