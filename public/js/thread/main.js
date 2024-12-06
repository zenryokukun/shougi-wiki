function main() {
    const icons = document.querySelectorAll(".reply-icon");
    for (const icon of icons) {
        icon.addEventListener("click", (e) => {
            e.preventDefault()
            const parentNode = icon.closest(".comment-info");
            if (parentNode === null) return;
            const seqNode = parentNode.querySelector(".info-id");
            if (seqNode === null) return;
            // リプライを選択したコメントのseq
            const seq = seqNode.textContent;
            // 返信seqを入力欄に自動セットする
            const targ = document.querySelector("#comment");
            targ.textContent = ">>" + seq + "\n";
            // textareaにフォーカスを当て、カーソルを入力欄の最後に移し、textareaのところまで自動スクロールさせる。
            targ.focus();
            const len = targ.value.length;
            targ.setSelectionRange(len, len);
            targ.scrollIntoView({ behavior: "smooth" });
        });
    }

}

window.addEventListener("DOMContentLoaded", main);