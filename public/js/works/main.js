import Canvas from "../shougi/canvas.js";
import GameHistory from "../shougi/game-history.js";
import initIcon from "../shougi/icon.js";
import initToggle from "../answer.js";
import initEval from "./eval.js";
import initPostEval, { initNewPostEval } from "./post.js";
import getParamId from "./param.js";
import { showModal } from "./modal.js";

function init() {
    const cvs = document.querySelector("canvas");
    const sprite = document.querySelector("#sprite");
    const mainStr = cvs.dataset["main"];
    const tegomaStr = cvs.dataset["tegoma"];
    const main = JSON.parse(mainStr);
    const tegoma = JSON.parse(tegomaStr);
    const game = new GameHistory(main, tegoma);
    const mycvs = new Canvas(cvs, sprite);
    initIcon(game, mycvs);
    initToggle();
    mycvs.render(game);
    initEval();
    initPostEval();
    initSubmit();
    initGetNextPost();
    initOtherWork();
    initRevise();
    const trash = document.querySelector(".delete")
    trash.addEventListener("click", showModal);
}

/**「編集する」を押したときの処理 */
function initRevise() {
    const lnk = document.querySelector(".revise");
    lnk.addEventListener("click", e => {
        const ans = confirm("編集画面に進みます。よろしいですか？")
        if (!ans) {
            e.preventDefault();
            return;
        }
    });
}

/**「前の作品」「次の作品」を押したときの処理 */
function initOtherWork() {
    const next = document.getElementById("next");
    if (next === null) return;
    const id = getParamId();
    const tesuNode = document.querySelector("[data-tesu]");
    const tesu = tesuNode.dataset.tesu;

    function customFetch(node, param) {
        if (node.classList.contains("posts-fetch-disabled")) return;
        fetch("/api/next-work?" + param.toString())
            .then(res => {
                if (!res.ok) {
                    throw new Error("次のデータがありません")
                }
                return res.text();
            })
            .then(value => window.location.href = "/works/?id=" + value)
            .catch(err => {
                // paramはuRLSearchParams型。getで値をし、取得した値はstringとなる点に注意
                const v = param.get("value");
                const which = v === "1" ? "次" : "前";
                const msg = `この手数では、${which}の作品はありません。`;
                alert(msg);
                node.classList.add("posts-fetch-disabled");
            })
    }

    next.addEventListener("click", e => {
        e.preventDefault();
        const param = new URLSearchParams({ id, tesu, value: 1 })
        customFetch(next, param);
    })

    const prev = document.getElementById("prev");
    // prevはnullの場合もある
    if (prev === null) return;
    prev.addEventListener("click", e => {
        e.preventDefault();
        const param = new URLSearchParams({ id, tesu, value: -1 })
        customFetch(prev, param);
    });
}

/**次の投稿内容を取得する*/
function initGetNextPost() {

    const parent = document.querySelector(".posts");
    // コメントが無い場合は.postsのHTML要素は存在しない。
    // その場合は何もせずリターン
    if (parent === null) return;
    const node = document.querySelector(".posts-fetch");
    const id = getParamId();
    node.addEventListener("click", e => {
        const lastNode = document.querySelector(".post:last-of-type");
        const nodeWithSeq = lastNode.querySelector("[data-seq]");
        const seq = nodeWithSeq.dataset.seq;

        const params = new URLSearchParams({ id, seq });
        fetch("/api/get-next-posts?" + params.toString())
            .then(res => {
                if (!res.ok) {
                    throw new Error("投稿内容が取得できませんでした");
                }
                if (res.status === 204) {
                    node.disabled = true;
                    node.classList.add("posts-fetch-disabled");
                    node.textContent = "全て取得済み";
                }
                return res.text();
            })
            .then(txt => {
                // innerHTMLだと、配下の要素につけたイベントがリセットされる。insertAdjacentHTMLだと消えない
                parent.insertAdjacentHTML("beforeend", txt);
                // 追加されたpostにイベント等を設定
                initNewPostEval(seq);
            })
            .catch(err => alert(err))
    });
}

/**内部関数　initSubmitで利用 */
function getEval() {
    const good = document.querySelector("#good .on");
    if (good) return "good";
    const bad = document.querySelector("#bad .on");
    if (bad) return "bad";
    const demand = document.querySelector("#demand .on");
    if (demand) return "demand";
    return null;
}

/**コメントを投稿するときの処理 */
function initSubmit() {
    const form = document.querySelector(".comment-form");
    // urlのquery paramのidを取得
    const id = getParamId()
    form.addEventListener("submit", e => {
        e.preventDefault();
        const typeEval = getEval();
        if (typeEval === null) {
            alert("コメントを投稿するには、いずれかの評価のアイコンを選択してください。");
            return
        }
        const fd = new FormData(form);
        // idをformdataに追加
        fd.append("id", id);
        // クリックしたアイコンをformdataに追加
        fd.append("type", typeEval);
        //  fetchでFormDataを渡すときはheadersは指定しない
        fetch("/api/insert-post", {
            method: "post",
            body: fd,
        }).then(res => {
            if (!res.ok) {
                throw new Error("投稿失敗");
            }
            return res.text();
        }).then(html => {
            // 投稿ステータスを表示し、ボタンを押せなくする（連投禁止）
            const node = document.querySelector(".submit-status");
            node.textContent = "投稿成功";
            const btn = document.querySelector(".submit");
            btn.disabled = true;
            // 投稿したコメントを表示
            const parent = document.querySelector(".posts");
            if (parent === null) return;
            // innerHTMLだとイベント消えるのでこっち使う
            parent.insertAdjacentHTML("afterbegin", html);

            // コメントがない場合は「次のコメントを取得」のボタンが隠れているので、表示させる
            const pfetch = document.querySelector(".posts-fetch")
            if (pfetch === null) return;
            pfetch.classList.remove("hidden");

        }).catch(err => {
            const node = document.querySelector(".submit-status");
            node.textContent = err;
        });
    })
}

window.addEventListener("load", init);