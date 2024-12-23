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

const LONG = 99999

// *sql.DBや*sql.TxでもQueryRowが使えるように、、、
type Transer interface {
	QueryRow(string, ...any) *sql.Row
	Exec(string, ...any) (sql.Result, error)
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

// サイドバーの作品リンク用
type WorkLink struct {
	Id    int
	Title string
}

// 作品一覧に表示する情報
type WorkListData struct {
	Id          int
	Tesu        int
	Title       string
	Author      string
	Kihu        string
	PublishDate string
	EditDate    string
	Good        int
	Bad         int
	Thumb       string
}

type WorkList []WorkListData

// tempalteに渡す用
type WorkListTmpl struct {
	Heading string
	Tesu    int
	WorkList
}

// WorkListフィールドの最大IDを取得
func (tmp WorkListTmpl) Max() int {
	max := 0
	for _, data := range tmp.WorkList {
		if max <= data.Id {
			max = data.Id
		}
	}
	return max
}

// Threadsテーブルのレコードの型
type ThreadRecord struct {
	Id                  int
	Title               string
	Author              string
	CreatedDateUnix     int
	LastCommentDateUnix int
	// 日付変換後（カラムにはない）
	CreatedDateStr     string
	LastCommentDateStr string
}

// Commentsテーブルのレコードの型
type CommentRecord struct {
	ThreadId    int
	Seq         int
	Commenter   string
	Comment     string
	ReplyTo     int
	CommentDate int
	// 日付変換後（カラムにはない）
	CommentDateStr string
}

// 削除されたWorkのサマリ情報
type DeletedWork struct {
	Id      int
	Title   string
	Kihu    string
	DelDate string
	Reason  string
	// サムネのパス
	Thumb string
}

type WorksMap map[int][]WorkLink

type WorksCache struct {
	data         WorksMap
	SectionCache []Section
}

func (wc *WorksCache) Update(db *sql.DB) error {
	// 初期化しないとエラーになる。更新の都度初期化する。そうしないとUpdateで二重に設定される。
	wc.data = WorksMap{}
	err := wc.data.update(db)
	if err != nil {
		// updateに失敗したらreturn。SectionCacheも更新しない。
		err = stack("&WorksCache.Update", err)
		return err
	}
	wc.SectionCache = wc.data.section()
	return err
}

func (c WorksMap) update(db *sql.DB) error {
	rows, err := db.Query(`
		SELECT ID,TESU,TITLE FROM WORKS WHERE DEL_FLG IS NULL
		ORDER BY TESU ASC,PUBLISH_DATE DESC
		LIMIT 50;
	`)
	if err != nil {
		err = stack("WorksMap.update", err)
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var id int
		var tesu int
		var title string
		err = rows.Scan(&id, &tesu, &title)
		if err != nil {
			err = stack("WorksMap.update", err)
			return err
		}
		info := WorkLink{id, title}
		if tesu > 11 {
			c[LONG] = append(c[LONG], info)
		} else {
			c[tesu] = append(c[tesu], info)
		}
	}
	return err
}

func (c WorksMap) section() []Section {
	ret := []Section{}
	keys := c.Sort()
	for _, key := range keys {
		sec := Section{}
		if key == LONG {
			sec.Heading = "長手数集"
			sec.Tesu = LONG_TESU
		} else {
			sec.Heading = fmt.Sprintf("%v手詰集", key)
			sec.Tesu = key
		}

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
		err = stack("insertWork", err)
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
		err = stack("insertWork", err)
		return maxId, err
	}

	defer tx.Rollback()

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
		err = stack("insertWork", err)
		return maxId, err
	}

	err = tx.Commit()
	if err != nil {
		err = stack("insertWork", err)
	}

	return maxId, err
}

