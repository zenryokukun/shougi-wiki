package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"slices"
	"time"
)

const LONG = -1

// *sql.DBや*sql.TxでもQueryRowが使えるように、、、
type Transer interface {
	QueryRow(query string, args ...any) *sql.Row
}

// dbから直接抜いたデータ。Nullを含むカラムがエラーになるので、、、
type WorkRawRecord struct {
	Id              int
	Tesu            int
	Title           string
	Explanation     string
	Author          string
	Editor          sql.NullString
	Main            string
	Tegoma          string
	GoteTegoma      string
	KihuJ           string
	PublishDateUnix sql.NullInt64
	EditDateUnix    sql.NullInt64
	Good            sql.NullInt64
	Bad             sql.NullInt64
	Demand          sql.NullInt64
}

type PostRecord struct {
	Id       int
	Seq      int
	Name     string
	Comment  string
	Type     string
	PostDate int
	Good     int
	Bad      int
	// dbの項目ではないけど
	PostDateStr string
}

type WorkRecord struct {
	Id              int
	Tesu            int
	Title           string
	Explanation     string
	Author          string
	Editor          string
	Main            string
	Tegoma          string
	GoteTegoma      string
	KihuJ           string
	PublishDateUnix int
	EditDateUnix    int
	Good            int
	Bad             int
	Demand          int
	// parseして生成
	Kihu        []string
	PublishDate string
	EditDate    string
	// workTemplateで必要
	IsPreview bool
	// 投稿（post）。ない場合もある。
	Posts []PostRecord
}

func (wr *WorkRawRecord) parse() WorkRecord {
	rec := WorkRecord{}
	rec.Id = wr.Id
	rec.Tesu = wr.Tesu
	rec.Title = wr.Title
	rec.Explanation = wr.Explanation
	rec.Author = wr.Author
	rec.Main = wr.Main
	rec.Tegoma = wr.Tegoma
	rec.GoteTegoma = wr.GoteTegoma
	rec.KihuJ = wr.KihuJ

	if wr.Editor.Valid {
		rec.Editor = wr.Editor.String
	} else {
		rec.Editor = ""
	}

	if wr.PublishDateUnix.Valid {
		rec.PublishDateUnix = int(wr.PublishDateUnix.Int64)
	} else {
		rec.PublishDateUnix = 0
	}
	if wr.EditDateUnix.Valid {
		rec.EditDateUnix = int(wr.EditDateUnix.Int64)
	} else {
		rec.EditDateUnix = 0
	}
	if wr.Good.Valid {
		rec.Good = int(wr.Good.Int64)
	} else {
		rec.Good = 0
	}
	if wr.Bad.Valid {
		rec.Bad = int(wr.Bad.Int64)
	} else {
		rec.Bad = 0
	}
	if wr.Demand.Valid {
		rec.Demand = int(wr.Demand.Int64)
	} else {
		rec.Demand = 0
	}

	json.Unmarshal([]byte(wr.KihuJ), &rec.Kihu)
	// 棋譜の第一要素は必ず空白になるので、除外
	rec.Kihu = rec.Kihu[1:]

	if rec.PublishDateUnix == 0 {
		rec.PublishDate = "-"
	} else {
		rec.PublishDate = unixToStr(int64(rec.PublishDateUnix))
	}
	if rec.EditDateUnix == 0 {
		rec.EditDate = "-"
	} else {
		rec.EditDate = unixToStr(int64(rec.EditDateUnix))
	}

	rec.IsPreview = false

	return rec
}

type WorkLink struct {
	Id    int
	Title string
}

type WorksMap map[int][]WorkLink

type WorksCache struct {
	data         WorksMap
	SectionCache []Section
}

func (wc *WorksCache) Update(db *sql.DB) {
	// 初期化しないとエラーになる。更新の都度初期化する。そうしないとUpdateで二重に設定される。
	wc.data = WorksMap{}
	wc.data.update(db)
	wc.SectionCache = wc.data.section()
}

