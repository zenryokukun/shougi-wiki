// html/layout.htmlの.Content,.Meta,.Currentに対応したデータを保持するための機能群

package main

import (
	"html/template"
	"os"
	"path/filepath"
)

type Section struct {
	Heading string
	// sidebarのテンプレ内でクエリパラメタとして使う
	Tesu int
	// liタグ内がただのテキストの場合、aタグが入る場合があるため、
	// template.HTML型にしている。string型も設定できる（内部的には同じ）
	// でも、逆にここをstring型にしてしまうと、template.HTML型が設定できなくなる。
	List []template.HTML
}

type Record struct {
	Meta     template.HTML
	Content  template.HTML
	Current  string
	Sections []Section
	Err      error
}

type RootData map[string]Record

func NewRootData() RootData {
	rd := RootData{}
	// "/" -> home
	meta, content, err := layoutData("home")
	rd["home"] = Record{
		Meta:    meta,
		Content: content,
		Current: "HOME",
		Err:     err,
	}
	// "/rule/" -> rule
	meta, content, err = layoutData("rule")
	rd["rule"] = Record{
		Meta:    meta,
		Content: content,
		Current: "RULE",
		Err:     err,
	}

	return rd
}

// [internal] layoutDataで利用
func readFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	return string(b), err
}

func layoutData(route string) (template.HTML, template.HTML, error) {
	fol := filepath.Join("./html", route)
	contentPath := filepath.Join(fol, "content.html")
	metaPath := filepath.Join(fol, "meta.html")
	content, err := readFile(contentPath)
	if err != nil {
		err = stack("layoutData", err)
		return "", "", err
	}
	meta, err := readFile(metaPath)
	if err != nil {
		err = stack("layoutData", err)
		return "", "", err
	}
	return template.HTML(meta), template.HTML(content), err
}

// NewRootDataはRecord型のmapで、キャッシュとして利用する。
// NewRootRecordはRecord型で都度ファイルを読み取って返す関数。
// content.html、とmeta.htmlが/html/*route*に格納されている必要あり
func NewRootRecord(route string) Record {
	meta, content, err := layoutData(route)
	if err != nil {
		err = stack("NewRootRecord", err)
	}
	// .Currentは呼び出し下で対応すること
	rec := Record{
		Meta:    meta,
		Content: content,
		Err:     err,
	}

	return rec
}
