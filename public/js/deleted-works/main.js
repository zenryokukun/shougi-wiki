
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
    }
    // const nodes = document.querySelectorAll(".deleted-item");
    // for (const node of nodes) {
    //     const hiddenNode = node.querySelector(".full-info");
    //     const tokens = hiddenNode.classList;
    //     node.addEventListener("touchstart", () => {
    //         if (tokens.contains("hide")) {
    //             // タップされたdeleted-item直下のfull-infoは表示
    //             tokens.remove("hide");
    //             // 他のfull-infoは全て非表示にする
    //             const hiddenNodes = document.querySelectorAll(".full-info");
    //             for (const hn of hiddenNodes) {
    //                 if (hiddenNode === hn) continue;
    //                 hn.classList.add("hide");
    //             }
    //         } else {
    //             // .hideクラスが含まれない（表示状態）の場合、非表示にする
    //             tokens.add("hide");
    //         }
    //     });
    // }
}

function main() {
    addClickEvent();
}
window.addEventListener("DOMContentLoaded", main);