// 投稿速品（work）を更新する。編集画面からの登録時に利用
func updateWork(db *sql.DB, bd *WorkBody) error {
	tx, err := db.Begin()
	if err != nil {
		err = stack("updateWork", err)
		return err
	}

	defer tx.Rollback()

	// バックアップから復元されたWorkかをチェック
	isRestored, err := isRestoredWork(tx, bd.Id)
	if err != nil {
		err = stack("updateWork", err)
		return err
	}

	// バックアップ日時
	now := currentDateUnix()

	// バックアップ復元されていない場合、バックアップ取得する
	if !isRestored {
		err = backupWork(tx, bd.Id, now)
		if err != nil {
			err = stack("updateWork", err)
			return err
		}
	}

	// workテーブルを更新する
	err = updateWorkFromEdit(tx, bd, now)
	if err != nil {
		err = stack("updateWork", err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		err = stack("updateWork", err)
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
		err = stack("backupWork", err)
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
		err = stack("backupWork", err)
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
		err = stack("updateWorkFromEdit", err)
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
		err = stack("updateWorkFromEdit", err)
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
		err = stack("getHistory", err)
		return rec, err
	}
	work := raw.parse()
	rec = HistoryRecord{
		Seq:        rawSeq,
		BackupDate: backupDate,
		WorkRecord: work,
	}
	return rec, err
}

// historyテーブルのサマリーを取得。編集画面のリンク表示用
func getHistorySummary(db *sql.DB, id int) ([]HistorySummary, error) {
	var summary []HistorySummary
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
		err = stack("getHistorySummary", err)
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
			err = stack("getHistorySummary", err)
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
		err = stack("undoWorkFromHistory", err)
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
		err = stack("undoWorkFromHistory", err)
		return err
	}

	if !isRestored {
		err = backupWork(tx, id, now)
		if err != nil {
			err = stack("undoWorkFromHistory", err)
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
		err = stack("undoWorkFromHistory", err)
		return err
	}

	// 編集コメント。ID-SEQを編集IDとしてコメントを設定
	editId := fmt.Sprintf("%v-%v", id, seq)
	cmt := fmt.Sprintf("編集ID:%vから復元", editId)
	// historyのSEQを新たに採番
	maxSeq, err := historyMaxSeq(tx, id)
	if err != nil {
		err = stack("undoWorkFromHistory", err)
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
		err = stack("undoWorkFromHistory", err)
		return err
	}

	// 正常終了。commit
	err = tx.Commit()
	if err != nil {
		err = stack("undoWorkFromHistory", err)
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
		err = stack("updateWorkEval", err)
		return cnt, err
	}

	defer tx.Rollback()

	query := fmt.Sprintf("SELECT %v FROM WORKS WHERE ID=?", key)
	row := tx.QueryRow(query, id)
	err = row.Scan(&cnt)
	if err != nil {
		err = stack("updateWorkEval", err)
		return cnt, err
	}

	cnt += eb.Value // eb.Valueは1 or -1

	query = fmt.Sprintf("UPDATE WORKS SET %v=%v WHERE ID=?", key, cnt)
	_, err = tx.Exec(query, id)
	if err != nil {
		err = stack("updateWorkEval", err)
		return cnt, err
	}

	err = tx.Commit()
	if err != nil {
		err = stack("updateWorkEval", err)
	}
	return cnt, err
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
		FROM WORKS 
		WHERE 
			ID=? 
	`, id)

	var raw WorkRawRecord

	err := row.Scan(
		&raw.Id, &raw.Tesu, &raw.Title, &raw.Explanation, &raw.Author, &raw.Editor,
		&raw.Main, &raw.Tegoma, &raw.GoteTegoma, &raw.KihuJ,
		&raw.PublishDateUnix, &raw.EditDateUnix,
		&raw.Good, &raw.Bad, &raw.Demand, &raw.BackupSeq,
	)

	if err != nil {
		err = stack("getWork", err)
		return WorkRecord{}, err
	}

	rec := raw.parse()
	return rec, err
}

// WORKSの削除フラグを設定する
func deleteWork(db *sql.DB, id int, editor, reason string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	dt := currentDateUnix()

	_, err = tx.Exec(`
		UPDATE WORKS SET
		  DEL_FLG=?
		  ,DEL_BY=?
		  ,DEL_REASON=?
		  ,DEL_DATE=?
		WHERE ID=?
	`, 1, editor, reason, dt, id)

	if err != nil {
		err = stack("deleteWork", err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		err = stack("deleteWork", err)
	}

	return err
}

// WORKテーブルを復元（削除から戻す）
func restoreWork(db *sql.DB, id int) error {
	tx, err := db.Begin()
	if err != nil {
		err = stack("restoreWork", err)
		return err
	}

	defer tx.Rollback()

	_, err = tx.Exec(`
		UPDATE WORKS SET
		  DEL_FLG=NULL
		  ,DEL_BY=NULL
		  ,DEL_REASON=NULL
		  ,DEL_DATE=NULL
		WHERE ID=?
	`, id)

	if err != nil {
		err = stack("restoreWork", err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		err = stack("restoreWork", err)
	}

	return err
}

// 手数に応じたworksを取得。作品一覧ページで利用
// 手数を問わずに抽出したい場合、tesuは0で呼ぶこと
func getWorksList(db *sql.DB, lastId, tesu int) ([]WorkListTmpl, error) {
	var query string
	// 長手数はまとめるので閾値。
	longThresh := 11

	wmap := map[int]WorkList{}
	var tmpls []WorkListTmpl

	if tesu == 0 {
		// 全量
		// db.QueryRowのパラメタ数をそろえるためにTESU>=?としている。呼び出し時に0を指定すること、、、
		query = `
			SELECT 
				ID,TESU,TITLE,KIHU,AUTHOR,PUBLISH_DATE,EDIT_DATE,GOOD,BAD
			FROM WORKS
			WHERE 
				ID>=?
				AND TESU>=?
				AND DEL_FLG IS NULL
			ORDER BY TESU,ID
			LIMIT 50
		`
	} else {
		// 手数指定。長手数は閾値を超える場合はまとめる
		if tesu > longThresh {
			// 長手数の場合　WHEREのTESUが>=になってる
			query = `
			SELECT 
				ID,TESU,TITLE,KIHU,AUTHOR,PUBLISH_DATE,EDIT_DATE,GOOD,BAD 
			FROM WORKS
			WHERE 
				ID>=?
				AND TESU>11
				AND DEL_FLG IS NULL
			ORDER BY TESU,ID
			LIMIT 20
			`
		} else {
			// 短手数　TESU＝？
			query = `
			SELECT 
				ID,TESU,TITLE,KIHU,AUTHOR,PUBLISH_DATE,EDIT_DATE,GOOD,BAD 
			FROM WORKS
			WHERE 
				ID>=?
				AND TESU=?
				AND DEL_FLG IS NULL
			ORDER BY TESU,ID
			LIMIT 20
			`
		}
	}

	rows, err := db.Query(query, lastId, tesu)
	if err != nil {
		err = stack("getWorksList", err)
		return tmpls, err
	}

	defer rows.Close()

	for rows.Next() {
		var id, tesu int
		var title, kihu, author string
		var pdate int64
		var edate sql.NullInt64
		var good, bad int
		err := rows.Scan(&id, &tesu, &title, &kihu, &author, &pdate, &edate, &good, &bad)

		if err != nil {
			err = stack("getWorksList", err)
			return tmpls, err
		}

		// kihuを整形
		var kArr []string
		err = json.Unmarshal([]byte(kihu), &kArr)
		if err != nil {
			err = stack("getWorksList", err)
			return tmpls, err
		}
		if len(kArr) <= 1 {
			err = errors.New("length of kihu is less than 1")
			err = stack("getWorksList", err)
			return tmpls, err
		}
		// 最初の要素は空白なので落とす
		kArr = kArr[1:]
		kihuStr := kArr[0]
		for _, k := range kArr[1:] {
			kihuStr += " " + k
		}
		// pdate,edateをYYYY-MM-DDに変換
		pdateStr := unixToStr(pdate)
		var edateStr string
		if edate.Valid {
			edateStr = unixToStr(edate.Int64)
		} else {
			edateStr = "-"
		}
		// サムネのパス
		thumb := "/thumb/" + fmt.Sprint(id) + ".png"

		data := WorkListData{
			Id:          id,
			Tesu:        tesu,
			Title:       title,
			Kihu:        kihuStr,
			Author:      author,
			PublishDate: pdateStr,
			EditDate:    edateStr,
			Good:        good,
			Bad:         bad,
			Thumb:       thumb,
		}

		if tesu > longThresh {
			wmap[LONG] = append(wmap[LONG], data)
		} else {
			wmap[tesu] = append(wmap[tesu], data)
		}
	}

	// WorkListMap型をWorkListTmpl型に変換
	// まずはmapのキーを取得
	var keys []int
	for k := range wmap {
		keys = append(keys, k)
	}
	// mapだと走査時の順番が保証されないので、キーをソートしてvalueにアクセス
	slices.Sort(keys)
	for _, key := range keys {
		tmpl := WorkListTmpl{
			Tesu:     key,
			WorkList: wmap[key],
		}
		if key == LONG {
			tmpl.Heading = "長手数作品"
		} else {
			tmpl.Heading = fmt.Sprintf("%v手詰作品", key)
		}
		tmpls = append(tmpls, tmpl)
	}
	return tmpls, err
}

// 削除されたworksを取得
func getDeletedWorks(db *sql.DB) ([]DeletedWork, error) {

	var list []DeletedWork

	rows, err := db.Query(`
		SELECT 
			ID,TITLE,KIHU,DEL_DATE,DEL_REASON 
		FROM WORKS
		WHERE DEL_FLG = 1
		ORDER BY TESU
		LIMIT 50
	`)
	if err != nil {
		err = stack("getDeletedWorks", err)
		return list, err
	}

	defer rows.Close()

	for rows.Next() {
		var id, dateInt int
		var title, reason, kihuJ string
		err := rows.Scan(&id, &title, &kihuJ, &dateInt, &reason)
		if err != nil {
			err = stack("getDeletedWorks", err)
			return list, err
		}
		// unix->文字列
		dateStr := unixToStr(int64(dateInt))
		// json形式の棋譜→文字列。最初の要素は空白なので除外
		var kihuArr []string
		err = json.Unmarshal([]byte(kihuJ), &kihuArr)
		if err != nil {
			err = stack("getDeletedWorks", err)
			return list, err
		}
		kihuArr = kihuArr[1:]
		var kihu string
		for _, k := range kihuArr {
			kihu += " " + k
		}

		// サムネのパス
		thumb := "/thumb/" + fmt.Sprint(id) + ".png"
		list = append(list, DeletedWork{
			Id:      id,
			Title:   title,
			Kihu:    kihu,
			DelDate: dateStr,
			Reason:  reason,
			Thumb:   thumb,
		})
	}

	return list, err
}

// WORKSテーブルから最大IDを取得。サムネ保存場所に必要
func worksMaxId(db Transer) int {
	var m int
	r := db.QueryRow("SELECT MAX(ID) FROM WORKS")
	r.Scan(&m)
	return m
}

// 作品ページのコメントを投稿する。
// 掲示板のコメントではないので注意
func insertPost(db *sql.DB, name, comment, commentType string, id int) (PostRecord, error) {
	var rec PostRecord
	tx, err := db.Begin()
	if err != nil {
		return rec, err
	}

	defer tx.Rollback()

	maxSeq, err := postMaxSeq(tx, id)
	if err != nil {
		err = stack("insertPost", err)
		return rec, err
	}

	nextSeq := maxSeq + 1

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
		err = stack("insertPost", err)
		return rec, err
	}

	err = tx.Commit()
	if err != nil {
		err = stack("insertPost", err)
		return rec, err
	}

	// 投稿内容をクライアントに表示するために必要
	rec = PostRecord{
		Id: id, Seq: nextSeq,
		Name: name, Comment: comment,
		Type:        commentType,
		PostDate:    int(pdate),
		PostDateStr: unixToStr(pdate),
	}

	return rec, nil
}

// POSTSテーブルからidに対応した最大のSEQカラムを取得
func postMaxSeq(db Transer, id int) (int, error) {
	var m int
	r := db.QueryRow(`SELECT COALESCE(MAX(SEQ),0) FROM POSTS WHERE ID=?`, id)
	err := r.Scan(&m)
	if err != nil {
		err = stack("postMaxSeq", err)
	}
	return m, err
}

// COMMENTテーブルからidに対応した最大のSEQカラムを取得
func commentMaxSeq(db Transer, id int) (int, error) {
	var m int
	r := db.QueryRow(`SELECT COALESCE(MAX(SEQ),0) FROM COMMENTS WHERE THREAD_ID=?`, id)
	err := r.Scan(&m)
	if err != nil {
		err = stack("commentMaxSeq", err)
	}
	return m, err
}

// HISTORYテーブルからIDに対応した最大のSEQカラムを取得
func historyMaxSeq(db Transer, id int) (int, error) {
	var m int
	r := db.QueryRow(`SELECT COALESCE(MAX(SEQ),0) FROM HISTORY WHERE ID=?`, id)
	err := r.Scan(&m)
	if err != nil {
		err = stack("historyMaxSeq", err)
	}
	return m, err
}

// 投稿内容を取得。降順で表示する。offset=0は、初回表示
func getPosts(db *sql.DB, id, offset, limit int) ([]PostRecord, error) {
	recs := []PostRecord{}
	var query string
	var rows *sql.Rows
	var err error
	if offset == 0 {
		query = fmt.Sprintf(`
		SELECT * FROM POSTS 
		WHERE ID=?
		ORDER BY SEQ DESC
		LIMIT %v
		`, limit)
		rows, err = db.Query(query, id)
	} else {
		query = fmt.Sprintf(`
		SELECT * FROM POSTS 
		WHERE ID=? 
		AND SEQ<?
		ORDER BY SEQ DESC
		LIMIT %v
		`, limit)
		rows, err = db.Query(query, id, offset)
	}

	if err != nil {
		err = stack("getPosts", err)
		return recs, err
	}

	defer rows.Close()

	for rows.Next() {
		r := PostRecord{}
		err = rows.Scan(
			&r.Id, &r.Seq, &r.Name, &r.Comment, &r.Type,
			&r.PostDate, &r.Good, &r.Bad,
		)
		if err != nil {
			err = stack("getPosts", err)
			return recs, err
		}
		r.PostDateStr = unixToStr(int64(r.PostDate))
		recs = append(recs, r)
	}

	return recs, err
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
		WHERE TESU=? AND ID>? AND DEL_FLG IS NULL
		ORDER BY ID ASC
		LIMIT 1
		`
	} else {
		query = `
		SELECT ID FROM WORKS 
		WHERE TESU=? AND ID<? AND DEL_FLG IS NULL
		ORDER BY ID DESC
		LIMIT 1
		`
	}

	var retId int
	row := db.QueryRow(query, tesu, id)
	err := row.Scan(&retId)
	if err != nil {
		err = stack("getNextWork", err)
	}
	return retId, err
}

