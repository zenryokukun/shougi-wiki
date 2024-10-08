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
	BackupSeq       sql.NullInt64
	Comment         sql.NullString
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

type HistorySummary struct {
	Id       int
	Seq      int
	Comment  string
	Editor   string
	EditDate string
	// DB値ではない。履歴から復元した時に挿入する"履歴レコード"ならtrue
	IsRestored bool
}

// IdはWorkRawRecordにある
type HistoryRawRecord struct {
	Seq        int
	BackupDate int
	WorkRawRecord
}

type HistoryRecord struct {
	Seq        int
	BackupDate int
	WorkRecord
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
	BackupSeq       int
	Comment         string
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
	if wr.BackupSeq.Valid {
		rec.BackupSeq = int(wr.BackupSeq.Int64)
	} else {
		rec.BackupSeq = 0
	}
	if wr.Comment.Valid {
		rec.Comment = wr.Comment.String
	} else {
		rec.Comment = ""
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
func insertWork(db *sql.DB, bd *WorkBody) (int, error) {
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
			GOOD,BAD,DEMAND,
			BACKUP_SEQ
		)
		VALUES (
			?,?,?,?,?,
			?,?,?,?,?,
			?,?,?,?,?
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
		0,
	)

	maxId = worksMaxId(tx)

	if err != nil {
		tx.Rollback()
		return maxId, err
	}

	tx.Commit()
	return maxId, nil
}

// 投稿速品（work）を更新する。編集画面からの登録時に利用
func updateWork(db *sql.DB, bd *WorkBody) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	// バックアップから復元されたWorkかをチェック
	isRestored, err := isRestoredWork(tx, bd.Id)
	if err != nil {
		return err
	}
	// var bkup sql.NullInt64
	// row := tx.QueryRow(`SELECT BACKUP_SEQ FROM WORKS WHERE ID=?`, bd.Id)
	// err = row.Scan(&bkup)
	// if err != nil {
	// 	txErr := tx.Rollback()
	// 	if txErr != nil {
	// 		fmt.Println(txErr)
	// 	}
	// 	return err
	// }
	// // バックアップから復元されている場合true
	// isRestoredWork := bkup.Valid

	// バックアップ日時
	now := currentDateUnix()

	// バックアップ復元されていない場合、バックアップ取得する
	if !isRestored {
		err = backupWork(tx, bd.Id, now)
		if err != nil {
			return err
		}
	}

	// workテーブルを更新する
	err = updateWorkFromEdit(tx, bd, now)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// updateWork関数内で利用
// workのバックアップを実行（historyにinsert）
// .Rollback,Commitは呼び出し元で実行すること
func backupWork(tx *sql.Tx, id int, now int64) error {
	// 最大連番を取得
	maxSeq, err := historyMaxSeq(tx, id)
	if err != nil {
		return err
	}

	// 最大連番+1を挿入する連番とする
	nextSeq := maxSeq + 1

	// 更新前の情報をバックアップ
	query := fmt.Sprintf(`
		INSERT INTO HISTORY
		(
			ID,SEQ,BACKUP_DATE,
			TESU,TITLE,KIHU,
			EXPLANATION,AUTHOR,EDITOR,
			MAIN,TEGOMA,GOTETEGOMA,
			PUBLISH_DATE,EDIT_DATE,
			GOOD,BAD,DEMAND,
			BACKUP_SEQ,COMMENT
		)
		SELECT 
			ID,%v,%v,
			TESU,TITLE,KIHU,
			EXPLANATION,AUTHOR,EDITOR,
			MAIN,TEGOMA,GOTETEGOMA,
			PUBLISH_DATE,EDIT_DATE,
			GOOD,BAD,DEMAND,
			BACKUP_SEQ,COMMENT
		FROM WORKS
		WHERE ID=?
	`, nextSeq, now)

	_, err = tx.Exec(query, id)
	if err != nil {
		return err
	}
	return nil
}

// updateWork関数内で利用
// 編集画面からWorkを更新。good,bad,demand等の評価は更新対象外
// tx.Commit,tx.Rollbackは呼び出し下で実行すること
func updateWorkFromEdit(tx *sql.Tx, bd *WorkBody, now int64) error {

	// 手数を計算
	l, err := bd.steps()
	if err != nil {
		return err
	}
	tesu := l - 1
	if tesu <= 0 {
		err := errors.New("Main配列が不正です。手数が0以下になってしまいます。")
		return err
	}

	// Worksを更新
	// 編集モードで更新される場合、BACKUP_SEQはnullにする必要がある
	// (null以外だと、historyから復元されたデータとして扱われ、historyに入らない)
	_, err = tx.Exec(`
		UPDATE WORKS SET
		  TESU=?
		  ,TITLE=?
		  ,KIHU=?
		  ,EXPLANATION=?
		  ,EDITOR=?
		  ,MAIN=?
		  ,TEGOMA=?
		  ,GOTETEGOMA=?
		  ,EDIT_DATE=?
		  ,BACKUP_SEQ=NULL
		  ,COMMENT=?
		WHERE ID=?
	`,
		tesu, bd.Title, bd.Kihu, bd.Explanation,
		bd.Editor, bd.Main, bd.Tegoma, bd.GoteTegoma,
		now, bd.Comment,
		bd.Id,
	)

	if err != nil {
		return err
	}

	return nil
}

// historyテーブルのレコードを全カラム分取得
func getHistory(db Transer, id, seq int) (HistoryRecord, error) {
	var rawSeq, backupDate int
	var raw WorkRawRecord
	var rec HistoryRecord
	row := db.QueryRow(`
		SELECT 
			ID,SEQ,BACKUP_DATE,
			TESU,TITLE,KIHU,EXPLANATION,
			AUTHOR,EDITOR,
			MAIN,TEGOMA,GOTETEGOMA,
			PUBLISH_DATE,EDIT_DATE,
			GOOD,BAD,DEMAND 
		FROM HISTORY
		WHERE ID=? AND SEQ=?
	`, id, seq)
	err := row.Scan(
		&raw.Id, &rawSeq, &backupDate,
		&raw.Tesu, &raw.Title, &raw.KihuJ, &raw.Explanation,
		&raw.Author, &raw.Editor,
		&raw.Main, &raw.Tegoma, &raw.GoteTegoma,
		&raw.PublishDateUnix, &raw.EditDateUnix,
		&raw.Good, &raw.Bad, &raw.Demand,
	)
	if err != nil {
		return rec, nil
	}
	work := raw.parse()
	rec = HistoryRecord{
		Seq:        rawSeq,
		BackupDate: backupDate,
		WorkRecord: work,
	}
	return rec, nil
}

// historyテーブルのサマリーを取得。編集画面のリンク表示用
func getHistorySummary(db *sql.DB, id int) ([]HistorySummary, error) {
	var summary []HistorySummary
	// rows, err := db.Query(`
	// 	SELECT ID,SEQ,EDITOR,EDIT_DATE,COMMENT
	// 	FROM HISTORY
	// 	WHERE ID=?
	// `, id)

	// historyテーブルにはバックアップを取得した日付（BACKUP_DATE）と
	// WORKテーブルにある編集した日付（EDIT_DATE）の両方ある。
	// 同じ日付となることがほとんどだが、HISTORYから復元した場合、HISOTRYテーブルのEDIT_DATEは入らない
	// （履歴レコードは簡易的なものなので）
	// 履歴レコード（復元された時にHISTORYに入れてるレコード）の場合、BAKCUP_DATEを返す。画面で表示したいので。
	// 初期投稿もCOMMENTが入っていない点に注意
	// EDIT_DATEがNULL かつCOMMENTがNULL ➡　初期投稿
	// EDIT_DATEがNULL かつCOMMENsTが非NULL ➡　履歴レコード
	rows, err := db.Query(`
		SELECT 
			ID,SEQ,EDITOR,
			CASE
				WHEN EDIT_DATE IS NULL AND COMMENT IS NULL THEN NULL
				WHEN EDIT_DATE IS NULL AND COMMENT IS NOT NULL THEN BACKUP_DATE
				ELSE EDIT_DATE
			END AS TEMP_DATE,
			COMMENT,
			CASE
				WHEN  EDIT_DATE IS NULL AND COMMENT IS NOT NULL THEN 1
				ELSE 0
			END AS IS_RESTORED
		FROM HISTORY
		WHERE ID=?
		ORDER BY SEQ DESC
		LIMIT 20
	`, id)

	if err != nil {
		return summary, err
	}

	defer rows.Close()

	for rows.Next() {
		var id int
		var seq int
		var editorCol sql.NullString
		var editDateCol sql.NullInt64
		var commentCol sql.NullString
		var isRestoredCol int

		err = rows.Scan(&id, &seq, &editorCol, &editDateCol, &commentCol, &isRestoredCol)
		if err != nil {
			return summary, err
		}

		// コメント（修正理由）抽出
		var comment string
		if commentCol.Valid {
			comment = commentCol.String
		} else {
			comment = "-"
		}

		// 初回投稿のバックアップは編集者・編集日の設定がない点に留意
		var editor string
		var editDate string
		if editorCol.Valid {
			editor = editorCol.String
		} else {
			editor = "-"
			// コメントが"-"の場合、初期投稿となる（COMMENTカラムがnull）
			if comment == "-" {
				editor += "(初期投稿)"
			}
		}
		if editDateCol.Valid {
			editDate = unixToStr(editDateCol.Int64)
		} else {
			editDate = "-"
		}

		var isRestored bool
		if isRestoredCol == 1 {
			isRestored = true
		} else {
			isRestored = false
		}

		s := HistorySummary{
			Id:         id,
			Seq:        seq,
			Editor:     editor,
			EditDate:   editDate,
			Comment:    comment,
			IsRestored: isRestored,
		}
		summary = append(summary, s)

	}

	return summary, nil
}

func undoWorkFromHistory(db *sql.DB, id, seq int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// tx.Commitが成功している場合、このRollbackはNOPとなる
	defer tx.Rollback()

	// historyに復元したことが分かるようにレコードを追加する処理
	// 現在日付
	now := currentDateUnix()

	// 復元されたworkでなければ、バックアップを取る
	isRestored, err := isRestoredWork(db, id)
	if err != nil {
		return err
	}

	if !isRestored {
		err = backupWork(tx, id, now)
		if err != nil {
			return err
		}
	}

	// WORKSをHISTORYから復元
	// BACKUP_SEQのみSELECT結果でなく変数を利用している点に留意
	_, err = tx.Exec(`
		UPDATE WORKS SET
			TESU = H.TESU
			,TITLE = H.TITLE
			,KIHU = H.KIHU
			,EXPLANATION = H.EXPLANATION
			,AUTHOR = H.AUTHOR
			,EDITOR = H.EDITOR
			,MAIN = H.MAIN
			,TEGOMA = H.TEGOMA
			,GOTETEGOMA = H.GOTETEGOMA
			,PUBLISH_DATE = H.PUBLISH_DATE
			,EDIT_DATE = H.EDIT_DATE
			,GOOD = H.GOOD
			,BAD = H.BAD
			,DEMAND = H.DEMAND
			,BACKUP_SEQ = ?
			,COMMENT = H.COMMENT
		FROM (
			SELECT
				TESU,TITLE,KIHU,EXPLANATION
				,AUTHOR,EDITOR
				,MAIN,TEGOMA,GOTETEGOMA
				,PUBLISH_DATE,EDIT_DATE
				,GOOD,BAD,DEMAND
				,COMMENT
			FROM HISTORY
			WHERE ID=? AND SEQ=?
		) AS H
		WHERE ID=?
	`, seq, id, seq, id)

	if err != nil {
		return err
	}

	// 編集コメント。ID-SEQを編集IDとしてコメントを設定
	editId := fmt.Sprintf("%v-%v", id, seq)
	cmt := fmt.Sprintf("編集ID:%vから復元", editId)
	// historyのSEQを新たに採番
	maxSeq, err := historyMaxSeq(tx, id)
	if err != nil {
		return err
	}
	nSeq := maxSeq + 1
	_, err = tx.Exec(`
		INSERT INTO HISTORY (
			ID,SEQ,
			BACKUP_DATE,
			COMMENT
		)
		VALUES (
			?,?,?,?
		)
	`, id, nSeq, now, cmt)

	if err != nil {
		return err
	}

	// 正常終了。commit
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
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
			GOOD,BAD,DEMAND,
			BACKUP_SEQ
		FROM WORKS WHERE ID=?
	`, id)

	var raw WorkRawRecord

	err := row.Scan(
		&raw.Id, &raw.Tesu, &raw.Title, &raw.Explanation, &raw.Author, &raw.Editor,
		&raw.Main, &raw.Tegoma, &raw.GoteTegoma, &raw.KihuJ,
		&raw.PublishDateUnix, &raw.EditDateUnix,
		&raw.Good, &raw.Bad, &raw.Demand, &raw.BackupSeq,
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

// HISTORYテーブルからIDに対応した最大のSEQカラムを取得
func historyMaxSeq(db Transer, id int) (int, error) {
	var m int
	r := db.QueryRow(`SELECT COALESCE(MAX(SEQ),0) FROM HISTORY WHERE ID=?`, id)
	err := r.Scan(&m)
	return m, err
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

// historyから復元されたworkなのかチェックする
func isRestoredWork(db Transer, id int) (bool, error) {
	var bkup sql.NullInt64
	row := db.QueryRow(`SELECT BACKUP_SEQ FROM WORKS WHERE ID=?`, id)
	err := row.Scan(&bkup)
	if bkup.Valid {
		if bkup.Int64 == 0 {
			// backup_seqがnullでなくても、0ならfalseを返す。
			// insert時に0を設定しているため必要
			return false, err
		}
		// 0以外が設定されている場合
		return true, err
	}
	// nullの場合
	return false, err
}
