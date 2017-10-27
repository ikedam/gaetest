package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"strconv"
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

func callHandlerEntityPut(t *testing.T, inst aetest.Instance, id int64, reqdata interface{}) (*httptest.ResponseRecorder, error) {
	var data []byte
	var err error
	if data, err = json.Marshal(reqdata); err != nil {
		t.Fatalf("Failt to request PUT /entity/%d: %v", id, err)
	}

	req, err := inst.NewRequest("PUT", fmt.Sprintf("/entity/%d", id), bytes.NewReader(data))
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")

	e := echo.New()
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	c.SetParamNames("id")
	c.SetParamValues(strconv.FormatInt(id, 10))

	return res, handlerEntityPut(c)
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
	var id2 int64
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
		resdata := res.Body.Bytes()
		var result Entity
		if err := json.Unmarshal(resdata, &result); err != nil {
			t.Errorf("Failed to parse: %v", resdata)
		} else {
			id2 = result.ID
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

	// データの更新
	if res, err := callHandlerEntityPut(t, inst, id2, &struct {
		Name string `json:"name"`
	}{
		Name: "Testdata2.1",
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
			if result.ID != id2 {
				t.Errorf("Expect %v, but was %v", id2, result.ID)
			}
			if result.Name != "Testdata2.1" {
				t.Errorf("Expect Testdata2.1, but was %v", result.Name)
			}
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
				if result[0].Name != "Testdata2.1" {
					t.Errorf("Expect Testdata2.1, but was %v", result[0].Name)
				}
				if result[1].Name != "Testdata1" {
					t.Errorf("Expect Testdata1, but was %v", result[1].Name)
				}
			}
		}
	}
}

func TestManyEntity(t *testing.T) {
	inst := testutil.GetAppengineInstance()
	ctx := testutil.GetAppengineContextFor(inst)

	// Entity を空にする
	if keyList, err := datastore.NewQuery("Entity").KeysOnly().GetAll(ctx, nil); err != nil {
		panic(err)
	} else {
		if err := datastore.DeleteMulti(ctx, keyList); err != nil {
			panic(err)
		}
	}

	// 100 件のデータを昇順に登録
	keys := []*datastore.Key{}
	entities := []Entity{}
	for i := 0; i < 100; i++ {
		key := datastore.NewIncompleteKey(ctx, "Entity", nil)
		entity := Entity{
			Name:      fmt.Sprintf("test%d", i),
			CreatedAt: time.Now().AddDate(-1, 0, 0).Add(time.Duration(i) * time.Second),
		}
		keys = append(keys, key)
		entities = append(entities, entity)
	}
	if _, err := datastore.PutMulti(ctx, keys, entities); err != nil {
		panic(err)
	}

	testutil.FlushGoonCache(ctx)

	// 降順にデータが返る
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
			if len(result) != 100 {
				t.Errorf("Expect 100 records, but was %v", result)
			} else {
				for idx, r := range result {
					expect := fmt.Sprintf("test%d", (99 - idx))
					if r.Name != expect {
						t.Errorf("Expect %v, but was %v", expect, r.Name)
						break
					}
				}
			}
		}
	}

	// Entity を空にする
	if keyList, err := datastore.NewQuery("Entity").KeysOnly().GetAll(ctx, nil); err != nil {
		panic(err)
	} else {
		if err := datastore.DeleteMulti(ctx, keyList); err != nil {
			panic(err)
		}
	}

	// 100 件のデータを降順に登録
	keys = []*datastore.Key{}
	entities = []Entity{}
	for i := 0; i < 100; i++ {
		key := datastore.NewIncompleteKey(ctx, "Entity", nil)
		entity := Entity{
			Name:      fmt.Sprintf("test%d", i),
			CreatedAt: time.Now().Add(-time.Duration(i) * time.Second),
		}
		keys = append(keys, key)
		entities = append(entities, entity)
	}
	if _, err := datastore.PutMulti(ctx, keys, entities); err != nil {
		panic(err)
	}

	testutil.FlushGoonCache(ctx)

	// 降順にデータが返る
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
			if len(result) != 100 {
				t.Errorf("Expect 100 records, but was %v", result)
			} else {
				for idx, r := range result {
					expect := fmt.Sprintf("test%d", idx)
					if r.Name != expect {
						t.Errorf("Expect %v, but was %v", expect, r.Name)
						break
					}
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

func TestEntityPutBadParameters(t *testing.T) {
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

	// データの投入
	var id int64
	if res, err := callHandlerEntityPost(t, inst, &struct {
		Name string `json:"name"`
	}{
		Name: "Testdata1",
	}); err != nil {
		t.Fatalf("Expected no error but %v", err)
	} else if res.Code != http.StatusOK {
		t.Fatalf("Expected 200, but %v", res.Code)
	} else {
		resdata := res.Body.Bytes()
		var result Entity
		if err := json.Unmarshal(resdata, &result); err != nil {
			t.Fatalf("Failed to parse: %v", resdata)
		}
		id = result.ID
	}

	if res, err := callHandlerEntityListGet(t, inst); err != nil {
		t.Fatalf("Expected no error but %v", err)
	} else if res.Code != http.StatusOK {
		t.Fatalf("Expected 200, but %v", res.Code)
	} else {
		resdata := res.Body.Bytes()
		var result []Entity
		if err := json.Unmarshal(resdata, &result); err != nil {
			t.Fatalf("Failed to parse: %v", resdata)
		}
		if len(result) != 1 {
			t.Fatalf("Expect 1 records, but was %v", result)
		}
		if result[0].Name != "Testdata1" {
			t.Errorf("Expect Testdata1, but was %v", result[0].Name)
		}
	}

	// データの更新
	if res, err := callHandlerEntityPut(t, inst, id, &struct {
		ID        int64     `json:"id"`
		Name      string    `json:"name"`
		CreatedAt time.Time `json:"createdAt"`
	}{
		ID:        id + 1,
		Name:      "Testdata1.1",
		CreatedAt: time.Now().AddDate(1, 0, 0),
	}); err != nil {
		t.Fatalf("Expected no error but %v", err)
	} else if res.Code != http.StatusOK {
		t.Fatalf("Expected 200, but %v", res.Code)
	} else {
		resdata := res.Body.Bytes()
		var result Entity
		if err := json.Unmarshal(resdata, &result); err != nil {
			t.Errorf("Failed to parse: %v", resdata)
		} else {
			if result.Name != "Testdata1.1" {
				t.Errorf("Expect Testdata1.1, but was %v", result.Name)
			}
		}
	}

	if res, err := callHandlerEntityListGet(t, inst); err != nil {
		t.Fatalf("Expected no error but %v", err)
	} else if res.Code != http.StatusOK {
		t.Fatalf("Expected 200, but %v", res.Code)
	} else {
		resdata := res.Body.Bytes()
		var result []Entity
		if err := json.Unmarshal(resdata, &result); err != nil {
			t.Fatalf("Failed to parse: %v", resdata)
		}
		if len(result) != 1 {
			t.Fatalf("Expect 1 records, but was %v", result)
		}
		if result[0].Name != "Testdata1.1" {
			t.Errorf("Expect Testdata1.1, but was %v", result[0].Name)
		}
	}
}
