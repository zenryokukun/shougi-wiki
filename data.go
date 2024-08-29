// html/layout.htmlの.Content,.Meta,.Currentに対応したデータを保持するための機能群

package main

import (
	"html/template"
	"log"
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
}

type RootData map[string]Record

func NewRootData() RootData {
	rd := RootData{}
	// "/" -> home
	meta, content := layoutData("home")
	rd["home"] = Record{
		Meta:    meta,
		Content: content,
		Current: "HOME",
	}
	// "/rule/" -> rule
	meta, content = layoutData("rule")
	rd["rule"] = Record{
		Meta:    meta,
		Content: content,
		Current: "RULE",
	}
	// "/edit/" -> edit
	meta, content = layoutData("edit")
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
	}
	return rd
}

// [internal] layoutDataで利用
func readFile(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	return string(b)
}

func layoutData(route string) (template.HTML, template.HTML) {
	fol := filepath.Join("./html", route)
	contentPath := filepath.Join(fol, "content.html")
	metaPath := filepath.Join(fol, "meta.html")
	content := readFile(contentPath)
	meta := readFile(metaPath)
	return template.HTML(meta), template.HTML(content)
}
