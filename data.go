// html/layout.htmlの.Content,.Meta,.Currentに対応したデータを保持するための機能群

package main

import (
	"html/template"
	"os"
	"path/filepath"
)

type Section struct {
	Heading string
	IsLink  bool
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
	// "/edit/" -> edit
	meta, content, err = layoutData("edit")
	rd["edit"] = Record{
		Meta:    meta,
		Content: content,
		Current: "EDIT",
		Sections: []Section{
			{
				Heading: "操作手順",
				IsLink:  false,
				List: []template.HTML{
					"盤上に駒を配置してください",
					"盤上の駒をクリックし、別のマスをクリックすれば動かせます",
					"盤上の駒を消すには、駒をクリックし、ゴミ箱アイコンをクリックしてください。キーボードのdeleteキーでもOKです",
					"完成したら、確定アイコンをクリックしてください",
				},
			},
			{
				Heading: "その他仕様",
				IsLink:  false,
				List: []template.HTML{
					"既に駒が置いてあるマスをクリックした場合、上書きされます",
					"解説の作成は、確定アイコンをクリック後、別ページにて行います",
					"簡易的なチェックしか実装していません。確定前にチェックをお願いします",
				},
			},
		},
		Err: err,
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
		return "", "", err
	}
	meta, err := readFile(metaPath)
	if err != nil {
		return "", "", err
	}
	return template.HTML(meta), template.HTML(content), err
}

// NewRootDataはRecord型のmapで、キャッシュとして利用する。
// NewRootRecordはRecord型で都度ファイルを読み取って返す関数。
// content.html、とmeta.htmlが/html/*route*に格納されている必要あり
func NewRootRecord(route string) Record {
	meta, content, err := layoutData(route)
	// .Currentは呼び出し下で対応すること
	return Record{
		Meta:    meta,
		Content: content,
		Err:     err,
	}
}
