function getParamTesu() {
    const param = new URLSearchParams(window.location.search);
    return param.get("tesu");
}

function main() {
    // 単一の手数作品の表示上限がある。次の一覧を取得する処理
    const lnk = document.querySelector("#next-list");
    // 表示されている作品で、最大のID
    const idNode = document.querySelector("[data-lastid]");
    const idStr = idNode.dataset.lastid
    const lastId = parseInt(idStr);
    const tesuStr = getParamTesu();
    let tesu = parseInt(tesuStr);
    // 単一作品の場合は手数指定されているので、想定はないが、tesuクエリパラメタが取得できず
    // NaNとなった場合、0（全量）を便宜上設定
    if (isNaN(tesu)) {
        tesu = 0;
    }

    lnk.addEventListener("click", () => {
        const nextId = lastId + 1;
        const url = `/works-list/?tesu=${tesu}&start=${nextId}`;
        // .disabledクラスが付いている場合、次の作品はもうないとみなす
        if (lnk.classList.contains("disabled")) return;
        fetch(url)
            .then(res => {
                /**
                 * @todo エラー制御はサーバのほうでももっとしたほうが、、、
                 */
                if (!res.ok) throw new Error();
                window.location.href = url;
            })
            .catch(() => {
                // anchorタグはdisabled属性がないので、CSSで無効化された見た目にする
                lnk.classList.add("disabled");
                alert("次の作品はありません");
            });
    });
}

window.addEventListener("DOMContentLoaded", main);