// historyから復元されたworkなのかチェックする
func isRestoredWork(db Transer, id int) (bool, error) {
	var bkup sql.NullInt64
	row := db.QueryRow(`SELECT BACKUP_SEQ FROM WORKS WHERE ID=?`, id)
	err := row.Scan(&bkup)
	if err != nil {
		err = stack("isRestoredWork", err)
		return false, err
	}

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

// Threadテーブルに新規レコードを挿入し、最大ID（挿入されたレコードのID）を返す
// LAST_COMMENT_DATE（最終コメント投稿日）は、レコード作成日と同じ値をセット
func insertThread(db *sql.DB, name, title string) (int, error) {
	var maxId int
	tx, err := db.Begin()
	if err != nil {
		err = stack("insertThread", err)
		return maxId, err
	}
	// tx.Commitが成功している場合、このRollbackはNOPとなる
	// good practiceらしいぞ。
	defer tx.Rollback()

	now := currentDateUnix()

	_, err = tx.Exec(`
		INSERT INTO THREADS
		(
			TITLE,AUTHOR,CREATED_DATE,LAST_COMMENT_DATE
		)
		VALUES (
			?,?,?,?
		);
	`, title, name, now, now)

	if err != nil {
		err = stack("insertThread", err)
		return maxId, err
	}

	maxId, err = getMaxThreadId(tx)

	if err != nil {
		err = stack("insertThread", err)
		return maxId, err
	}

	err = tx.Commit()
	if err != nil {
		err = stack("insertThread", err)
	}

	return maxId, err
}

func updateThreadCommentDate(db Transer, id int, dt int64) error {
	_, err := db.Exec(`
		UPDATE THREADS SET LAST_COMMENT_DATE=? WHERE ID=?
	`, dt, id)

	if err != nil {
		err = stack("updateThreadCommentDate", err)
		return err
	}
	return nil
}

func getMaxThreadId(tx Transer) (int, error) {
	row := tx.QueryRow(`SELECT MAX(ID) FROM THREADS`)
	var id int
	err := row.Scan(&id)
	return id, err
}

// 掲示板ページで表示するスレッド一覧を取得する。
// 表示順はコメント投稿日の降順、かつLIMITで取得数を区切っている。
// そのため、次の一覧を取得する歳、「現在表示されている日付より前の日付」に絞る必要あり
// OFFSET句を使う手もあるが、今回は使わない
func getThreads(db *sql.DB, limit int, lastDate int64) ([]ThreadRecord, error) {
	var recs []ThreadRecord
	var query string
	var rows *sql.Rows
	var err error
	if lastDate == 0 {
		// 初期表示のとき。投稿日で絞る必要がないのでWHERE句なし
		query = fmt.Sprintf(`
		  SELECT ID,TITLE,AUTHOR,CREATED_DATE,LAST_COMMENT_DATE
		  FROM THREADS
		  ORDER BY LAST_COMMENT_DATE DESC
		  LIMIT %v
		`, limit)
		rows, err = db.Query(query)
	} else {
		// 次の一覧を取得するとき。lastDateより前のデータに絞って取得する必要あり
		query = fmt.Sprintf(`
		  SELECT ID,TITLE,AUTHOR,CREATED_DATE,LAST_COMMENT_DATE
		  FROM THREADS
		  WHERE LAST_COMMENT_DATE < ?
		  ORDER BY LAST_COMMENT_DATE DESC
		  LIMIT %v
	    `, limit)
		rows, err = db.Query(query, lastDate)
	}

	if err != nil {
		err = stack("getThreads", err)
		return recs, err
	}

	defer rows.Close()

	for rows.Next() {
		var r ThreadRecord
		err = rows.Scan(
			&r.Id, &r.Title, &r.Author,
			&r.CreatedDateUnix, &r.LastCommentDateUnix,
		)

		// scanがエラーの場合はreturn
		if err != nil {
			err = stack("getThreads", err)
			return recs, err
		}

		r.CreatedDateStr = unixToStr(int64(r.CreatedDateUnix))
		if r.LastCommentDateUnix == 0 {
			// db上はnullだけどThreadRecord型としてはintで定義しているので、nullのとき0になる。その場合はハイフンを設定
			r.LastCommentDateStr = "-"
		} else {
			r.LastCommentDateStr = unixToStr(int64(r.LastCommentDateUnix))
		}
		recs = append(recs, r)
	}

	return recs, nil
}

// 掲示板のスレッドへのコメントを総んっゆう
func insertComment(db *sql.DB, id, reply int, author, comment string) error {
	tx, err := db.Begin()
	if err != nil {
		err = stack("insertComment", err)
		return err
	}

	defer tx.Rollback()

	maxSeq, err := commentMaxSeq(tx, id)

	if err != nil {
		err = stack("insertComment", err)
		return err
	}

	nextSeq := maxSeq + 1
	now := currentDateUnix()

	_, err = tx.Exec(`
	  INSERT INTO COMMENTS (
		THREAD_ID,
		SEQ,
		COMMENTER,
		COMMENT,
		REPLY_TO,
		COMMENT_DATE
	  )
	  VALUES (
		?,?,?,
		?,?,?
	  )
	`, id, nextSeq, author, comment, reply, now)

	if err != nil {
		err = stack("insertComment", err)
		return err
	}

	// Threadsテーブルの最終投稿日も更新
	err = updateThreadCommentDate(tx, id, now)
	if err != nil {
		err = stack("insertComment", err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		err = stack("insertComment", err)
	}

	return err
}

func getComments(db *sql.DB, threadId int) ([]CommentRecord, error) {
	var comments []CommentRecord
	rows, err := db.Query(`
		SELECT 
		  THREAD_ID,SEQ,COMMENTER,COMMENT,REPLY_TO,COMMENT_DATE
		FROM COMMENTS
		WHERE THREAD_ID=?
		LIMIT 300
	`, threadId)

	if err != nil {
		err = stack("getComments", err)
		return comments, err
	}

	defer rows.Close()

	for rows.Next() {
		var c CommentRecord
		err = rows.Scan(&c.ThreadId, &c.Seq, &c.Commenter, &c.Comment, &c.ReplyTo, &c.CommentDate)
		// scanがエラーだったらreturn
		if err != nil {
			err = stack("getComments", err)
			return comments, err
		}
		// 日付をフォーマットして設定
		c.CommentDateStr = unixToStr(int64(c.CommentDate))
		comments = append(comments, c)
	}

	return comments, nil
}
