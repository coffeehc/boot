package client

import (
	"encoding/base64"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

func closeBody(body io.ReadCloser) {
	if body != nil {
		body.Close()
	}
}

func cloneHeader(h http.Header) http.Header {
	h2 := make(http.Header, len(h))
	for k, vv := range h {
		vv2 := make([]string, len(vv))
		copy(vv2, vv)
		h2[k] = vv2
	}
	return h2
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func setRequestCancel(req *http.Request, rt http.RoundTripper, deadline time.Time) (stopTimer func(), wasCanceled func() bool) {
	if deadline.IsZero() {
		return nop, alwaysFalse
	}

	initialReqCancel := req.Cancel // the user's original Request.Cancel, if any

	cancel := make(chan struct{})
	req.Cancel = cancel

	wasCanceled = func() bool {
		select {
		case <-cancel:
			return true
		default:
			return false
		}
	}

	doCancel := func() {
		// The new way:
		close(cancel)

		// The legacy compatibility way, used only
		// for RoundTripper implementations written
		// before Go 1.5 or Go 1.6.
		type canceler interface {
			CancelRequest(*http.Request)
		}
		switch v := rt.(type) {
		case *http.Transport: //, *http2Transport:
		// Do nothing. The net/http package's transports
		// support the new Request.Cancel channel
		case canceler:
			v.CancelRequest(req)
		}
	}

	stopTimerCh := make(chan struct{})
	var once sync.Once
	stopTimer = func() { once.Do(func() { close(stopTimerCh) }) }

	timer := time.NewTimer(deadline.Sub(time.Now()))
	var timedOut atomicBool

	go func() {
		select {
		case <-initialReqCancel:
			doCancel()
			timer.Stop()
		case <-timer.C:
			timedOut.setTrue()
			doCancel()
		case <-stopTimerCh:
			timer.Stop()
		}
	}()

	return stopTimer, wasCanceled
}

type atomicBool int32

func (b *atomicBool) isSet() bool { return atomic.LoadInt32((*int32)(b)) != 0 }
func (b *atomicBool) setTrue()    { atomic.StoreInt32((*int32)(b), 1) }

func nop() {}

func alwaysFalse() bool { return false }

// cancelTimerBody is an io.ReadCloser that wraps rc with two features:
// 1) on Read error or close, the stop func is called.
// 2) On Read failure, if reqWasCanceled is true, the error is wrapped and
//    marked as net.Error that hit its timeout.
type cancelTimerBody struct {
	stop           func() // stops the time.Timer waiting to cancel the request
	rc             io.ReadCloser
	reqWasCanceled func() bool
}

func (b *cancelTimerBody) Read(p []byte) (n int, err error) {
	n, err = b.rc.Read(p)
	if err == nil {
		return n, nil
	}
	b.stop()
	if err == io.EOF {
		return n, err
	}
	if b.reqWasCanceled() {
		err = &httpError{
			err:     err.Error() + " (Client.Timeout exceeded while reading body)",
			timeout: true,
		}
	}
	return n, err
}

func (b *cancelTimerBody) Close() error {
	err := b.rc.Close()
	b.stop()
	return err
}

type httpError struct {
	err     string
	timeout bool
}

func (e *httpError) Error() string   { return e.err }
func (e *httpError) Timeout() bool   { return e.timeout }
func (e *httpError) Temporary() bool { return true }

func deadline(timeout time.Duration) time.Time {
	if timeout > 0 {
		return time.Now().Add(timeout)
	}
	return time.Time{}
}

func valueOrDefault(value, def string) string {
	if value != "" {
		return value
	}
	return def
}
