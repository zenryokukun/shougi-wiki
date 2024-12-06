function main() {
    /**
     * modalのオンオフ
     */
    // modalを表示
    const btn = document.querySelector("#show-modal");
    const modal = document.querySelector(".modal");
    const svg = document.querySelector(".modal svg");
    btn.addEventListener("click", () => {
        modal.classList.toggle("off");
    });
    // modalを非表示
    modal.addEventListener("click", e => {
        // modal内の子要素をクリックしたら何もしない
        if (e.target !== modal) return;
        modal.classList.add("off");
    })
    // 閉じるのSVGをクリックしたら非表示
    svg.addEventListener("click", () => {
        modal.classList.add("off");
    });

    const next = document.querySelector("#next");
    next.addEventListener("click", () => {
        alert("test")
    })

}

window.addEventListener("DOMContentLoaded", main);