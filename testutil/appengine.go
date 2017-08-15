package testutil

// aetest の高速化をはかります。

import (
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
)

var (
	inst aetest.Instance
)

func setupAppengine() {
	// noop
}

func teardownAppengine() {
	RefreshAppengineInstance()
}

// GetAppengineInstance はテスト用の GAE のインスタンスを返します。
// aetest.NewInstance と同じですが、インスタンスの使い回しをするので高速です。
// 新規のインスタンスが必要な場合、事前に RefreshAppengineInstance を呼び出してください。
func GetAppengineInstance() aetest.Instance {
	if inst != nil {
		return inst
	}
	var err error
	inst, err = aetest.NewInstance(&aetest.Options{
		StronglyConsistentDatastore: true,
	})
	if err != nil {
		panic(err)
	}
	return inst
}

// RefreshAppengineInstance は使用中の GAE インスタンスを破棄し、
// 次から新しいインスタンスを利用します。
func RefreshAppengineInstance() {
	if inst == nil {
		return
	}
	inst.Close()
	inst = nil
}

// GetAppengineContext はテスト用の新しい GAE コンテキストを取得します。
// aetest.NewContext とほぼ同等ですが、インスタンスを再利用します。
// また、そのため終了処理が必要ありません。
func GetAppengineContext() context.Context {
	return GetAppengineContextFor(GetAppengineInstance())
}

// GetAppengineContextFor は指定のインスタンスに対する新しい GAE コンテキストを取得します。
func GetAppengineContextFor(inst aetest.Instance) context.Context {
	req, err := inst.NewRequest("GET", "/", nil)
	if err != nil {
		panic(err)
	}
	return appengine.NewContext(req)
}

// AppengineMock は Appengine の API のモック化の機能を
// 提供します。
type AppengineMock struct {
	mockList []AppengineAPICallMock
}

// AppengineAPICallMock は Appengine の API 呼び出しのモック化の方法を設定します。
type AppengineAPICallMock struct {
	// Count はモックを実行する回数を設定します。
	// 0 を設定すると永久に繰り返します。
	Count int

	// Service はモック化対象のサービスを設定します。
	// 設定しない場合、全サービスが対象になります。
	// 先頭一致でチェックするので、 datastore などを設定すれば
	// datastore_v3 なども対象になります。
	Service string

	// Method はモック化対象のメソッドを設定します。
	// 設定しない場合、全メソッドが対象になります。
	// 完全一致でチェックします。
	Method string

	// Error は API 呼び出しを error にする場合に設定します。
	Error error
}

func (apiMock AppengineAPICallMock) apiCall(ctx context.Context, service, method string, in, out proto.Message) error {
	if apiMock.Error != nil {
		return apiMock.Error
	}
	return callOriginalAPICall(ctx, service, method, in, out)
}

const (
	appengineMockKey = "github/ikedam/gaetest/testutil:AppengineMock"
)

// NewAppengineMock は新しい AppengineMock を返します。
func NewAppengineMock() *AppengineMock {
	return new(AppengineMock)
}

// MockContext は Appengine の Context をモック化します。
func (m *AppengineMock) MockContext(ctx context.Context) context.Context {
	if ctx.Value(appengineMockKey) != nil {
		// already mocked
		return ctx
	}
	f := func(ctx context.Context, service, method string, in, out proto.Message) error {
		return m.apiCall(ctx, service, method, in, out)
	}
	return context.WithValue(
		appengine.WithAPICallFunc(ctx, f),
		appengineMockKey,
		ctx,
	)
}

// AddAPICallMock は API 呼び出しのモック処理を追加します。
func (m *AppengineMock) AddAPICallMock(mock AppengineAPICallMock) {
	m.mockList = append(m.mockList, mock)
}

func getOriginalContext(ctx context.Context) context.Context{
	_baseCtx := ctx.Value(appengineMockKey)
	if _baseCtx != nil {
		return ctx
	}
	baseCtx, ok := _baseCtx.(context.Context)
	if !ok {
		panic(fmt.Sprintf("Unknown base context: %v (%T)", _baseCtx, _baseCtx))
	}
	return baseCtx
}

func callOriginalAPICall(ctx context.Context, service, method string, in, out proto.Message) error {
	return appengine.APICall(getOriginalContext(ctx), service, method, in, out)
}

func (m *AppengineMock) apiCall(ctx context.Context, service, method string, in, out proto.Message) error {
	for i, _ := range m.mockList {
		apiMock := &m.mockList[i]
		if apiMock.Service != "" && !strings.HasPrefix(service, apiMock.Service) {
			// Service 不一致
			continue
		}
		if apiMock.Method != "" && method != apiMock.Method {
			// Method 不一致
			continue
		}
		err := apiMock.apiCall(ctx, service, method, in, out)
		if apiMock.Count > 0 {
			apiMock.Count = apiMock.Count - 1
			if apiMock.Count <= 0 {
				m.mockList = append(
					m.mockList[:i],
					m.mockList[i+1:]...,
				)
			}
		}
		return err
	}
	return callOriginalAPICall(ctx, service, method, in, out)
}
