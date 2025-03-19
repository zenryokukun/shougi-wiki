package main

import (
	"fmt"

	"github.com/zenryokukun/gotweet"
)

// /api/insert-work成功時にTwitterに投稿するAPI
func TweetInsertWork(tw gotweet.Twitter, bd *WorkBody, thumb string) error {
	// 手数を計算
	l, err := bd.steps()
	if err != nil {
		err = stack("TweetInsertWork", err)
		return err
	}
	tesu := l - 1
	msg := ""
	msg += fmt.Sprintf("%v手詰作品が投稿されました。【%v】様、ありがとうございます！\n\n", tesu, bd.Author)
	msg += "【タイトル】" + bd.Title + "\n"
	msg += "【投稿日】" + currentDateStr() + "\n"

	// tweet
	tw.Tweet(msg, thumb)

	return err
}
