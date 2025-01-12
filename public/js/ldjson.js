/**
 * 【概要】
 * LD+JSONでBreadCrumbsを作成
 * https://developers.google.com/search/docs/appearance/structured-data/breadcrumb?hl=ja
 * 【疑問】
 * 　　最後のitemListElementに、`item`プロパティは必要？
 * 　　（上のリンクでは省略されているが、ChatGPTは必要といってる）
 * 
 * 【canonical path】 *ni:no-index
 *  - home: home
 *  - rule: home > rule
 *  - board: home > board
 *  - thread: home > board > thread
 *  - policy: home > policy
 *  - browser-support: home > browser-support
 *  - deleted-works: home > deleted-works
 *  - [ni]restore: home > deleted-works > restore
 *  - works-list: home > works-list
 *  - works: home > works-list > works
 *  - [ni]revise: home > works-list > works > revise
 *  - [ni]undo: home > > works-list > works > undo
 *  - edit: home > edit
 *  - [ni]description: home > edit > description
 *  - [ni]preview: home > edit > description > preview
 */
(function () {
    function genLdJson(paths) {

        if (paths === null || paths === undefined) {
            return null;
        }

        // 最後のパスをもとに判定する
        const lastPath = paths[paths.length - 1];

        // インデックス登録をしないページは何もしない
        if (lastPath === "revise" || lastPath === "restore"
            || lastPath === "undo" || lastPath === "undo"
        ) {
            return null;
        }

        const origin = "https://tsume-shougi-wiki.com/";

        const genItem = (i, name, url) => {
            const ret = {
                "@type": "ListItem",
                "position": i,
                "name": name,
            };

            // 最後の要素は、itemが無くても良いみたいなので（↑のURL参照）
            if (url) {
                ret["item"] = url;
            }

            return ret;
        };

        const breadCrumbList = {
            "@context": "https://schema.org",
            "@type": "BreadcrumbList",
            "itemListElement": [],
        };

        // homeは必ず入るのでセット
        const rootItem = genItem(1, "home", origin);
        breadCrumbList.itemListElement.push(rootItem);

        // "/"の場合はホーム画面
        if (paths.length === 0) {
            return breadCrumbList;
        }

        // 例外処理:works
        if (lastPath === "works") {
            let url = origin + "works-list/";
            let item = genItem(2, "作品一覧", url);
            breadCrumbList.itemListElement.push(item);
            item = genItem(3, "作品", null);
            breadCrumbList.itemListElement.push(item);
            return breadCrumbList;
        }

        // 例外処理:thread
        if (lastPath === "thread") {
            let url = origin + "board/";
            let item = genItem(2, "掲示板", url);
            breadCrumbList.itemListElement.push(item);
            item = genItem(3, "スレッド", null);
            breadCrumbList.itemListElement.push(item);
            return breadCrumbList;
        }

        // 通常
        // 上記除外ページは含まず
        const nameResolver = {
            "rule": "ルール",
            "board": "掲示板",
            "policy": "サイトポリシー",
            "browser-support": "動作環境",
            "deleted-works": "削除一覧",
            "works-list": "作品一覧",
            "edit": "編集",
        };
        // let url = origin + lastPath + "/";
        let name = nameResolver[lastPath];
        let item = genItem(2, name, null);
        breadCrumbList.itemListElement.push(item);
        return breadCrumbList;

    }

    const script = document.createElement("script");
    script.setAttribute("type", "application/ld+json");
    const url = new URL(window.location.href);
    const paths = url.pathname.split("/").filter(v => v.length > 0);

    const ld = genLdJson(paths)

    // pathsが何らかの理由で取得できない場合はld+jsonを生成しないでリターン
    if (ld === null) {
        script = null;
        return;
    }
    script.textContent = JSON.stringify(ld);
    document.head.appendChild(script);
})();