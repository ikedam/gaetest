package testutil

// aetest の高速化をはかります。

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/golang/protobuf/proto"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
)

var (
	inst aetest.Instance

	// ErrNotSupported は操作がサポートされていない場合に返るエラーです。
	ErrNotSupported = errors.New("Not supported in this environment")

	// APIPort は dev_appserver の API Port です。
	APIPort = 8090
)

func setupAppengine() {
	// noop
}

func teardownAppengine() {
	RefreshAppengineInstance()
}

func findDevAppserver() string {
	if orig := os.Getenv("APPENGINE_DEV_APPSERVER"); orig != "" {
		return orig
	}

	var path string
	if _path, err := exec.LookPath("dev_appserver.py"); err == nil {
		path = _path
	} else {
		panic(err)
	}
	return path
}

func findDevAppserverWrapper() string {
	_, me, _, _ := runtime.Caller(0)
	wrapper := filepath.Join(filepath.Dir(me), "dev_appserver_wrapper.py")
	return wrapper
}

// saveEnv は指定の環境変数を復旧するためのクロージャを返します。
// defer で呼び出してください。
func saveEnv(envs ...string) func() {
	envsToRestore := []string{}
	savedEnvs := map[string]string{}
	for _, env := range envs {
		if val, ok := os.LookupEnv(env); ok {
			envsToRestore = append(envsToRestore, env)
			savedEnvs[env] = val
		} else {
			envsToRestore = append(envsToRestore, env)
		}
	}
	return func() {
		for _, env := range envsToRestore {
			if val, ok := savedEnvs[env]; ok {
				if err := os.Setenv(env, val); err != nil {
					panic(err)
				}
			} else {
				if err := os.Unsetenv(env); err != nil {
					panic(err)
				}
			}
		}
	}
}

// NewInstance はテスト用に最適化したオプションで GAE のインスタンスを起動します。
func NewInstance(opts *aetest.Options) (aetest.Instance, error) {
	restoreEnv := saveEnv(
		"APPENGINE_DEV_APPSERVER",
		"APPENGINE_DEV_APPSERVER_BASE",
		"DEV_APPSERVER_API_PORT",
	)
	defer restoreEnv()

	if err := os.Setenv("APPENGINE_DEV_APPSERVER_BASE", findDevAppserver()); err != nil {
		panic(err)
	}
	if err := os.Setenv("APPENGINE_DEV_APPSERVER", findDevAppserverWrapper()); err != nil {
		panic(err)
	}
	if err := os.Setenv("DEV_APPSERVER_API_PORT", fmt.Sprintf("%v", APIPort)); err != nil {
		panic(err)
	}
	return aetest.NewInstance(opts)
}

