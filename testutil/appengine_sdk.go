// +build appengine

package testutil

import (
	sdk_appengine "appengine"
	"appengine_internal"

	"fmt"
	"io"
	"net/http"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
)

func (m *AppengineMock) mockInstance(inst aetest.Instance) aetest.Instance {
	return &mockInstance{
		base:   inst,
		mocker: m,
	}
}

// implements sdk_appengine.Context
type internal_context_mock struct {
	baseAppengineContext sdk_appengine.Context
	baseContext          context.Context
	req                  *http.Request
	mocker               *AppengineMock
}

func (i *mockInstance) newRequest(method, urlStr string, body io.Reader) (*http.Request, error) {
	baseReq, err := i.base.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, err
	}

	// as there's no way to replace the context already registered,
	// create a new one and register it.
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, err
	}

	baseAppengineContext := appengine_internal.NewContext(baseReq)
	baseContext := appengine.NewContext(baseReq)
	ctx := &internal_context_mock{
		baseAppengineContext: baseAppengineContext,
		baseContext:          i.mocker.MockContext(baseContext),
		req:                  req,
		mocker:               i.mocker,
	}

	r, release := appengine_internal.RegisterTestContext(req, ctx)
	i.releaseList = append(i.releaseList, release)

	return r, nil
}

func (c *internal_context_mock) Debugf(format string, args ...interface{}) {
	c.mocker.logList = append(c.mocker.logList, logRecord{
		level:   LogLevelDebug,
		message: fmt.Sprintf(format, args...),
	})
	c.baseAppengineContext.Debugf(format, args...)
}

func (c *internal_context_mock) Infof(format string, args ...interface{}) {
	c.mocker.logList = append(c.mocker.logList, logRecord{
		level:   LogLevelInfo,
		message: fmt.Sprintf(format, args...),
	})
	c.baseAppengineContext.Infof(format, args...)
}

func (c *internal_context_mock) Warningf(format string, args ...interface{}) {
	c.mocker.logList = append(c.mocker.logList, logRecord{
		level:   LogLevelWarning,
		message: fmt.Sprintf(format, args...),
	})
	c.baseAppengineContext.Warningf(format, args...)
}

func (c *internal_context_mock) Errorf(format string, args ...interface{}) {
	c.mocker.logList = append(c.mocker.logList, logRecord{
		level:   LogLevelError,
		message: fmt.Sprintf(format, args...),
	})
	c.baseAppengineContext.Errorf(format, args...)
}

func (c *internal_context_mock) Criticalf(format string, args ...interface{}) {
	c.mocker.logList = append(c.mocker.logList, logRecord{
		level:   LogLevelCritical,
		message: fmt.Sprintf(format, args...),
	})
	c.baseAppengineContext.Criticalf(format, args...)
}

func (c *internal_context_mock) Call(service, method string, in, out appengine_internal.ProtoMessage, opts *appengine_internal.CallOptions) error {
	// return c.baseAppengineContext.Call(service, method, in, out, opts)
	mocked := c.baseContext
	cancel := func() {}
	if opts != nil && opts.Timeout > 0 {
		mocked, cancel = context.WithTimeout(mocked, opts.Timeout)
	}
	defer cancel()
	return appengine.APICall(mocked, service, method, in, out)
}

func (c *internal_context_mock) FullyQualifiedAppID() string {
	return c.baseAppengineContext.FullyQualifiedAppID()
}

func (c *internal_context_mock) Request() interface{} {
	return c.req
}
