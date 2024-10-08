package main

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const PORT = 8000
const DB = "./data.db"
const POST_LIMIT = 5

type (
	// /api/insert-workのBody部分
	WorkBody struct {
		Main        string `json:"main"`
		Tegoma      string `json:"tegoma"`
		GoteTegoma  string `json:"goteTegoma"`
		Kihu        string `json:"kihu"`
		Title       string `json:"title"`
		Explanation string `json:"explanation"`
		Author      string `json:"author"`
		Pic         string `json:"pic"`
		Id          int    `json:"id"`      // 編集モードで利用
		Editor      string `json:"editor"`  // 編集モードで利用
		Comment     string `json:"comment"` // 編集モードで利用
	}
)

type (
	// /api/update-evalで作品の評価を更新するときに利用。
	UpdateEvalBody struct {
		Id    int    `json:"id"`
		Key   string `json:"key"`
		Value int    `json:"value"`
	}
	// /api/update-post-evalで、投稿の評価を更新するときに利用。
	UpdatePostEvalBody struct {
		Id    int    `json:"id"`
		Seq   int    `json:"seq"`
		Key   string `json:"key"`
		Value int    `json:"value"`
	}
	// /api/update-undoで利用
	UpdateUndoBody struct {
		Id  int `json:"id"`
		Seq int `json:"seq"`
	}
)

func (b *WorkBody) steps() (int, error) {
	m := [][]int32{}
	err := json.Unmarshal([]byte(b.Main), &m)
	return len(m), err
}

// ミドルウェア
// trailing-slashがあると、以降のパスが設定されていても404にならない。
// 　ex: "/edit/"のhandler対して、"/edit/a/b/c"でアクセスしても"/edit/"のhandlerが効いてしまう。
// trailing-slashの有無を問わず、同じページを表示するために使う。
// pathが"/{path}"か"/{path}/"の形ならtrue、以外はfalse
func checkRoute(target string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/"+target || path == "/"+target+"/" {
			next.ServeHTTP(w, r)
			return
		}
		http.NotFound(w, r)
	}
}

// 現在の日付をYYYY/MM/DD HH:MM:SS形式でかえす
func currentDateStr() string {
	now := time.Now()
	nstr := now.Format("2006/01/02 15:04:05")
	return nstr
}

// 現在の日付をUNIXで取得。time.Unix(u,0)でtime型に変換できる（locationも考慮される）
func currentDateUnix() int64 {
	n := time.Now()
	u := n.Unix()
	return u
}

func unixToStr(unix int64) string {
	t := time.Unix(unix, 0)
	tstr := t.Format("2006/01/02 15:04:05")
	return tstr
}

// WORKSテーブルのIDからサムネの保存場所を計算
func thumbPath(id int) string {
	idStr := fmt.Sprint(id)
	fpath := "./public/thumb/" + idStr + ".png"
	return fpath
}

