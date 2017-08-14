// Package testutil はテストで使用するユーティリティ群です。
package testutil

// Setup はテスト共通の前準備を行います。
// TestMain で呼び出してください。
func Setup() {
	setupAppengine()
}

// Teardown はテスト共通の後処理を行います。
// TestMain で呼び出してください。
// defer で呼び出す場合、 os.Exit と同じメソッド内だと
// ロックすることがあるため、TestMain を下請けメソッドなどに分離して
// 呼び出してください。
func Teardown() {
	teardownAppengine()
}
