package log

import "log"

// FatalError はエラーチェックをしてエラーが発生していたら強制終了します。
func FatalError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
