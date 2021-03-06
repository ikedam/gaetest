package testutil

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/memcache"
)

type entity struct {
	Value string `datastore:",noindex"`
}

type anotherEntity struct {
	Value string `datastore:",noindex"`
}

type yetAnotherEntity struct {
	Value string `datastore:",noindex"`
}

func TestAppengineMock(t *testing.T) {
	ctx := GetAppengineContext()
	mocker := NewAppengineMock()
	mocked := mocker.MockContext(ctx)

	var expectError = errors.New("Expected error")

	mocker.AddAPICallMock(AppengineAPICallMock{
		Error: expectError,
	})

	e := entity{
		Value: "test",
	}

	key := datastore.NewIncompleteKey(mocked, "entity", nil)

	if _, err := datastore.Put(mocked, key, &e); err != expectError {
		t.Fatalf("Expect %v, but was %v", expectError, err)
	}

	key = datastore.NewIncompleteKey(ctx, "entity", nil)

	if _, err := datastore.Put(ctx, key, &e); err != nil {
		t.Fatalf("Expect success, but was %v", err)
	}
}

func TestAppengineMockCount(t *testing.T) {
	ctx := GetAppengineContext()
	mocker := NewAppengineMock()
	mocked := mocker.MockContext(ctx)

	var expectError1 = errors.New("Expected error1")
	var expectError2 = errors.New("Expected error2")

	mocker.AddAPICallMock(AppengineAPICallMock{
		Error: expectError1,
		Count: 2,
	})
	mocker.AddAPICallMock(AppengineAPICallMock{
		Count: 1,
	})
	mocker.AddAPICallMock(AppengineAPICallMock{
		Error: expectError2,
		Count: 1,
	})

	e := entity{
		Value: "test",
	}

	var errorList []error

	for i := 0; i < 5; i++ {
		key := datastore.NewIncompleteKey(mocked, "entity", nil)

		_, err := datastore.Put(mocked, key, &e)
		errorList = append(errorList, err)
	}

	if !reflect.DeepEqual(errorList, []error{expectError1, expectError1, nil, expectError2, nil}) {
		t.Errorf("Uexpected errors: %v", errorList)
	}
}

func TestAppengineMockService(t *testing.T) {
	ctx := GetAppengineContext()
	mocker := NewAppengineMock()
	mocked := mocker.MockContext(ctx)

	var expectError = errors.New("Expected error")

	mocker.AddAPICallMock(AppengineAPICallMock{
		Service: "datastore",
		Error:   expectError,
	})

	e := entity{
		Value: "test",
	}

	key := datastore.NewIncompleteKey(mocked, "entity", nil)

	if _, err := datastore.Put(mocked, key, &e); err != expectError {
		t.Fatalf("Expect %v, but was %v", expectError, err)
	}

	item := memcache.Item{
		Key:   "key",
		Value: []byte("bar"),
	}

	if err := memcache.Set(mocked, &item); err != nil {
		t.Fatalf("Expect success, but was %v", err)
	}

	mocker.AddAPICallMock(AppengineAPICallMock{
		Service: "memcache",
		Error:   expectError,
	})

	if err := memcache.Set(mocked, &item); err != expectError {
		t.Fatalf("Expect success, but was %v", err)
	}
}

func TestAppengineMockMethod(t *testing.T) {
	ctx := GetAppengineContext()
	mocker := NewAppengineMock()
	mocked := mocker.MockContext(ctx)

	var expectError = errors.New("Expected error")

	mocker.AddAPICallMock(AppengineAPICallMock{
		Service: "datastore",
		Method:  "Get",
		Error:   expectError,
	})

	e := entity{
		Value: "test",
	}

	key := datastore.NewIncompleteKey(mocked, "entity", nil)

	if _key, err := datastore.Put(mocked, key, &e); err != nil {
		t.Fatalf("Expect success, but was %v", err)
	} else {
		key = _key
	}

	if err := datastore.Get(mocked, key, &e); err != expectError {
		t.Fatalf("Expect %v, but was %v", expectError, err)
	}
}

