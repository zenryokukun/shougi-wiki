function hamClick() {
    const nav = document.querySelector(".nav-wrapper");
    const ham = document.querySelector(".hamburger");
    toggleClass(nav, "accordion");
    toggleClass(ham, "rotate");
}

function sbMenuClick(i) {
    const grid = document.querySelector(`[data-index='${i}']`);
    const carret = document.querySelector(`[data-caret='${i}']`);
    toggleClass(grid, "accordion");
    toggleClass(carret, "rotate");
}


/**
 * styleで指定した文字列がnodeのclassNameに存在しなければ追加、
 * 存在すれば削除する関数
 * @param {HTMLElement} node 
 * @param {string} style 
 * @returns void
 */
function toggleClass(node, style) {
    /**@type {DOMTokenList} */
    const tokens = node.classList
    if (tokens.contains(style)) {
        tokens.remove(style);
        return;
    }
    tokens.add(style);
}

window.addEventListener("load", e => {
    const ham = document.querySelector(".hamburger");
    ham.addEventListener("click", hamClick);

    /**
     * サイドバーのメニューのアコーディオンを展開するイベント（モバイルのみ）
     * サイドバーの作品一覧は固定のため、別のクラスを設定してある。ここでは取得されない点に注意
     */
    const sbMenu = document.querySelectorAll(".sb-heading-wrapper");

    const sbMenuArray = Array.from(sbMenu);
    for (let i = 0; i < sbMenuArray.length; i++) {
        sbMenuArray[i].addEventListener("click", e => sbMenuClick(i))
    }

    /**
     * サイドバーの作品一覧（固定）のアコーディオンを展開するイベント
     */
    const sbFixedMenu = document.querySelector(".sb-heading-wrapper-fixed");
    sbFixedMenu.addEventListener("click", () => {
        // data-index,data-caretに設定されているっ固定値:fixed-list
        sbMenuClick("fixed-list");
    });
});