func (c WorksMap) update(db *sql.DB) {
	// @Todo 全量取得しているんどえ、直近N個に絞るような仕組みを導入
	rows, err := db.Query(`
		SELECT ID,TESU,TITLE FROM WORKS
		ORDER BY TESU ASC,ID ASC;
	`)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var tesu int
		var title string
		rows.Scan(&id, &tesu, &title)
		info := WorkLink{id, title}
		if tesu > 11 {
			c[LONG] = append(c[LONG], info)
		} else {
			c[tesu] = append(c[tesu], info)
		}
	}
}

func (c WorksMap) section() []Section {
	ret := []Section{}
	keys := c.Sort()
	for _, key := range keys {
		sec := Section{}
		sec.Heading = fmt.Sprintf("%v手詰集", key)
		sec.IsLink = true
		wk := c[key]
		for _, v := range wk {
			link := fmt.Sprintf(`<a href="/works?id=%v">%v</a>`, v.Id, v.Title)
			sec.List = append(sec.List, template.HTML(link))
		}
		ret = append(ret, sec)
	}
	return ret
}

func (c WorksMap) Sort() []int {
	var ret []int
	for k := range c {
		ret = append(ret, k)
	}
	slices.Sort(ret)
	return ret
}

// 投稿作品（Work）を挿入する
func insertWork(db *sql.DB, bd *InsertWorkBody) (int, error) {
	var maxId int
	l, err := bd.steps()
	if err != nil {
		return maxId, err
	}
	// 配列の長さ-1が手数になる（1手詰みなら配列は2つあるよね。）
	tesu := l - 1
	if tesu <= 0 {
		err := errors.New("Main配列が不正です。手数が0以下になってしまいます。")
		return maxId, err
	}

	// 現在時刻（UNIX）
	pdate := currentDateUnix()

	tx, err := db.Begin()
	if err != nil {
		return maxId, err
	}

	_, err = tx.Exec(`
		INSERT INTO WORKS
		(
			TESU,
			TITLE,
			KIHU,
			EXPLANATION,
			AUTHOR,
			EDITOR,
			MAIN,TEGOMA,GOTETEGOMA,
			PUBLISH_DATE,EDIT_DATE,
			GOOD,BAD,DEMAND
		)
		VALUES (
			?,?,?,?,?,
			?,?,?,?,?,
			?,?,?,?
		);
	`,
		tesu,
		bd.Title,
		bd.Kihu,
		bd.Explanation,
		bd.Author,
		nil,
		bd.Main, bd.Tegoma, bd.GoteTegoma,
		pdate, nil,
		0, 0, 0,
	)

	maxId = worksMaxId(tx)

	if err != nil {
		tx.Rollback()
		return maxId, err
	}

	tx.Commit()
	return maxId, nil
}

// 投稿作品（Work）の評価を更新する
func updateWorkEval(db *sql.DB, eb *UpdateEvalBody) (int, error) {
	key := eb.Key
	id := eb.Id

	var cnt int

	tx, err := db.Begin()
	if err != nil {
		return cnt, err
	}
	query := fmt.Sprintf("SELECT %v FROM WORKS WHERE ID=?", key)
	row := tx.QueryRow(query, id)
	err = row.Scan(&cnt)
	if err != nil {
		tx.Rollback()
		return cnt, err
	}

	cnt += eb.Value // eb.Valueは1 or -1

	query = fmt.Sprintf("UPDATE WORKS SET %v=%v WHERE ID=?", key, cnt)
	_, err = tx.Exec(query, id)
	if err != nil {
		tx.Rollback()
		return cnt, err
	}
	tx.Commit()
	return cnt, nil
}

// WORKテーブルからレコードを取得する
func getWork(db *sql.DB, id int) (WorkRecord, error) {
	// QueryRowはエラーは返さない。結果無の場合はrow.Scanでエラーになる仕様
	row := db.QueryRow(`
		SELECT 
			ID,TESU,TITLE,EXPLANATION,AUTHOR,EDITOR,
			MAIN,TEGOMA,GOTETEGOMA,KIHU,
			PUBLISH_DATE,EDIT_DATE,
			GOOD,BAD,DEMAND
		FROM WORKS WHERE ID=?
	`, id)

	var raw WorkRawRecord

	err := row.Scan(
		&raw.Id, &raw.Tesu, &raw.Title, &raw.Explanation, &raw.Author, &raw.Editor,
		&raw.Main, &raw.Tegoma, &raw.GoteTegoma, &raw.KihuJ,
		&raw.PublishDateUnix, &raw.EditDateUnix,
		&raw.Good, &raw.Bad, &raw.Demand,
	)

	if err != nil {
		fmt.Println(err)
		msg := fmt.Sprintf("workテーブルからデータを取れませんでした。ID:%v", id)
		err = errors.New(msg)
	}
	rec := raw.parse()
	return rec, err
}

