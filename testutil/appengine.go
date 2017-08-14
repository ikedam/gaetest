package testutil

// aetest の高速化をはかります。

import (
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
