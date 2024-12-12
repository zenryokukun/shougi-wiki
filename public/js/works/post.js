// 投稿内のgood,bad評価をする。
// 「コメント」（eval.js）とは別なので注意
import getParamId from "./param.js";

export default function initPostEval() {
    const goods = document.querySelectorAll('[data-type="good"]');
    const bads = document.querySelectorAll('[data-type="bad"]');
    const pid = getParamId();
    for (let i = 0; i < goods.length; i++) {
        const good = goods[i];
        const bad = bads[i];
        good.addEventListener("click", () => click(pid, good, bad));
        bad.addEventListener("click", () => click(pid, bad, good));
    }
}

// 「次のpost」を取得した歳、新たに取得したpostにイベントをつける関数
export function initNewPostEval(lastSeq) {
    const pid = getParamId();
    const nodes = document.querySelectorAll(`[data-seq]`);
    for (const node of nodes) {
        const seq = node.dataset.seq;
        const seqInt = parseInt(seq);
        const lastSeqInt = parseInt(lastSeq);
        // lastSeqまではイベントがアタッチされているので何もしない
        if (seqInt >= lastSeqInt) { continue; }
        const good = node.querySelector(`[data-type="good"]`)
        const bad = node.querySelector(`[data-type="bad"]`)
        good.addEventListener("click", () => click(pid, good, bad));
        bad.addEventListener("click", () => click(pid, bad, good));
    }
}

function click(pid, node, otherNode) {
    const svg = node.querySelector("svg");
    const otherSvg = otherNode.querySelector("svg");
    const tokens = svg.classList;
    const seqNode = node.closest("[data-seq]");
    const seq = seqNode.dataset.seq;
    const countNode = node.parentElement.querySelector("span");
    let value = 0;
    const evalType = node.dataset.type;

    if (tokens.contains("on")) {
        // on -> off
        tokens.remove("on");
        tokens.add("off");
        value = -1;
    } else if (tokens.contains("off")) {
        if (otherSvg.classList.contains("on")) {
            // 既に逆の評価がおされている場合は何もしない。
            // goodとbad両方押された状態にしたくないため。
            return;
        }
        // off -> on
        tokens.remove("off");
        tokens.add("on");
        value = 1;
    }
    submit(
        parseInt(pid),
        parseInt(seq),
        value,
        evalType.toUpperCase(),
        countNode
    );
}

function submit(pid, seq, value, key, targetNode) {
    const param = { id: pid, seq, key, value };
    fetch("/api/update-post-eval", {
        method: "post",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(param),
    }).then(res => {
        if (!res.ok) {
            throw new Error("更新失敗。後で試してください。")
        }
        return res.text();
    }).then(val => {
        targetNode.textContent = val;
    }).catch(err => {
        alert(err);
    });
}