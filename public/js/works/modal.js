/**
 * 削除アイコンを押したときに表示するモーダル
 * svgタグとpathタグはcreateElementでなくcreateElementNSが必要
 */
import getParamId from "./param.js";

export function showModal() {
    /**
     * モーダルの表示/非表示で毎度nodeを生成、削除するのは面倒なので、
     * 生成は初回のみとして、以降はcssで表示⇔非表示を切り替える方式にする
     * .modalクラスのnodeが存在すれば、生成済みとみなす
     */
    const _modal = document.querySelector(".modal");
    if (_modal !== null) {
        // 生成済みの場合、cssで"表示"に切り替える
        // もともとdisplay:flexなので、blockでなくflexにしてあげる
        _modal.style.display = "flex";
        return;
    }

    // modalはbodyにappendする
    const body = document.querySelector("body");
    // modalのラッパ
    const modal = document.createElement("div");
    modal.className = "modal";
    // svg,pathはcreateElementNSで生成し、setAttributeで属性を設定する必要あり。
    const svg = document.createElementNS("http://www.w3.org/2000/svg", "svg");
    svg.setAttribute("xmlns", "http://www.w3.org/2000/svg");
    svg.setAttribute("viewBox", "0 0 512 512");
    const path = document.createElementNS("http://www.w3.org/2000/svg", "path");
    path.setAttribute("d", "M64 80c-8.8 0-16 7.2-16 16l0 320c0 8.8 7.2 16 16 16l384 0c8.8 0 16-7.2 16-16l0-320c0-8.8-7.2-16-16-16L64 80zM0 96C0 60.7 28.7 32 64 32l384 0c35.3 0 64 28.7 64 64l0 320c0 35.3-28.7 64-64 64L64 480c-35.3 0-64-28.7-64-64L0 96zm175 79c9.4-9.4 24.6-9.4 33.9 0l47 47 47-47c9.4-9.4 24.6-9.4 33.9 0s9.4 24.6 0 33.9l-47 47 47 47c9.4 9.4 9.4 24.6 0 33.9s-24.6 9.4-33.9 0l-47-47-47 47c-9.4 9.4-24.6 9.4-33.9 0s-9.4-24.6 0-33.9l47-47-47-47c-9.4-9.4-9.4-24.6 0-33.9z");
    svg.append(path);
    // 入力form
    const form = document.createElement("form");
    form.action = "/api/delete";
    form.method = "post";
    // 編集者のラベル、ラッパ、input
    const lblEditor = document.createElement("label");
    lblEditor.htmlFor = "editor";
    lblEditor.textContent = "名前";
    const wrapperEditor = document.createElement("div");
    const editor = document.createElement("input");
    editor.id = "editor";
    editor.name = "editor";
    editor.placeholder = "名前(30文字以内)";
    editor.maxLength = 30;
    editor.required = true;
    // 削除理由のラベル、ラッパ、input
    const wrapperReason = document.createElement("div");
    const lblReason = document.createElement("label");
    lblReason.htmlFor = "reason";
    lblReason.textContent = "理由";
    const reason = document.createElement("textarea");
    reason.id = "reason";
    reason.name = "reason";
    reason.placeholder = "削除理由(100文字以内)";
    reason.maxLength = 100;
    reason.required = true;
    // 送信ボタン
    const submit = document.createElement("button");
    submit.type = "submit";
    submit.textContent = "削除";
    // documentに生成したnode達を追加
    wrapperEditor.append(lblEditor, editor);
    wrapperReason.append(lblReason, reason);
    form.append(wrapperEditor, wrapperReason, submit);
    modal.append(svg, form);
    body.append(modal);

    modal.addEventListener("click", e => {
        const targ = e.target;
        // モーダル内の要素（form）をクリックした場合、何もしない
        // svg内の空白は場合targetとして認識されないので、個別にclickイベントを追加して対応する（ここで除外しない）
        if (targ !== modal) {
            return;
        }
        // form外のモーダルをクリックした場合、非表示にする
        modal.style.display = "none";
    });

    // 閉じるアイコンをクリックしたら非表示にする
    svg.addEventListener("click", () => modal.style.display = "none");

    form.addEventListener("submit", e => {
        e.preventDefault();

        if (!form.reportValidity()) {
            alert("名前と削除理由は両方入力必須です");
            return
        }
        const ok = confirm("十分に議論されてから削除をお願いします。OKを押すと削除されます。よろしいですか？");
        if (!ok) return;

        const idStr = getParamId();

        if (idStr === null) return;
        const id = parseInt(idStr);

        const body = {
            id,
            editor: editor.value,
            reason: reason.value,
        };
        fetch(form.action, {
            method: "post",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(body),
        })
            .then(res => {
                if (!res.ok) {
                    throw new Error("");
                }
                return res.text();
            })
            .then(txt => {
                txt += "\n" + "ホーム画面に戻ります。"
                alert(txt);
                // 削除された画面が表示され続けるため、ボタン押下時に
                // 404となる場合有り。ホーム画面に戻ってもらう
                window.location.href = "/";
            })
            .catch(err => alert("異常終了しました。後でまた試してください。"))
            .finally(() => modal.style.display = "none");

    });

    // enterでsubmitされるのを防ぐ。textareaは改行されるので大丈夫
    editor.addEventListener("keydown", e => {
        if (e.code === "Enter") {
            e.preventDefault();
        }
    })

}