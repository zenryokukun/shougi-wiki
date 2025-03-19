# 詰将棋WIKI

## TODOs
- line-height調整。zenだと1.9にしているみたい。⇒どの要素に対してつけるか？（body?）は要検討
- デプロイ後に対応：　ogp対応（ルートページしか対応していない）。
- certbotのrenew対応。3か月で失効するため。cron使うみたい。certbotのHP参考に。自動で更新かかる説もあるので、様子見か。。
- 追加予定ページ（優先順位）：管理人について、ブログ
- NewRootData（キャッシュ）とNewRootRecord（オンデマンドで開く）のどっち使うかはっきりさせる
- defaultLayoutHandler関数の用途がよく分からないので精査
　（/edit/や/rule/でも使えそうな気がするものの、、、ナビゲーションにないページの場合のみに使用？）
- innerHTMLにdompurify導入
- bread-crumb（画面上部のナビゲーション）の追加検討
- 名前入力のautocomplete
- main.goのNewRootData -> 都度ファイルを開いて返す形にすることも検討。レスポンスが課題になるようであればキャッシュで良い気が、、
- esbuildとかのバンドラ導入（リリース後でも）

## 技術スタック

- vanilla js
- html/template(GOのビルドインのtemplateエンジン)
- net/http
- sqlite3 

## フォルダ構成

### public

staticフォルダ。cssは```/public/css```、jsは```/public/js```フォルダへ。

svgは一応```/svg```フォルダへ。基本は用途別のフォルダに入れる。

各フォルダ内では、route単位にフォルダを分けること。例えば、```https://tsume-shogi-wiki/rule```のcssは```/public/css/rule```、jsは```/public/js/rule```に入れること。

## ファイル名のルール

ビルド・ツール導入を見据え（？）、各routeをエントリーポイントとし、cssは```main.css```、jsは```main.js```の名称にする。

共通モジュールはここからimportして使うこと。

### html

templateエンジンで使うhtmlテンプレート

## TODOS

### must

- ビルドツールの導入。esbuildとか。

### might

- layout.htmlの右サイドバーを追加しても良いかもしれない。更新情報、作者情報など。asideタグとかで。コンテンツがある程度増えてからでも良いかもだけど。
- AWSでデプロイ

## deploy

サーバ側での実行

### アプリ起動

```bash
# 適宜git pullしておく
# 最後の&はバックグラウンドで実行するために必要。SSHが切れても動かせ続けるために、nohupコマンドを使う
nohup ./executable &
```

### アプリ停止

```bash
# process idを確認
sudo ps aux | grep "executable"

# そのidをタスクキル
kill id-number
```