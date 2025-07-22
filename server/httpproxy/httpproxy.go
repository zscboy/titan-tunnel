package httpproxy

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	anonymousUserName = "anonymous"
)

type HttpProxy struct {
	tunMgr *TunnelManager
}

func newHttProxy(tunMgr *TunnelManager) *HttpProxy {
	return &HttpProxy{tunMgr: tunMgr}
}

func (httpProxy *HttpProxy) HandleProxy(w http.ResponseWriter, r *http.Request) {
	logx.Debug("HandleProxy")
	auth := r.Header.Get("Proxy-Authorization")
	userName, ok := httpProxy.checkAuth(auth)
	if !ok {
		w.Header().Set("Proxy-Authenticate", `Basic realm="Restricted"`)
		w.WriteHeader(http.StatusProxyAuthRequired)
		_, _ = w.Write([]byte("407 Proxy Authentication Required\n"))
		return
	}

	if r.Method == http.MethodConnect {
		httpProxy.handleHTTPS(w, r, userName)
	} else {
		httpProxy.handleHTTP(w, r, userName)
	}
}

func (httpProxy *HttpProxy) handleHTTPS(w http.ResponseWriter, r *http.Request, userName string) {
	hij, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hij.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	target := r.Host
	serverConn, err := net.Dial("tcp", target)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		clientConn.Close()
		return
	}

	_, _ = clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

	go io.Copy(serverConn, clientConn)
	go io.Copy(clientConn, serverConn)
}

func (httpProxy *HttpProxy) handleHTTP(w http.ResponseWriter, r *http.Request, userName string) {
	// req, err := http.NewRequest(r.Method, r.RequestURI, r.Body)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusBadRequest)
	// 	return
	// }
	// req.Header = r.Header

	// resp, err := http.DefaultTransport.RoundTrip(req)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusBadGateway)
	// 	return
	// }
	// defer resp.Body.Close()

	// for k, vv := range resp.Header {
	// 	for _, v := range vv {
	// 		w.Header().Add(k, v)
	// 	}
	// }
	// w.WriteHeader(resp.StatusCode)
	// io.Copy(w, resp.Body)
	// req := &ProxyRequest{
	// 	ID:      uuid.NewString(),
	// 	Method:  r.Method,
	// 	URL:     r.RequestURI,
	// 	Header:  r.Header,
	// 	BodyLen: r.ContentLength,
	// }
	hij, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, bufrw, err := hij.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer clientConn.Close()

	req := ProxyRequest{
		ID:     uuid.NewString(),
		Method: r.Method,
		URL:    r.URL.String(),
		Header: r.Header,
	}

	targetInfo := &TargetInfo{conn: clientConn, req: &req, userName: userName}
	if bufrw.Reader.Buffered() > 0 {
		targetInfo.extraBytes, err = httpProxy.drainBuffered(bufrw)
		if err != nil {
			logx.Errorf("drainBuffered %v", err)
			return
		}
	}

	if err := httpProxy.tunMgr.onHTTPRequest(targetInfo); err != nil {
		logx.Errorf("onHTTPRequest %v", err)
	}
}

func (httpProxy *HttpProxy) drainBuffered(bufrw *bufio.ReadWriter) ([]byte, error) {
	buffered := bufrw.Reader.Buffered()
	extraBytes := make([]byte, buffered)
	n, err := bufrw.Reader.Read(extraBytes)
	if err != nil {
		return nil, fmt.Errorf("localhttp.drainBuffered read bufrw failed:%s", err)
	}

	if n != buffered {
		return nil, fmt.Errorf("localhttp.ServeHTTP read bufrw not match, expected:%d, read:%d", buffered, n)
	}

	return extraBytes, nil
}

func (httpProxy *HttpProxy) checkAuth(auth string) (string, bool) {
	// if !strings.HasPrefix(auth, "Basic ") {
	// 	return false
	// }
	// payload, err := base64.StdEncoding.DecodeString(auth[len("Basic "):])
	// if err != nil {
	// 	return false
	// }
	// pair := strings.SplitN(string(payload), ":", 2)
	// return len(pair) == 2 && pair[0] == authUser && pair[1] == authPass
	// _ = payload
	return anonymousUserName, true
}