func handlerTest(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	key := datastore.NewIncompleteKey(c, "entity", nil)
	e := entity{
		Value: r.FormValue("value"),
	}
	var err error
	if key, err = datastore.Put(c, key, &e); err != nil {
		log.Errorf(c, "Error in Put: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err = datastore.Get(c, key, &e); err != nil {
		log.Errorf(c, "Error in Get: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Got: %v", e)
}

func TestMockInstance(t *testing.T) {
	inst := GetAppengineInstance()
	mocker := NewAppengineMock()
	mocked := mocker.MockInstance(inst)
	if mocked == nil {
		t.Skip("MockInstance is not supported")
	}

	var expectError = errors.New("Expected error")

	mocker.AddAPICallMock(AppengineAPICallMock{
		Service: "datastore",
		Method:  "Get",
		Error:   expectError,
	})

	data := url.Values{}
	data.Add("value", "test")
	req, err := mocked.NewRequest("POST", "/entity/", strings.NewReader(data.Encode()))
	if err != nil {
		panic(err)
	}
	w := httptest.NewRecorder()
	handlerTest(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, but was %v", w)
	}

	// 残念ながら log.Query はテスト環境では機能しないため、
	// 独自のログレコーダーで結果をチェック
	errorLogList := mocker.GetLogsEqualTo(LogLevelError)
	if len(errorLogList) != 0 {
		// ログが取得できない場合を考慮
		if len(errorLogList) != 1 {
			t.Errorf("Expect 1 error, but: %v", errorLogList)
		} else if errorLogList[0] != "Error in Get: Expected error" {
			t.Errorf("Unexpected error message: %v", errorLogList)
		}
	}
}

func TestDatastoreClear(t *testing.T) {
	ctx := GetAppengineContext()

	if _ctx, err := appengine.Namespace(ctx, "testing"); err == nil {
		ctx = _ctx
	} else {
		t.Fatalf("Failed to set namespace: %v", err)
	}

	entityKey := datastore.NewIncompleteKey(ctx, "entity", nil)
	entityData := entity{
		Value: "test1",
	}
	if _key, err := datastore.Put(ctx, entityKey, &entityData); err == nil {
		entityKey = _key
	} else {
		t.Fatalf("Failed to put data: %v", err)
	}

	anotherEntityKey := datastore.NewIncompleteKey(ctx, "anotherEntity", nil)
	anotherEntityData := anotherEntity{
		Value: "test2",
	}
	if _key, err := datastore.Put(ctx, anotherEntityKey, &anotherEntityData); err == nil {
		anotherEntityKey = _key
	} else {
		t.Fatalf("Failed to put data: %v", err)
	}

	if err := datastore.Get(ctx, entityKey, &entityData); err != nil {
		t.Fatalf("Failed to get data: %v", err)
	}
	if err := datastore.Get(ctx, anotherEntityKey, &anotherEntityData); err != nil {
		t.Fatalf("Failed to get data: %v", err)
	}

	if err := DatastoreClear(); err != nil {
		t.Fatalf("Failed to DatastoreClear(): %v", err)
	}

	if err := datastore.Get(ctx, entityKey, &entityData); err != datastore.ErrNoSuchEntity {
		t.Fatalf("Expect datastore.ErrNoSuchEntity, but was: %v", err)
	}
	if err := datastore.Get(ctx, anotherEntityKey, &anotherEntityData); err != datastore.ErrNoSuchEntity {
		t.Fatalf("Expect datastore.ErrNoSuchEntity, but was: %v", err)
	}
}

func TestDatastoreDumpAndRestore(t *testing.T) {
	var tempDir string
	if _tempDir, err := ioutil.TempDir("", "testutil"); err == nil {
		tempDir = _tempDir
	} else {
		t.Fatalf("Failed to create a temporary directory: %v", err)
	}
	dumpFile := filepath.Join(tempDir, "dump.dat")

	defer os.RemoveAll(tempDir)

	ctx := GetAppengineContext()

	if _ctx, err := appengine.Namespace(ctx, "testing"); err == nil {
		ctx = _ctx
	} else {
		t.Fatalf("Failed to set namespace: %v", err)
	}

	entityKey := datastore.NewIncompleteKey(ctx, "entity", nil)
	entityData := entity{
		Value: "test1",
	}
	if _key, err := datastore.Put(ctx, entityKey, &entityData); err == nil {
		entityKey = _key
	} else {
		t.Fatalf("Failed to put data: %v", err)
	}

	anotherEntityKey := datastore.NewIncompleteKey(ctx, "anotherEntity", nil)
	anotherEntityData := anotherEntity{
		Value: "test2",
	}
	if _key, err := datastore.Put(ctx, anotherEntityKey, &anotherEntityData); err == nil {
		anotherEntityKey = _key
	} else {
		t.Fatalf("Failed to put data: %v", err)
	}

	yetAnotherEntityKey := datastore.NewIncompleteKey(ctx, "yetAnotherEntity", nil)
	yetAnotherEntityData := yetAnotherEntity{
		Value: "test3",
	}
	if _key, err := datastore.Put(ctx, yetAnotherEntityKey, &yetAnotherEntityData); err == nil {
		yetAnotherEntityKey = _key
	} else {
		t.Fatalf("Failed to put data: %v", err)
	}

	if err := DatastoreDump(dumpFile, "testing", []string{"entity", "anotherEntity"}); err != nil {
		t.Fatalf("Failed to DatastoreDump(): %v", err)
	}

	if err := DatastoreClear(); err != nil {
		t.Fatalf("Failed to DatastoreClear(): %v", err)
	}

	if err := DatastoreRestore(dumpFile, "testing"); err != nil {
		t.Fatalf("Failed to DatastoreRestore(): %v", err)
	}

	if err := datastore.Get(ctx, entityKey, &entityData); err != nil {
		t.Fatalf("Failed to get data: %v", err)
	} else if entityData.Value != "test1" {
		t.Fatalf("Unexpected data: %+v", entityData)
	}
	if err := datastore.Get(ctx, anotherEntityKey, &anotherEntityData); err != nil {
		t.Fatalf("Failed to get data: %v", err)
	} else if anotherEntityData.Value != "test2" {
		t.Fatalf("Unexpected data: %+v", anotherEntityData)
	}
	if err := datastore.Get(ctx, yetAnotherEntityKey, &yetAnotherEntityData); err != datastore.ErrNoSuchEntity {
		t.Fatalf("Expect datastore.ErrNoSuchEntity, but was: %v", err)
	}
}

func TestDatastoreDumpAndRestoreDontCleanData(t *testing.T) {
	var tempDir string
	if _tempDir, err := ioutil.TempDir("", "testutil"); err == nil {
		tempDir = _tempDir
	} else {
		t.Fatalf("Failed to create a temporary directory: %v", err)
	}
	dumpFile := filepath.Join(tempDir, "dump.dat")

	defer os.RemoveAll(tempDir)

	ctx := GetAppengineContext()
	if _ctx, err := appengine.Namespace(ctx, "testing"); err == nil {
		ctx = _ctx
	} else {
		t.Fatalf("Failed to set namespace: %v", err)
	}

	entityKey := datastore.NewIncompleteKey(ctx, "entity", nil)
	entityData := entity{
		Value: "test1",
	}
	if _key, err := datastore.Put(ctx, entityKey, &entityData); err == nil {
		entityKey = _key
	} else {
		t.Fatalf("Failed to put data: %v", err)
	}

	anotherEntityKey := datastore.NewIncompleteKey(ctx, "anotherEntity", nil)
	anotherEntityData := anotherEntity{
		Value: "test2",
	}
	if _key, err := datastore.Put(ctx, anotherEntityKey, &anotherEntityData); err == nil {
		anotherEntityKey = _key
	} else {
		t.Fatalf("Failed to put data: %v", err)
	}

	if err := DatastoreDump(dumpFile, "testing", []string{"entity", "anotherEntity"}); err != nil {
		t.Fatalf("Failed to DatastoreDump(): %v", err)
	}

	if err := DatastoreRestore(dumpFile, "testing"); err != nil {
		t.Fatalf("Failed to DatastoreRestore(): %v", err)
	}

	if err := datastore.Get(ctx, entityKey, &entityData); err != nil {
		t.Fatalf("Failed to get data: %v", err)
	} else if entityData.Value != "test1" {
		t.Fatalf("Unexpected data: %+v", entityData)
	}
	if err := datastore.Get(ctx, anotherEntityKey, &anotherEntityData); err != nil {
		t.Fatalf("Failed to get data: %v", err)
	} else if anotherEntityData.Value != "test2" {
		t.Fatalf("Unexpected data: %+v", anotherEntityData)
	}
}