func main() {
	// db接続
	db, err := sql.Open("sqlite3", DB)
	_ = db
	if err != nil {
		log.Fatal(err)
	}

	cache := &WorksCache{}
	cache.Update(db)
	// templateに埋め込むデータとか
	rootData := NewRootData()

	// template内で実行できるカスタム関数
	customFunc := template.FuncMap{
		"add": func(v, inc int) int {
			return v + inc
		},
	}

	// html template
	// tmpl, err := template.ParseFiles("./html/edit-description.html", "./html/preview.html")
	tmpl, err := template.New("edit-description").Funcs(customFunc).ParseFiles("./html/edit-description.html", "./html/preview.html")
	if err != nil {
		log.Fatal(err)
	}

	// html template for root page
	rootTmpl, err := template.ParseFiles("./html/layout.html", "./html/nav.html", "./html/sidebar.html")
	if err != nil {
		log.Fatal(err)
	}

	// 詰将棋のコンテンツ部分のtemplate
	worksTmpl, err := template.New("works").Funcs(customFunc).ParseFiles("./html/works.html", "./html/works-meta.html", "./html/posts.html")
	if err != nil {
		log.Fatal(err)
	}

	// static folder
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	publicDir := filepath.Join(wd, "public")
	fs := http.FileServer(http.Dir(publicDir))

	// routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			// ルートページを返す
			// 現在のページをハイライトするために指定。
			// HOME -> ルートページ、EDIT -> 編集ページ、BOARD -> 掲示板
			// data := struct{ Current string }{"HOME"}
			data, ok := rootData["home"]
			if !ok {
				http.NotFound(w, r)
				return
			}
			data.Sections = cache.SectionCache
			err := rootTmpl.ExecuteTemplate(w, "layout.html", data)
			if err != nil {
				fmt.Println(err)
			}
			return
		}
		// static folder（publicフォルダ）のデータを返す。CSSやJS等。
		fs.ServeHTTP(w, r)
	})

	http.Handle("/rule/", checkRoute("rule", func(w http.ResponseWriter, r *http.Request) {
		data, ok := rootData["rule"]
		if !ok {
			http.NotFound(w, r)
			return
		}
		data.Sections = cache.SectionCache
		err := rootTmpl.ExecuteTemplate(w, "layout.html", data)
		if err != nil {
			fmt.Println(err)
		}
	}))

	http.HandleFunc("/edit/", checkRoute("edit", func(w http.ResponseWriter, r *http.Request) {
		data, ok := rootData["edit"]
		if !ok {
			http.NotFound(w, r)
			return
		}
		err := rootTmpl.ExecuteTemplate(w, "layout.html", data)
		if err != nil {
			fmt.Println(err)
		}
	}))

	// 新規作成画面
	http.HandleFunc("/edit/description", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		main := r.FormValue("main")
		tegoma := r.FormValue("tegoma")
		Data := struct {
			Main     string
			Tegoma   string
			IsRevise bool
		}{
			Main: main, Tegoma: tegoma, IsRevise: false,
		}
		// w.Header().Add("Content-Type", "text/html")
		// w.Write([]byte(`<h1>TEST</h1>`))
		err = tmpl.ExecuteTemplate(w, "edit-description.html", Data)
		if err != nil {
			fmt.Println(err)
		}
	})

	// 修正画面
	http.HandleFunc("/revise/", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		idStr := query.Get("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			fmt.Println(err)
			return
		}
		rec, err := getWork(db, int(id))
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		edits, err := getHistorySummary(db, int(id))
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		data := struct {
			Id              int
			Main            string
			Tegoma          string
			GoteTegoma      string
			KihuJ           string
			Title           string
			Explanation     string
			Author          string
			PublishDateUnix int
			IsRevise        bool
			Edits           []HistorySummary
		}{
			Id:   int(id),
			Main: rec.Main, Tegoma: rec.Tegoma,
			GoteTegoma: rec.GoteTegoma, KihuJ: rec.KihuJ,
			Title: rec.Title, Explanation: rec.Explanation, Author: rec.Author,
			PublishDateUnix: rec.PublishDateUnix,
			IsRevise:        true,
			Edits:           edits,
		}

		err = tmpl.ExecuteTemplate(w, "edit-description.html", data)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	http.HandleFunc("POST /preview", func(w http.ResponseWriter, r *http.Request) {
		data := r.FormValue("data")
		author := r.FormValue("author")
		exp := r.FormValue("explanation")
		title := r.FormValue("title")
		editor := r.FormValue("editor")
		comment := r.FormValue("comment")
		publishUnix := r.FormValue("publish-unix")
		// reviseが遷移元URLに含まれれば編集モード、含まれなければ新規モード
		isReviseMode := strings.Contains(r.Referer(), "/revise/")

		// title := r.FormValue("title")
		form := EditFormData{
			Explanation: exp,
			Author:      author,
			Editor:      editor,
			Title:       title,
		}

		// 編集モードならば編集者をセット。過去の編集者はカンマで区切って並べる
		// Idもセットする
		if isReviseMode {
			idStr := r.FormValue("id")
			form.WorkId = idStr
			form.Editor = editor
		} else {
			form.WorkId = "X"
		}
		// プレビューに必要なデータを取得
		var dt Dates
		if isReviseMode {
			publishUnixInt, err := strconv.ParseInt(publishUnix, 10, 64)
			if err != nil {
				fmt.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			dt.PublishDate = unixToStr(publishUnixInt)
			dt.EditDate = currentDateStr()
		} else {
			dt.PublishDate = currentDateStr()
			dt.EditDate = "-"
		}
		// dt := Dates{
		// 	PublishDate: currentDateStr(),
		// 	EditDate:    "-",
		// }
		wc := preview(data, form, dt)

		// templateに落とすための変数。io.Writer型である必要あり。
		// works.htmlに埋め込んだ値
		wBuf := &bytes.Buffer{}

		err = worksTmpl.ExecuteTemplate(wBuf, "works.html", wc)
		if err != nil {
			fmt.Println(err)
		}

		// preview.htmlテンプレートには３つのモード(revise,new,undo)がある。
		// /previewエンドポイントでは、revise、newのいずれかとなる。
		// /undoエンドポイントでもpreview.htmlを使い、一律undoとなる
		var mode string
		if isReviseMode {
			mode = "revise"
		} else {
			mode = "new"
		}
		wdata := struct {
			Mode          string
			CurrentEditor string
			EditComment   string
			Content       template.HTML
		}{
			Mode:          mode,
			CurrentEditor: editor,
			EditComment:   comment,
			Content:       template.HTML(wBuf.String()),
		}

		err = tmpl.ExecuteTemplate(w, "preview.html", wdata)
		if err != nil {
			fmt.Println(err)
		}
	})

	http.HandleFunc("POST /api/insert-work", func(w http.ResponseWriter, r *http.Request) {
		body := &WorkBody{}
		dec := json.NewDecoder(r.Body)
		dec.Decode(body)

		maxId, err := insertWork(db, body)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// dataURLの"data:image/png;base64,"部分を切り取るために、コンマでsplit
		dataArr := strings.Split(body.Pic, ",")
		if len(dataArr) != 2 {
			fmt.Println("画像のdataURL部分が不正です。リクエスト内の画像は以下のとおり:")
			fmt.Println(body.Pic)
			w.Write([]byte("サムネの登録ができませんでしたが、作品登録は問題なく出来ました。ご協力ありがとうございました。"))
			return
		}

		// base64部部のみを抽出し、デコード
		pic := dataArr[1]
		img, err := base64.StdEncoding.DecodeString(pic)
		if err != nil {
			fmt.Println(err)
			fmt.Println("画像をbase64からデコードできませんでした。リクエスト内の画像は以下のとおり:")
			fmt.Println(body.Pic)
			w.Write([]byte("サムネの登録ができませんでしたが、作品登録は問題なく出来ました。ご協力ありがとうございました。"))
			return
		}

		fpath := thumbPath(maxId)
		f, err := os.Create(fpath)
		if err != nil {
			fmt.Println(err)
			fmt.Println("ファイルの保存先が取得できませんでした。保存先パス：")
			fmt.Println(fpath)
			w.Write([]byte("サムネの登録ができませんでしたが、作品登録は問題なく出来ました。ご協力ありがとうございました。"))
			return
		}
		defer f.Close()

		_, err = f.Write(img)
		if err != nil {
			fmt.Println(err)
			fmt.Println("ファイルの書き込みが出来ませんでした")
			w.Write([]byte("サムネの登録ができませんでしたが、作品登録は問題なく出来ました。ご協力ありがとうございました。"))
			return
		}

		w.Write([]byte("登録成功しました。反映まで少し時間がかかる場合があります。編集ページは全て閉じてOKです。作品投稿ありがとうございましたm(__)m"))

		// リンクのcacheを更新
		cache.Update(db)
	})

	http.HandleFunc("/api/update-work", func(w http.ResponseWriter, r *http.Request) {
		body := &WorkBody{}
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(body)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// @Todo
		//  - backup and update work db
		err = updateWork(db, body)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write([]byte("更新成功"))
		// @Todo
		//  - update cache
		cache.Update(db)
	})

	http.HandleFunc("POST /api/update-eval", func(w http.ResponseWriter, r *http.Request) {
		data := &UpdateEvalBody{}
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(data)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		cnt, err := updateWorkEval(db, data)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Header().Add("Content-Type", "text/plain")
		w.Write([]byte(fmt.Sprint(cnt)))
	})

	http.HandleFunc("POST /api/insert-post", func(w http.ResponseWriter, r *http.Request) {
		name := r.FormValue("name")
		comment := r.FormValue("comment")
		idStr := r.FormValue("id")
		commentType := r.FormValue("type")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = insertPost(db, name, comment, commentType, int(id))
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(200)
	})

	http.HandleFunc("GET /works/", checkRoute("works", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		paramid := query.Get("id")

		id, err := strconv.ParseInt(paramid, 10, 64)

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// dbからデータ取得
		wr, err := getWork(db, int(id))
		if err != nil {
			fmt.Println(err)
			http.NotFound(w, r)
			return
		}

		// dbから投稿（post）を取得
		// エラーが出ても取得しないだでなのでエラーにしない
		posts, _ := getPosts(db, int(id), 0, POST_LIMIT)
		wr.Posts = posts

		// works.htmlテンプレートに埋め込み
		wBuf := &bytes.Buffer{}
		err = worksTmpl.ExecuteTemplate(wBuf, "works.html", wr)

		if err != nil {
			fmt.Println(err)
			http.NotFound(w, r)
			return
		}

		wdata := Record{
			Meta: template.HTML(`
			<meta name="description" content="投稿された詰将棋作品です。ぜひ解いて楽しんでください。また、改善点があれば、編集者のためにコメントを残してください。">
			<link rel="stylesheet" href="/css/works.css">
			<script src="/js/works/main.js" type="module"></script>
			<title>詰将棋投稿作品</title>
			`),
			Content:  template.HTML(wBuf.String()),
			Sections: cache.SectionCache,
		}
		err = rootTmpl.ExecuteTemplate(w, "layout.html", wdata)
		if err != nil {
			fmt.Println(err)
		}
	}))

	http.HandleFunc("POST /api/update-post-eval", func(w http.ResponseWriter, r *http.Request) {
		body := &UpdatePostEvalBody{}
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(body)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		cnt, err := updatePostEval(db, body)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", "text/plain")
		w.Write([]byte(fmt.Sprint(cnt)))
	})

	http.HandleFunc("GET /api/get-next-posts", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		idStr := query.Get("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		lastSeqStr := query.Get("seq")
		lastSeq, err := strconv.ParseInt(lastSeqStr, 10, 64)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// DBの最大SEQを取得し、超えているようならクライアントに伝える
		maxSeq := postMaxSeq(db, int(id))
		if int(lastSeq) >= maxSeq {
			// 204 No-Contentを返す
			w.WriteHeader(http.StatusNoContent)
			return
		}

		recs, err := getPosts(db, int(id), int(lastSeq), POST_LIMIT)

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		postsData := map[string][]PostRecord{
			"Posts": recs,
		}

		// {{define template "name"}}で宣言したテンプレートは、その名前を使う必要がある点に注意。ファイル名じゃダメ
		err = worksTmpl.ExecuteTemplate(w, "posts", postsData)

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	})

	// 同じ手数の次（前）の作品に遷移させる
	http.HandleFunc("GET /api/next-work", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		idStr := query.Get("id")
		tesuStr := query.Get("tesu")
		// valが正なら「次」の作品、0以下なら「前の作品」。bool値のが良かったかも？
		valueStr := query.Get("value")
		id, idErr := strconv.ParseInt(idStr, 10, 64)
		tesu, tesuErr := strconv.ParseInt(tesuStr, 10, 64)
		value, valueErr := strconv.ParseInt(valueStr, 10, 64)

		if idErr != nil || tesuErr != nil || valueErr != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		nextId, err := getNextWork(db, int(id), int(tesu), int(value))
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write([]byte(fmt.Sprint(nextId)))
	})

	// undoモードでプレビュー画面を表示する
	http.HandleFunc("GET /undo", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		idStr := query.Get("id")
		seqStr := query.Get("seq")

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		seq, err := strconv.ParseInt(seqStr, 10, 64)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		hist, err := getHistory(db, int(id), int(seq))

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		form := EditFormData{
			Explanation: hist.Explanation,
			Author:      hist.Author,
			Editor:      hist.Editor,
			Title:       hist.Title,
			WorkId:      fmt.Sprint(hist.Id),
		}

		dates := Dates{
			PublishDate: hist.PublishDate,
			EditDate:    hist.EditDate,
		}

		wc := &WorksContent{
			EditFormData: form,
			// Dates:        dates,
			Main:       hist.Main,
			Tegoma:     hist.Tegoma,
			GoteTegoma: hist.GoteTegoma,
			Kihu:       hist.Kihu,
			KihuJ:      hist.KihuJ,
			IsPreview:  true,
			Tesu:       hist.Tesu,
		}

		wc.PublishDate = dates.PublishDate
		wc.EditDate = dates.EditDate

		fmt.Println(wc.Dates.EditDate, wc.Dates.PublishDate)

		wBuf := &bytes.Buffer{}
		err = worksTmpl.ExecuteTemplate(wBuf, "works.html", wc)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		wdata := struct {
			Mode    string
			Content template.HTML
		}{
			"undo",
			template.HTML(wBuf.String()),
		}

		err = tmpl.ExecuteTemplate(w, "preview.html", wdata)
		if err != nil {
			fmt.Println(err)
		}
	})

	// preview画面（undoモード）の登録処理。workの内容をhistoryから戻す
	http.HandleFunc("/api/update-undo", func(w http.ResponseWriter, r *http.Request) {
		bd := &UpdateUndoBody{}
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(bd)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = undoWorkFromHistory(db, bd.Id, bd.Seq)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write([]byte("更新成功"))
		cache.Update(db)
	})

	// localhostをつけないと、起動時にfw許可のメッセージが出る
	// つけると、スマホ等別デバイスからのアクセスができなくなる
	http.ListenAndServe("localhost:8000", nil)
	// http.ListenAndServe(":8000", nil)

}
