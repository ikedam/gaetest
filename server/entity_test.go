package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ikedam/gaetest/testutil"
	"github.com/labstack/echo"

	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
)

func callHandlerEntityListGet(t *testing.T, inst aetest.Instance) (*httptest.ResponseRecorder, error) {
	req, err := inst.NewRequest("GET", "/entity/", nil)
	if err != nil {
		panic(err)
	}

	e := echo.New()
	res := httptest.NewRecorder()

	return res, handlerEntityListGet(e.NewContext(req, res))
}

func callHandlerEntityPost(t *testing.T, inst aetest.Instance, reqdata interface{}) (*httptest.ResponseRecorder, error) {
	var data []byte
	var err error
	if data, err = json.Marshal(reqdata); err != nil {
		t.Fatalf("Failt to request POST /entity/: %v", err)
	}
	return callHandlerEntityPostRaw(t, inst, data)
}

func callHandlerEntityPostRaw(t *testing.T, inst aetest.Instance, data []byte) (*httptest.ResponseRecorder, error) {
	req, err := inst.NewRequest("POST", "/entity/", bytes.NewReader(data))
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")

	e := echo.New()
	res := httptest.NewRecorder()

	return res, handlerEntityPost(e.NewContext(req, res))
}

func TestEntity(t *testing.T) {
	// Entity を空にする
	inst := testutil.GetAppengineInstance()
	ctx := testutil.GetAppengineContextFor(inst)

	if keyList, err := datastore.NewQuery("Entity").KeysOnly().GetAll(ctx, nil); err != nil {
		panic(err)
	} else {
		if err := datastore.DeleteMulti(ctx, keyList); err != nil {
			panic(err)
		}
	}
	testutil.FlushGoonCache(ctx)

	// 最初はリストに何も返らない
	if res, err := callHandlerEntityListGet(t, inst); err != nil {
		t.Fatalf("Expected no error but %v", err)
	} else {
		if res.Code != http.StatusOK {
			t.Errorf("Expected 200, but %v", res.Code)
		}
		resdata := res.Body.Bytes()
		var result []Entity
		if err := json.Unmarshal(resdata, &result); err != nil {
			t.Errorf("Failed to parse: %v", resdata)
		} else {
			if len(result) != 0 {
				t.Errorf("Expect empty, but was %v", result)
			}
		}
	}

	// データの投入
	if res, err := callHandlerEntityPost(t, inst, &struct {
		Name string `json:"name"`
	}{
		Name: "Testdata1",
	}); err != nil {
		t.Fatalf("Expected no error but %v", err)
	} else {
		if res.Code != http.StatusOK {
			t.Errorf("Expected 200, but %v", res.Code)
		}
		resdata := res.Body.Bytes()
		var result Entity
		if err := json.Unmarshal(resdata, &result); err != nil {
			t.Errorf("Failed to parse: %v", resdata)
		} else {
			if result.Name != "Testdata1" {
				t.Errorf("Expect Testdata1, but was %v", result.Name)
			}
			if result.ID == 0 {
				t.Errorf("Expect non-0, but was %v", result.ID)
			}
			if math.Abs(time.Now().Sub(result.CreatedAt).Seconds()) > 5.0 {
				t.Errorf(
					"Expect time is set a proper value, but was %v (now %v)",
					result.CreatedAt,
					time.Now(),
				)
			}
		}
	}
	// 投入したデータが得られる
	if res, err := callHandlerEntityListGet(t, inst); err != nil {
		t.Fatalf("Expected no error but %v", err)
	} else {
		if res.Code != http.StatusOK {
			t.Errorf("Expected 200, but %v", res.Code)
		}
		resdata := res.Body.Bytes()
		var result []Entity
		if err := json.Unmarshal(resdata, &result); err != nil {
			t.Errorf("Failed to parse: %v", resdata)
		} else {
			if len(result) != 1 {
				t.Errorf("Expect 1 entry, but was %v", result)
			} else {
				if result[0].Name != "Testdata1" {
					t.Errorf("Expect Testdata1, but was %v", result[0].Name)
				}
			}
		}
	}

	// データの投入
	if res, err := callHandlerEntityPost(t, inst, &struct {
		Name string `json:"name"`
	}{
		Name: "Testdata2",
	}); err != nil {
		t.Fatalf("Expected no error but %v", err)
	} else {
		if res.Code != http.StatusOK {
			t.Errorf("Expected 200, but %v", res.Code)
		}
	}

	// 投入の逆順にデータが返る
	if res, err := callHandlerEntityListGet(t, inst); err != nil {
		t.Fatalf("Expected no error but %v", err)
	} else {
		if res.Code != http.StatusOK {
			t.Errorf("Expected 200, but %v", res.Code)
		}
		resdata := res.Body.Bytes()
		var result []Entity
		if err := json.Unmarshal(resdata, &result); err != nil {
			t.Errorf("Failed to parse: %v", resdata)
		} else {
			if len(result) != 2 {
				t.Errorf("Expect 2 records, but was %v", result)
			} else {
				if result[0].Name != "Testdata2" {
					t.Errorf("Expect Testdata2, but was %v", result[0].Name)
				}
				if result[1].Name != "Testdata1" {
					t.Errorf("Expect Testdata1, but was %v", result[1].Name)
				}
			}
		}
	}
}

