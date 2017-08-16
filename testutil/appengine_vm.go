// +build !appengine

package testutil

import (
	"io"
	"net/http"

	"google.golang.org/appengine/aetest"
)

func (m *AppengineMock) mockInstance(inst aetest.Instance) aetest.Instance {
	return nil
}

func (i *mockInstance) newRequest(method, urlStr string, body io.Reader) (*http.Request, error) {
	// As google.golang.org/appengine/internal.RegisterTestRequest requires apiURL,
	// which is stored in the internal structure of aetest,
	// NewRequest cannot be supported.
	return nil, ErrNotSupported
}
