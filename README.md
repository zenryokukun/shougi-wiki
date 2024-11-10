# 詰将棋WIKI

## TODOs

- /deleted-works,/works-listページのリンクの置き場所を考える（サイドバーかフッター？）
- サイドバーに/works-listに遷移するリンクをつける
- 入力欄の改行を反映れるようにする（<br>に置換、もしくは<p>&nbsp</p>に置換）
- サイドバーの作品一覧リンクの抽出方法を検討（現状全てなので）
- 追加予定ページ（優先順位）：掲示板、管理人について、
- main.goのNewRootData -> 都度ファイルを開いて返す形にすることも検討。レスポンスが課題になるようであればキャッシュで良い気が、、

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
