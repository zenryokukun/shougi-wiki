package main

import (
	"encoding/json"
	"fmt"
)

type (
	EditFormData struct {
		// 編集モードのみForm(hidden)に存在する項目
		// 画面に表示している作品IDで使う
		WorkId      string
		Explanation string
		Author      string
		Editor      string
		Title       string
	}
	EditCanvasData struct {
		Main       [][]int32          `json:"main"`
		Tegoma     []map[string]int32 `json:"tegoma"`
		GoteTegoma []map[string]int32 `json:"goteTegoma"`
		Kihu       []string           `json:"kihu"`
	}

	Dates struct {
		PublishDate string
		EditDate    string
	}

	WorksContent struct {
		EditFormData
		Dates
		// canvasのdata-*属性に埋め込むので、JSON形式でOK
		Main       string
		Tegoma     string
		GoteTegoma string
		// 描写用の棋譜。templateで直接展開するのでJSON変換不要
		Kihu []string
		// canvasに埋め込む用の棋譜。JSON変換要
		KihuJ string
		// 初回投稿日
		PublishDate string
		// 最終更新日
		EditDate  string
		IsPreview bool
		Tesu      int
	}

	WorksMeta struct {
		// metaタグ内のtitleタグ
		Title string
		// metaタグのdescription
		Description string
	}
)

func (wc *WorksContent) parse(data string) {
	var cdata = &EditCanvasData{}
	json.Unmarshal([]byte(data), cdata)
	bmain, err := json.Marshal(cdata.Main)
	if err != nil {
		fmt.Println(err)
		return
	}
	btegoma, err := json.Marshal(cdata.Tegoma)
	if err != nil {
		fmt.Println(err)
		return
	}
	bgoteTegoma, err := json.Marshal(cdata.GoteTegoma)
	if err != nil {
		fmt.Println(err)
		return
	}
	bKihuJ, err := json.Marshal(cdata.Kihu)
	if err != nil {
		fmt.Println(err)
		return
	}

	wc.Main = string(bmain)
	wc.Tegoma = string(btegoma)
	wc.GoteTegoma = string(bgoteTegoma)
	// 初期配置は棋譜がないため、空文字が設定されている。除外。
	wc.Kihu = cdata.Kihu[1:]
	wc.KihuJ = string(bKihuJ)

	wc.Tesu = len(cdata.Main) - 1
}

func (wc *WorksContent) addFormData(form EditFormData) {
	wc.Author = form.Author
	wc.Explanation = form.Explanation
	wc.Title = form.Title
	wc.Editor = form.Editor
	wc.WorkId = form.WorkId
}

func (wc *WorksContent) addDates(dates Dates) {
	wc.PublishDate = dates.PublishDate
	wc.EditDate = dates.EditDate
}

func preview(data string, form EditFormData, dates Dates) *WorksContent {
	wc := &WorksContent{
		IsPreview: true,
	}
	wc.parse(data)
	wc.addFormData(form)
	wc.addDates(dates)
	return wc
}