func TestEntityHandlerEntityListGetDatastoreError(t *testing.T) {
	inst := testutil.GetAppengineInstance()
	mocker := testutil.NewAppengineMock()
	mocked := mocker.MockInstance(inst)
	if mocked == nil {
		t.Skip("MockInstance is not supported")
	}
	mocker.AddAPICallMock(testutil.AppengineAPICallMock{
		Service: "datastore",
		Method:  "RunQuery",
		Error:   errors.New("Expected error"),
	})

	if res, err := callHandlerEntityListGet(t, mocked); err != nil {
		t.Fatalf("Expected no error but %v", err)
	} else {
		if res.Code != http.StatusInternalServerError {
			t.Errorf("Expected 500, but %v", res.Code)
		}
		errorLogList := mocker.GetLogsEqualTo(testutil.LogLevelError)
		if len(errorLogList) != 0 {
			if errorLogList[len(errorLogList)-1] != "Failed to query Entity: Expected error" {
				t.Errorf("Unexpected error message: %v", errorLogList)
			}
		}
	}
}

func TestEntityHandlerEntityPostDatastorePutError(t *testing.T) {
	inst := testutil.GetAppengineInstance()
	mocker := testutil.NewAppengineMock()
	mocked := mocker.MockInstance(inst)
	if mocked == nil {
		t.Skip("MockInstance is not supported")
	}
	mocker.AddAPICallMock(testutil.AppengineAPICallMock{
		Service: "datastore",
		Method:  "Put",
		Error:   errors.New("Expected error"),
	})

	if res, err := callHandlerEntityPost(t, mocked, &struct {
		Name string `json:"name"`
	}{
		Name: "Testdata1",
	}); err != nil {
		t.Fatalf("Expected no error but %v", err)
	} else {
		if res.Code != http.StatusInternalServerError {
			t.Errorf("Expected 500, but %v", res.Code)
		}
		errorLogList := mocker.GetLogsEqualTo(testutil.LogLevelError)
		if len(errorLogList) != 0 {
			if errorLogList[len(errorLogList)-1] != "Failed to put Entity: Expected error" {
				t.Errorf("Unexpected error message: %v", errorLogList)
			}
		}
	}
}

func TestEntityHandlerEntityPostDatastoreGetError(t *testing.T) {
	inst := testutil.GetAppengineInstance()
	mocker := testutil.NewAppengineMock()
	mocked := mocker.MockInstance(inst)
	if mocked == nil {
		t.Skip("MockInstance is not supported")
	}
	mocker.AddAPICallMock(testutil.AppengineAPICallMock{
		Service: "datastore",
		Method:  "Get",
		Error:   errors.New("Expected error"),
	})

	if res, err := callHandlerEntityPost(t, mocked, &struct {
		Name string `json:"name"`
	}{
		Name: "Testdata1",
	}); err != nil {
		t.Fatalf("Expected no error but %v", err)
	} else {
		if res.Code != http.StatusInternalServerError {
			t.Errorf("Expected 500, but %v", res.Code)
		}
		errorLogList := mocker.GetLogsEqualTo(testutil.LogLevelError)
		if len(errorLogList) != 0 {
			if !strings.HasPrefix(errorLogList[len(errorLogList)-1], "Failed to re-get Entity: Expected error, key=") {
				t.Errorf("Unexpected error message: %v", errorLogList)
			}
		}
	}
}

func TestEntityHandlerEntityPostMemcacheError(t *testing.T) {
	inst := testutil.GetAppengineInstance()
	mocker := testutil.NewAppengineMock()
	mocked := mocker.MockInstance(inst)
	if mocked == nil {
		t.Skip("MockInstance is not supported")
	}
	mocker.AddAPICallMock(testutil.AppengineAPICallMock{
		Service: "memcache",
		Error:   errors.New("Expected error"),
	})

	if res, err := callHandlerEntityPost(t, mocked, &struct {
		Name string `json:"name"`
	}{
		Name: "Testdata1",
	}); err != nil {
		t.Fatalf("Expected no error but %v", err)
	} else {
		if res.Code != http.StatusOK {
			t.Errorf("Expected 200, but %v", res.Code)
		}
		errorLogList := mocker.GetLogsEqualTo(testutil.LogLevelError)
		if len(errorLogList) != 0 {
			t.Errorf("Expected no error messages, but was: %v", errorLogList)
		}
	}
}


func TestEntityHandlerEntityPostBadRequestError(t *testing.T) {
	inst := testutil.GetAppengineInstance()

	if res, err := callHandlerEntityPostRaw(t, inst, []byte("xxxx")); err != nil {
		t.Fatalf("Expected no error but %v", err)
	} else {
		if res.Code != http.StatusBadRequest {
			t.Errorf("Expected 400, but %v", res.Code)
		}
	}
}