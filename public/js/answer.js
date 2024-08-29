export default function init() {
    const targ = document.querySelector(".info-wrapper")
    const btn = document.querySelector(".answer-handler-wrapper button");
    const mode = btn.dataset["inimode"]

    if (mode === "show") {
        targ.classList.toggle("accordion");
    }
    btn.addEventListener("click", toggle);
}

function toggle() {
    const targ = document.querySelector(".info-wrapper");
    const clist = targ.classList;
    clist.toggle("accordion");
    let txt = "";
    if (clist.contains("accordion")) {
        txt = "答えを隠す";
    } else {
        txt = "答えを表示する";
    }
    this.textContent = txt;
}