// WORKSテーブルから最大IDを取得。サムネ保存場所に必要
func worksMaxId(db Transer) int {
	var m int
	r := db.QueryRow("SELECT MAX(ID) FROM WORKS")
	r.Scan(&m)
	return m
}

// コメントを投稿する
func insertPost(db *sql.DB, name, comment, commentType string, id int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	maxSeq := postMaxSeq(tx, id)
	nextSeq := maxSeq + 1
	_ = nextSeq
	pdate := time.Now().Unix()
	_, err = tx.Exec(`
		INSERT INTO POSTS (
			ID,SEQ,
			NAME,COMMENT,
			TYPE,
			POST_DATE,
			GOOD,BAD	
		)
		VALUES (
			?,?,?,?,?,?,0,0
		)
	`,
		id, nextSeq,
		name, comment,
		commentType,
		pdate,
	)
	if err != nil {
		err = tx.Rollback()
		if err != nil {
			fmt.Println(err)
		}
		return err
	}
	err = tx.Commit()
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

// POSTSテーブルからidに対応した最大のSEQカラムを取得
func postMaxSeq(db Transer, id int) int {
	var m int
	r := db.QueryRow(`SELECT COALESCE(MAX(SEQ),0) FROM POSTS WHERE ID=?`, id)
	r.Scan(&m)
	return m
}

// 投稿内容を取得
func getPosts(db *sql.DB, id, offset, limit int) ([]PostRecord, error) {
	recs := []PostRecord{}
	query := fmt.Sprintf(`
		SELECT * FROM POSTS 
		WHERE ID=? AND SEQ>? 
		LIMIT %v
		`, limit)

	rows, err := db.Query(query, id, offset)

	if err != nil {
		return recs, err
	}

	defer rows.Close()

	for rows.Next() {
		r := PostRecord{}
		rows.Scan(
			&r.Id, &r.Seq, &r.Name, &r.Comment, &r.Type,
			&r.PostDate, &r.Good, &r.Bad,
		)
		r.PostDateStr = unixToStr(int64(r.PostDate))
		recs = append(recs, r)
	}

	return recs, nil
}

// post内の評価を更新
func updatePostEval(db *sql.DB, eb *UpdatePostEvalBody) (int, error) {
	var cnt int
	tx, err := db.Begin()
	if err != nil {
		return cnt, err
	}

	query := fmt.Sprintf("SELECT %v FROM POSTS WHERE ID=? AND SEQ=?", eb.Key)

	row := tx.QueryRow(query, eb.Id, eb.Seq)
	err = row.Scan(&cnt)
	if err != nil {
		terr := tx.Rollback()
		if terr != nil {
			return cnt, terr
		}
		return cnt, err
	}

	cnt += eb.Value

	query = fmt.Sprintf("UPDATE POSTS SET %v=%v WHERE ID=? AND SEQ=?", eb.Key, cnt)
	_, err = tx.Exec(query, eb.Id, eb.Seq)
	if err != nil {
		terr := tx.Rollback()
		if terr != nil {
			return cnt, terr
		}
		return cnt, err
	}

	err = tx.Commit()
	if err != nil {
		return cnt, err
	}

	return cnt, nil
}

// valが正なら「次」の作品、0以下なら「前の作品」
func getNextWork(db *sql.DB, id, tesu, val int) (int, error) {
	var query string
	if val > 0 {
		query = `
		SELECT ID FROM WORKS 
		WHERE TESU=? AND ID>?
		ORDER BY ID ASC
		LIMIT 1
		`
	} else {
		query = `
		SELECT ID FROM WORKS 
		WHERE TESU=? AND ID<?
		ORDER BY ID DESC
		LIMIT 1
		`
	}

	var retId int
	row := db.QueryRow(query, tesu, id)
	err := row.Scan(&retId)
	return retId, err
}
