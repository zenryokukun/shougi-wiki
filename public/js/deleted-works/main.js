
function addClickEvent() {
    const items = document.querySelectorAll(".deleted-item");
    for (const item of items) {
        // iアイコン div
        const trigger = item.querySelector(".modal-trigger");
        // 非表示になっているフル情報
        const info = item.querySelector(".full-info");
        // iアイコンのsvg
        const svg = item.querySelector("svg");
        trigger.addEventListener("click", () => {
            info.classList.toggle("hide");
            svg.classList.toggle("light-icon");
        });

        // 文字数が多いとモーダルがiアイコンに覆いかぶさってしまう
        // モーダルをクリックしたら閉じるようにする
        info.addEventListener("click", () => {
            info.classList.add("hide");
            svg.classList.remove("light-icon");
        });
    }

}

function main() {
    addClickEvent();
}
window.addEventListener("DOMContentLoaded", main);