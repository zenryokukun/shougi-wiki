package main

import (
	"fmt"
	"net/http"
)

func stack(msg string, err error) error {
	return fmt.Errorf("%v\n%w", msg, err)
}

func logErr(r *http.Request, err error) {
	if r == nil {
		fmt.Println(err)
		return
	}
	// エラーが生じたパス
	url := r.URL.Path
	// 遷移元の画面
	ref := r.Header.Get("Referer")
	// ブラウザなど？
	agt := r.Header.Get("User-Agent")

	// 現在時刻 秒まで
	now := currentDateStr()
	log := fmt.Sprintf("Date:%v\nURL:%v\nReferer:%v\n,User-Agent:%v\n[Error]%v", now, url, ref, agt, err)
	fmt.Println(log)
}