// GetAppengineInstance はテスト用の GAE のインスタンスを返します。
// aetest.NewInstance と同じですが、インスタンスの使い回しをするので高速です。
// 新規のインスタンスが必要な場合、事前に RefreshAppengineInstance を呼び出してください。
func GetAppengineInstance() aetest.Instance {
	if inst != nil {
		return inst
	}
	var err error
	inst, err = NewInstance(&aetest.Options{
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

// DatastoreClear は Datastore を初期化します。
func DatastoreClear() error {
	url := fmt.Sprintf("http://localhost:%v/clear?stub=datastore_v3", APIPort)
	var resp *http.Response
	if _resp, err := http.Get(url); err != nil {
		resp = _resp
	} else {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("/clear returned bad status: %v", resp.StatusCode)
	}
	return nil
}

func findPython() string {
	var path string
	if _path, err := exec.LookPath("python"); err == nil {
		path = _path
	} else {
		panic(err)
	}
	return path
}

func findAppcfg() string {
	var path string
	if _path, err := exec.LookPath("appcfg.py"); err == nil {
		path = _path
	} else {
		panic(err)
	}
	return path
}

// DatastoreDump は Datastore をダンプします。
func DatastoreDump(filename string, namespace string, kinds []string) error {
	var tempDir string
	if _tempDir, err := ioutil.TempDir("", "testutil"); err == nil {
		tempDir = _tempDir
	} else {
		return err
	}

	defer os.RemoveAll(tempDir)

	cmd := exec.Command(
		findPython(),
		findAppcfg(),
		"download_data",
		fmt.Sprintf("--url=http://localhost:%v/", APIPort),
		fmt.Sprintf("--filename=%v", filename),
		fmt.Sprintf("--namespace=%v", namespace),
		fmt.Sprintf("--kind=(%v)", strings.Join(kinds, ",")),
		fmt.Sprintf("--log_file=%v", filepath.Join(tempDir, "bulkloader.log")),
		fmt.Sprintf("--db_filename=%v", filepath.Join(tempDir, "bulkloader-progress.sql3")),
		fmt.Sprintf("--result_db_filename=%v", filepath.Join(tempDir, "bulkloader-results.sql3")),
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("Failed in download_data: %v output: %v", err, output)
	}

	return nil
}

// DatastoreRestore は Datastore をダンプから復旧します。
func DatastoreRestore(filename string, namespace string) error {
	var tempDir string
	if _tempDir, err := ioutil.TempDir("", "testutil"); err == nil {
		tempDir = _tempDir
	} else {
		return err
	}

	defer os.RemoveAll(tempDir)

	cmd := exec.Command(
		findPython(),
		findAppcfg(),
		"upload_data",
		fmt.Sprintf("--url=http://localhost:%v/", APIPort),
		fmt.Sprintf("--filename=%v", filename),
		fmt.Sprintf("--namespace=%v", namespace),
		fmt.Sprintf("--log_file=%v", filepath.Join(tempDir, "bulkloader.log")),
		fmt.Sprintf("--db_filename=%v", filepath.Join(tempDir, "bulkloader-progress.sql3")),
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("Failed in upload_data: %v output: %v", err, output)
	}

	return nil
}

// LogLevel はログレベルです。
type LogLevel int

const (
	// LogLevelDebug はデバッグレベルのログを識別します。
	LogLevelDebug = iota
	// LogLevelInfo は情報レベルのログを識別します。
	LogLevelInfo
	// LogLevelWarning は警告レベルのログを識別します。
	LogLevelWarning
	// LogLevelError はエラーレベルのログを識別します。
	LogLevelError
	// LogLevelCritical は致命レベルのログを識別します。
	LogLevelCritical
)

type logRecord struct {
	level   LogLevel
	message string
}

// AppengineMock は Appengine の API のモック化の機能を
// 提供します。
type AppengineMock struct {
	mockList []AppengineAPICallMock
	logList  []logRecord
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

var appengineMockKey = "github/ikedam/gaetest/testutil:AppengineMock"

// NewAppengineMock は新しい AppengineMock を返します。
func NewAppengineMock() *AppengineMock {
	return new(AppengineMock)
}

// MockContext は Appengine の Context をモック化します。
func (m *AppengineMock) MockContext(ctx context.Context) context.Context {
	if ctx.Value(&appengineMockKey) != nil {
		// already mocked
		return ctx
	}
	f := func(ctx context.Context, service, method string, in, out proto.Message) error {
		return m.apiCall(ctx, service, method, in, out)
	}
	return context.WithValue(
		appengine.WithAPICallFunc(ctx, f),
		&appengineMockKey,
		ctx,
	)
}

type mockInstance struct {
	base        aetest.Instance
	mocker      *AppengineMock
	releaseList []func()
}

// MockInstance は Appengine の Instance をモック化します。
// テストの完了時に Close を呼び出してください。
// モック元の Instance の Close は（必要であれば）別途呼び出す必要があります。
// また、本処理は Appengine SDK の内部処理に依存しているため、
// サポートされない合があります。
// その場合、単に nil を返しますので、テストをスキップしてください。
func (m *AppengineMock) MockInstance(inst aetest.Instance) aetest.Instance {
	return m.mockInstance(inst)
}

// NewRequest はモック化されたコンテキストを返すリクエストを作成します。
func (i *mockInstance) NewRequest(method, urlStr string, body io.Reader) (*http.Request, error) {
	return i.newRequest(method, urlStr, body)
}

func (i *mockInstance) Close() (err error) {
	for _, f := range i.releaseList {
		f()
	}
	return nil
}

// AddAPICallMock は API 呼び出しのモック処理を追加します。
func (m *AppengineMock) AddAPICallMock(mock AppengineAPICallMock) {
	m.mockList = append(m.mockList, mock)
}

func getOriginalContext(ctx context.Context) context.Context {
	_baseCtx := ctx.Value(&appengineMockKey)
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
	for i := range m.mockList {
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

// GetLogsEqualTo はモック化したインスタンスで取得した、
// 指定レベルのログを返します。
// Appengine SDK の内部実装に依存しているため、ログを取得できない場合もあります。
// ログが何も取得できない場合は結果の判定を行わないでください。
func (m *AppengineMock) GetLogsEqualTo(level LogLevel) []string {
	var ret []string
	for _, record := range m.logList {
		if record.level == level {
			ret = append(ret, record.message)
		}
	}
	return ret
}

// GetLogsEqualOrMore はモック化したインスタンスで取得した、
// 指定レベルのログを返します。
// Appengine SDK の内部実装に依存しているため、ログを取得できない場合もあります。
// ログが何も取得できない場合は結果の判定を行わないでください。
func (m *AppengineMock) GetLogsEqualOrMore(level LogLevel) []string {
	var ret []string
	for _, record := range m.logList {
		if record.level >= level {
			ret = append(ret, record.message)
		}
	}
	return ret
}
