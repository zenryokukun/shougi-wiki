package main

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const PORT = 8000
const DB = "./data.db"
const POST_LIMIT = 5
const LONG_TESU = 99999
const BOARD_LIMIT = 50

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
	if err != nil {
		err = stack("&WorkBody.steps", err)
	}
	return len(m), err
}

// ミドルウェア
// trailing-slashがあると、以降のパスが設定されていても404にならない。
// 　ex: "/edit/"のhandler対して、"/edit/a/b/c"でアクセスしても"/edit/"のhandlerが効いてしまう。
// trailing-slashの有無を問わず、同じページを表示するために使う。
// pathが"/{path}"か"/{path}/"の形ならtrue、以外はfalse
func checkRoute(target string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/"+target || path == "/"+target+"/" {
			next.ServeHTTP(w, r)
			return
		}
		err := errors.New("page not found at `checkRoute` func")
		logErr(r, err)
		http.NotFound(w, r)
	})
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
	// format: "2006/01/02 15:04:05"
	// 秒は入れない
	tstr := t.Format("2006/01/02 15:04")
	return tstr
}

// WORKSテーブルのIDからサムネの保存場所を計算
func thumbPath(id int) string {
	idStr := fmt.Sprint(id)
	fpath := "./public/thumb/" + idStr + ".png"
	return fpath
}

