package testutil

import (
	"errors"
	"reflect"
	"testing"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/memcache"
)

type entity struct {
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
		Error: expectError,
	})

	e := entity{
		Value: "test",
	}

	key := datastore.NewIncompleteKey(mocked, "entity", nil)

	if _, err := datastore.Put(mocked, key, &e); err != expectError {
		t.Fatalf("Expect %v, but was %v", expectError, err)
	}

	item := memcache.Item{
		Key:"key",
		Value: []byte("bar"),
	}

	if err := memcache.Set(mocked, &item); err != nil {
		t.Fatalf("Expect success, but was %v", err)
	}

	mocker.AddAPICallMock(AppengineAPICallMock{
		Service: "memcache",
		Error: expectError,
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
		Method: "Get",
		Error: expectError,
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
