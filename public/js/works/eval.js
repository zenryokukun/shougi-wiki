import getParamId from "./param.js";

let isErr = false;

export default function init() {
    const btns = document.querySelectorAll(".eval-wrapper button");
    for (let i = 0; i < btns.length; i++) {
        btns[i].addEventListener("click", e => { click(i, btns) });
    }
}

function click(at, btns) {

    if (isErr) {
        alert("エラーのため評価の反映ができません。時間を空けて試してください。");
        return;
    }

    for (let i = 0; i < btns.length; i++) {
        if (i === at) continue;
        const btn = btns[i];
        // on offのクラスをつけるのはbuttonタグでなく子要素のsvgタグ
        const svg = document.querySelector(`#${btn.id} svg`);
        // 自分以外のボタンが押されている場合は何もしない
        if (svg.classList.contains("on")) {
            return;
        }
    }
    // 自分以外のボタンが押されていない状態
    // on->off, off->onで処理分ける
    let value = 0;
    const btn = btns[at];
    const svg = document.querySelector(`#${btn.id} svg`);
    const tokens = svg.classList;
    if (tokens.contains("on")) {
        // on->off
        tokens.remove("on");
        tokens.add("off");
        value = -1;
    } else {
        tokens.remove("off");
        tokens.add("on");
        value = 1;
    }

    // 作品IDを取得。ページのqueryに?id=1のように設定されている
    const workId = getParamId();
    const cntNode = document.querySelector(`#${btn.id}+.count`)
    submit(parseInt(workId), btn.id.toUpperCase(), value, cntNode)
}

function submit(id, key, value, targetNode) {
    const data = { id, key, value };
    fetch("/api/update-eval", {
        method: "post",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(data),
    })
        .then(res => {
            if (!res.ok) {
                throw new Error(res.statusText);
            }
            return res.text()
        })
        .then(txt => {
            targetNode.textContent = txt
        })
        .catch(err => {
            alert("正しく評価の反映ができませんでした。時間を空けて試してください。");
            isErr = true;
        });
}