// content.html、meta.htmlをリクエストの都度開く（Rootを使わない）ページで利用するハンドラ。
// 同じ処理になるのでハンドラ化。cacheはサイドバーの表示で使う。
func defalutLayoutHandler(tmpl *template.Template, cache *WorksCache) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		route := strings.Trim(path, "/")
		arr := strings.Split(route, "/")
		if len(arr) > 1 {
			// route/another-routeのようなパスの場合、第一要素をrouteとみなして処理
			route = arr[0]
		}

		rec := NewRootRecord(route)
		if rec.Err != nil {
			err := stack("defaultLayoutHandler", rec.Err)
			logErr(r, err)
			http.NotFound(w, r)
			return
		}
		rec.Sections = cache.SectionCache
		err := tmpl.ExecuteTemplate(w, "layout.html", rec)

		if err != nil {
			err = stack("defaultLayoutHandler", err)
			logErr(r, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}

func main() {
	// db接続
	db, err := sql.Open("sqlite3", DB)
	if err != nil {
		log.Fatal(err)
	}

	// サイドバーのリンクのキャッシュを初期化
	cache := &WorksCache{}
	err = cache.Update(db)
	if err != nil {
		log.Fatal(err)
	}

	// templateに埋め込むデータとか
	rootData := NewRootData()

	// template内で実行できるカスタム関数
	customFunc := template.FuncMap{
		"add": func(v, inc int) int {
			return v + inc
		},
		"isProduction": func() bool {
			// 本番かどうかチェックする。GA4のスクリプト制御用（本番時のみ入れたい）
			// OSがWindowsなら開発環境、以外なら本番とみなす。
			return runtime.GOOS != "windows"
		},
	}

	// html template
	// 共通レイアウト部分（"layout"）、works部分（"works"）,それ以外（"edit-description"）で分かれている。名前に特に意味はない
	// tmpl, err := template.ParseFiles("./html/edit-description.html", "./html/preview.html")
	tmpl, err := template.New("edit-description").Funcs(customFunc).ParseFiles("./html/edit-description.html", "./html/preview.html", "./html/deleted-works.html", "./html/works-list.html", "./html/board.html", "./html/thread.html")
	if err != nil {
		log.Fatal(err)
	}

	// html template for root page
	rootTmpl, err := template.New("layout").Funcs(customFunc).ParseFiles("./html/layout.html", "./html/nav.html", "./html/sidebar.html")
	if err != nil {
		log.Fatal(err)
	}

	// 詰将棋のコンテンツ部分のtemplate
	worksTmpl, err := template.New("works").Funcs(customFunc).ParseFiles("./html/works.html", "./html/posts.html")
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

	// site-policyページ等、一部の静的ページで使うハンドラ。
	staticPageHandler := defalutLayoutHandler(rootTmpl, cache)

	// routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			// ルートページを返す
			// 現在のページをハイライトするために指定。
			// HOME -> ルートページ、EDIT -> 編集ページ、BOARD -> 掲示板
			data, ok := rootData["home"]
			if !ok || data.Err != nil {
				logErr(r, err)
				http.NotFound(w, r)
				return
			}

			data.Sections = cache.SectionCache

			err := rootTmpl.ExecuteTemplate(w, "layout.html", data)

			if err != nil {
				logErr(r, err)
				w.WriteHeader(http.StatusInternalServerError)
			}
			// ↑でrootTmpl.ExecuteTemplate(w,~,~)でwに書き込んでいるので、ここでリターンする必要がある。そうしないとfs.ServeHTTP(w,r)がワーニングになる（superfluous response.WriteHeader call）
			return
		}
		// static folder（publicフォルダ）のデータを返す。CSSやJS等。
		fs.ServeHTTP(w, r)
	})

	http.Handle("/rule/", checkRoute("rule", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, ok := rootData["rule"]
		if !ok || data.Err != nil {
			logErr(r, err)
			http.NotFound(w, r)
			return
		}
		data.Sections = cache.SectionCache
		err := rootTmpl.ExecuteTemplate(w, "layout.html", data)
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	})))

	http.Handle("GET /policy/", checkRoute("policy", staticPageHandler))
	http.Handle("GET /browser-support/", checkRoute("browser-support", staticPageHandler))

	http.Handle("/edit/", checkRoute("edit", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := NewRootRecord("edit")
		if rec.Err != nil {
			logErr(r, err)
			http.NotFound(w, r)
			return
		}

		rec.Sections = cache.SectionCache
		rec.Current = "EDIT"

		err := rootTmpl.ExecuteTemplate(w, "layout.html", rec)

		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	})))

	// 掲示板ページ。スレッドの一覧を表示。スレッド自体は/threadエンドポイントなので留意。
	http.Handle("GET /board/", checkRoute("board", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		lastDateStr := query.Get("last-date")
		var lastDate int64
		if len(lastDateStr) == 0 {
			lastDate = 0
		} else {
			lastDate, err = strconv.ParseInt(lastDateStr, 10, 64)
			if err != nil {
				logErr(r, err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		threads, err := getThreads(db, BOARD_LIMIT, lastDate)

		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// 取得したデータの中で、もっとも若い投稿日で更新（次の表示用）
		if len(threads) == 0 {
			lastDate = 0
		} else {
			// 投稿がないスレは投稿日は0なので、nullでない最後の値をセット。だる、、、
			i := len(threads) - 1
			for i >= 0 {
				t := threads[i]
				if t.LastCommentDateUnix != 0 {
					lastDate = int64(t.LastCommentDateUnix)
					break
				}
				i--
			}
		}

		buf := &bytes.Buffer{}
		boardData := struct {
			Threads  []ThreadRecord
			LastDate int64
		}{
			Threads: threads, LastDate: lastDate,
		}

		err = tmpl.ExecuteTemplate(buf, "board.html", boardData)

		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		rec := Record{
			Meta: template.HTML(`
			<meta name="description" content="">
			<link rel="stylesheet" href="/css/board.css">
			<script src="/js/board/main.js" type="module"></script>
			<title>掲示板</title>
			`),
			Content:  template.HTML(buf.String()),
			Current:  "BOARD",
			Sections: cache.SectionCache,
		}

		err = rootTmpl.ExecuteTemplate(w, "layout.html", rec)

		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusInternalServerError)
		}

	})))

	// threadページで利用。/api/insert-commentでも使っていたが外した。。
	loadThread := func(id int, title string, w http.ResponseWriter) error {
		comments, err := getComments(db, int(id))
		if err != nil {
			err = stack("loadThread", err)
			return err
		}

		data := struct {
			ThreadId int
			Title    string
			Comments []CommentRecord
		}{
			id, title, comments,
		}
		buf := &bytes.Buffer{}
		err = tmpl.ExecuteTemplate(buf, "thread.html", data)
		if err != nil {
			err = stack("loadThread", err)
			return err
		}

		meta := fmt.Sprintf(`
			<meta name="description" content="スレッドにコメントを書き込む画面です。他人の誹謗中傷や、トピックの関係の話題は禁止です。">
			<link rel="stylesheet" href="/css/thread.css">
			<script src="/js/thread/main.js" type="module"></script>
			<title>【スレッド】%v</title>
		`, title)
		rec := Record{
			Meta:     template.HTML(meta),
			Content:  template.HTML(buf.String()),
			Current:  "BOARD",
			Sections: cache.SectionCache,
		}

		err = rootTmpl.ExecuteTemplate(w, "layout.html", rec)

		if err != nil {
			err = stack("loadThread", err)
		}
		return err
	}

	// threadページ
	http.HandleFunc("GET /thread/", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		idStr := query.Get("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		title := query.Get("title")
		err = loadThread(int(id), title, w)
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	})

	// 新規作成画面
	http.HandleFunc("/edit/description/", func(w http.ResponseWriter, r *http.Request) {
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
			logErr(r, err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	})

	// 修正画面。edit/descriptionのテンプレートを利用
	http.HandleFunc("/revise/", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		idStr := query.Get("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		rec, err := getWork(db, int(id))
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
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
			logErr(r, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	http.HandleFunc("POST /preview/", func(w http.ResponseWriter, r *http.Request) {
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
				logErr(r, err)
				w.WriteHeader(http.StatusBadRequest)
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
			logErr(r, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// preview.htmlテンプレートには３つのモード(revise,new,undo,restore)がある。
		// /previewエンドポイントでは、revise、newのいずれかとなる。
		// /undoエンドポイントでもpreview.htmlを使い、一律undoとなる
		// /restoreエンドポイントでもpreview.htmlを使い、一律restoreとなる
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
			logErr(r, err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	})

	http.HandleFunc("POST /api/insert-work", func(w http.ResponseWriter, r *http.Request) {
		body := &WorkBody{}
		dec := json.NewDecoder(r.Body)
		dec.Decode(body)

		maxId, err := insertWork(db, body)
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// dataURLの"data:image/png;base64,"部分を切り取るために、コンマでsplit
		dataArr := strings.Split(body.Pic, ",")
		if len(dataArr) != 2 {
			fmt.Println("画像のdataURL部分が不正です。リクエスト内の画像は以下のとおり:")
			fmt.Println(body.Pic)
			w.Write([]byte("サムネの登録ができませんでしたが、作品登録は問題なく出来ました。ご協力ありがとうございました。ホーム画面に戻ります。"))
			return
		}

		// base64部部のみを抽出し、デコード
		pic := dataArr[1]
		img, err := base64.StdEncoding.DecodeString(pic)
		if err != nil {
			logErr(r, err)
			fmt.Println("画像をbase64からデコードできませんでした。リクエスト内の画像は以下のとおり:")
			fmt.Println(body.Pic)
			w.Write([]byte("サムネの登録ができませんでしたが、作品登録は問題なく出来ました。ご協力ありがとうございました。ホーム画面に戻ります。"))
			return
		}

		fpath := thumbPath(maxId)
		f, err := os.Create(fpath)
		if err != nil {
			logErr(r, err)
			fmt.Println("ファイルの保存先が取得できませんでした。保存先パス：")
			fmt.Println(fpath)
			w.Write([]byte("サムネの登録ができませんでしたが、作品登録は問題なく出来ました。ご協力ありがとうございました。ホーム画面に戻ります。"))
			return
		}
		defer f.Close()

		_, err = f.Write(img)
		if err != nil {
			logErr(r, err)
			fmt.Println("ファイルの書き込みが出来ませんでした")
			w.Write([]byte("サムネの登録ができませんでしたが、作品登録は問題なく出来ました。ご協力ありがとうございました。ホーム画面に戻ります。"))
			return
		}

		w.Write([]byte("登録成功しました。反映まで少し時間がかかる場合があります。作品投稿ありがとうございましたm(__)m。ホーム画面に戻ります。"))

		// リンクのcacheを更新
		err = cache.Update(db)
		if err != nil {
			logErr(r, err)
		}
	})

	http.HandleFunc("/api/update-work", func(w http.ResponseWriter, r *http.Request) {
		body := &WorkBody{}
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(body)
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		//  - backup and update work db
		err = updateWork(db, body)
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write([]byte("更新成功しました。反映まで少し時間がかかる場合があります。編集ありがとうございましたm(__)m。ホーム画面に戻ります。"))

		err = cache.Update(db)
		if err != nil {
			logErr(r, err)
		}
	})

	// 作品評価のアイコン（グッドやバッド等）を押したときの処理
	// DBを更新し、更新後の値を返す
	http.HandleFunc("POST /api/update-eval", func(w http.ResponseWriter, r *http.Request) {
		data := &UpdateEvalBody{}
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(data)
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		cnt, err := updateWorkEval(db, data)
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", "text/plain")
		w.Write([]byte(fmt.Sprint(cnt)))
	})

	// worksのコメントを投稿する時に呼び出されるAPI
	http.HandleFunc("POST /api/insert-post", func(w http.ResponseWriter, r *http.Request) {
		name := r.FormValue("name")
		comment := r.FormValue("comment")
		idStr := r.FormValue("id")
		commentType := r.FormValue("type")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		post, err := insertPost(db, name, comment, commentType, int(id))
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		postsData := map[string][]PostRecord{
			"Posts": {post},
		}

		err = worksTmpl.ExecuteTemplate(w, "posts", postsData)
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusInternalServerError)
		}

	})

	http.Handle("GET /works/", checkRoute("works", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		paramid := query.Get("id")

		id, err := strconv.ParseInt(paramid, 10, 64)

		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// dbからデータ取得
		wr, err := getWork(db, int(id))
		if err != nil {
			logErr(r, err)
			http.NotFound(w, r)
			return
		}

		// dbから投稿（post）を取得
		posts, err := getPosts(db, int(id), 0, POST_LIMIT)
		wr.Posts = posts
		// エラーが出ても取得しないだけなので、エラー表示のみとし、returnはしない。
		if err != nil {
			logErr(r, err)
		}

		// works.htmlテンプレートに埋め込み
		wBuf := &bytes.Buffer{}
		err = worksTmpl.ExecuteTemplate(wBuf, "works.html", wr)

		if err != nil {
			logErr(r, err)
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
			logErr(r, err)
		}
	})))

	// 作品のコメント投稿内の評価アイコン（グッドやバッド等）を押した時の処理
	// 評価を更新し、更新後の値を返す
	http.HandleFunc("POST /api/update-post-eval", func(w http.ResponseWriter, r *http.Request) {
		body := &UpdatePostEvalBody{}
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(body)
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		cnt, err := updatePostEval(db, body)
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", "text/plain")
		_, err = w.Write([]byte(fmt.Sprint(cnt)))
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	})

	// 作品のコメント投稿の次の一覧を取得する処理
	http.HandleFunc("GET /api/get-next-posts", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		idStr := query.Get("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		lastSeqStr := query.Get("seq")
		lastSeq, err := strconv.ParseInt(lastSeqStr, 10, 64)
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		recs, err := getPosts(db, int(id), int(lastSeq), POST_LIMIT)

		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// recsの長さがゼロなら、全て取得したことをクライアントに伝える
		if int(len(recs)) == 0 {
			// 204 No-Contentを返す
			w.WriteHeader(http.StatusNoContent)
			return
		}

		postsData := map[string][]PostRecord{
			"Posts": recs,
		}

		// {{define template "name"}}で宣言したテンプレートは、その名前を使う必要がある点に注意。ファイル名じゃダメ
		err = worksTmpl.ExecuteTemplate(w, "posts", postsData)

		if err != nil {
			logErr(r, err)
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
			logErr(r, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		nextId, err := getNextWork(db, int(id), int(tesu), int(value))
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err = w.Write([]byte(fmt.Sprint(nextId)))
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	})

	// undoモードでプレビュー画面を表示する
	http.HandleFunc("GET /undo/", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		idStr := query.Get("id")
		seqStr := query.Get("seq")

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		seq, err := strconv.ParseInt(seqStr, 10, 64)
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		hist, err := getHistory(db, int(id), int(seq))

		if err != nil {
			logErr(r, err)
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

		wBuf := &bytes.Buffer{}
		err = worksTmpl.ExecuteTemplate(wBuf, "works.html", wc)
		if err != nil {
			logErr(r, err)
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
			logErr(r, err)
		}
	})

	// preview画面（undoモード）の登録処理。workの内容をhistoryから戻す
	http.HandleFunc("/api/update-undo", func(w http.ResponseWriter, r *http.Request) {
		bd := &UpdateUndoBody{}
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(bd)
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = undoWorkFromHistory(db, bd.Id, bd.Seq)
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err = w.Write([]byte("更新成功"))
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = cache.Update(db)
		if err != nil {
			logErr(r, err)
		}
	})

	// restoreモードでpreview.htmlを表示する
	// @TODO /undoとだいぶ被るので共通化できるところはしたほうが良いかも。余力のあるときに。
	http.HandleFunc("GET /restore/", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		idStr := query.Get("id")
		id64, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		id := int(id64)

		wr, err := getWork(db, id)
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// preview.htmlに渡すデータ（WorksContent型）を作成
		form := EditFormData{
			WorkId:      idStr,
			Explanation: wr.Explanation,
			Author:      wr.Author,
			Editor:      wr.Editor,
			Title:       wr.Title,
		}

		dates := Dates{
			PublishDate: wr.PublishDate,
			EditDate:    wr.EditDate,
		}

		wc := &WorksContent{
			EditFormData: form,
			Dates:        dates,
			Main:         wr.Main,
			Tegoma:       wr.Tegoma,
			GoteTegoma:   wr.GoteTegoma,
			Kihu:         wr.Kihu,
			KihuJ:        wr.KihuJ,
			IsPreview:    true,
			Tesu:         wr.Tesu,
		}

		wBuf := &bytes.Buffer{}
		err = worksTmpl.ExecuteTemplate(wBuf, "works.html", wc)
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		wdata := struct {
			Mode    string
			Content template.HTML
		}{
			"restore",
			template.HTML(wBuf.String()),
		}

		err = tmpl.ExecuteTemplate(w, "preview.html", wdata)
		if err != nil {
			logErr(r, err)
		}
	})

	http.HandleFunc("POST /api/restore", func(w http.ResponseWriter, r *http.Request) {
		var bd struct {
			Id int `json:"id"`
		}
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&bd)
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusBadRequest)
		}

		err = restoreWork(db, bd.Id)
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err = w.Write([]byte("更新成功しました。反映まで少し時間がかかる場合があります。ホーム画面に戻ります。"))
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = cache.Update(db)
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	http.HandleFunc("GET /deleted-works/", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		// 画面表示されている削除一覧のうち、最小の削除日付がパラメタとして渡される
		lastDateStr := query.Get("last-date")

		var lastDate int64
		var err error

		if len(lastDateStr) == 0 {
			// パラメタがない場合、最小日付は0にする。0の場合、template側で初期表示とみなす。
			lastDate = 0
		} else {
			lastDate, err = strconv.ParseInt(lastDateStr, 10, 64)
			if err != nil {
				logErr(r, err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		// deleteされた作品の一覧。lastDateより前のデータが抽出される。
		// 0の場合、関数側で初期表示とみなす。
		wk, err := getDeletedWorks(db, lastDate)
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		buf := &bytes.Buffer{}
		// 取得した
		nextLastDateInt := wk.Min()
		delData := struct {
			Data        DeletedWorks
			LastDateInt int
		}{
			wk, nextLastDateInt,
		}

		err = tmpl.ExecuteTemplate(buf, "deleted-works.html", delData)
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		wdata := Record{
			Meta: template.HTML(`
			<meta name="description" content="削除作品の一覧です。削除された作品を元に戻すことができます。">
			<link rel="stylesheet" href="/css/deleted-works.css">
			<script src="/js/deleted-works/main.js" type="module"></script>
			<title>削除作品一覧</title>
			`),
			Content:  template.HTML(buf.String()),
			Sections: cache.SectionCache,
		}

		err = rootTmpl.ExecuteTemplate(w, "layout.html", wdata)
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	// workを削除するときに呼ばれるエンドポイント
	http.HandleFunc("POST /api/delete", func(w http.ResponseWriter, r *http.Request) {
		dec := json.NewDecoder(r.Body)
		data := struct {
			Id     int    `json:"id"`
			Editor string `json:"editor"`
			Reason string `json:"reason"`
		}{}
		err := dec.Decode(&data)
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = deleteWork(db, data.Id, data.Editor, data.Reason)
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err = w.Write([]byte("削除成功しました。"))
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = cache.Update(db)
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	// 作品一覧ページ
	http.HandleFunc("GET /works-list/", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		// 手数。0なら全量
		tesuStr := query.Get("tesu")
		// 手数指定の際に利用。表示作品には上限があるため、次の作品を取得できるように次の検索開始位置を保持
		startStr := query.Get("start")
		var tesu, start int64

		if len(tesuStr) == 0 {
			// tesuのクエリパラメタが無い場合、全量（0）とする
			tesu = 0
		} else {
			tesu, err = strconv.ParseInt(tesuStr, 10, 64)
			if err != nil {
				logErr(r, err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		if len(startStr) == 0 {
			start = 0
		} else {
			start, err = strconv.ParseInt(startStr, 10, 64)
			if err != nil {
				logErr(r, err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		wlist, err := getWorksList(db, int(start), int(tesu))
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// 手数指定の場合、（次のデータを取れるよう）取得したリストの中から
		// 最大のIDを設定する
		var lastId int
		if len(wlist) == 1 {
			lastId = wlist[0].Max()
		}

		wbuf := &bytes.Buffer{}
		wkTmpl := struct {
			IsSingle bool
			LastId   int
			List     []WorkListTmpl
		}{
			IsSingle: len(wlist) == 1,
			LastId:   lastId,
			List:     wlist,
		}
		err = tmpl.ExecuteTemplate(wbuf, "works-list.html", wkTmpl)
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		data := Record{
			Meta: `
			<meta name="description" content="投稿作品の一覧ページです。">
			<link rel="stylesheet" href="/css/works-list.css">
			<script src="/js/works-list/main.js" type="module"></script>
			<title>作品一覧</title>
			`,
			Content:  template.HTML(wbuf.String()),
			Sections: cache.SectionCache,
		}

		err = rootTmpl.ExecuteTemplate(w, "layout.html", data)
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	http.HandleFunc("POST /api/create-thread", func(w http.ResponseWriter, r *http.Request) {
		name := r.FormValue("name")
		title := r.FormValue("title")
		// maxIdはThreadsの最大ID（挿入されたレコードのIDとなる）
		maxId, err := insertThread(db, name, title)
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		url := fmt.Sprintf("/thread?id=%v", maxId)

		http.Redirect(w, r, url, http.StatusSeeOther)
	})

	// /threadページでコメントを投稿
	http.HandleFunc("POST /api/insert-comment", func(w http.ResponseWriter, r *http.Request) {
		author := r.FormValue("name")
		comment := r.FormValue("comment")

		threadIdStr := r.FormValue("thread-id")
		threadId, err := strconv.ParseInt(threadIdStr, 10, 64)
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		replyStr := r.FormValue("reply-to")
		reply, err := strconv.ParseInt(replyStr, 10, 64)
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		title := r.FormValue("title")

		err = insertComment(db, int(threadId), int(reply), author, comment)
		if err != nil {
			logErr(r, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		url := fmt.Sprintf("/thread/?id=%v&title=%v", threadId, title)

		// クライアントに最新投稿を表示するためリダイレクト
		http.Redirect(w, r, url, http.StatusSeeOther)
	})

	// 準備中の時に表示するページ
	http.Handle("GET /not-ready/", staticPageHandler)

	http.HandleFunc("/about-me/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/not-ready/", http.StatusSeeOther)
	})

	http.HandleFunc("/blog/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/not-ready/", http.StatusSeeOther)
	})

	// localhostをつけないと、起動時にfw許可のメッセージが出る
	// つけると、スマホ等別デバイスからのアクセスができなくなる
	// err = http.ListenAndServe("localhost:8000", nil)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	http.ListenAndServe(":8000", nil)
}
