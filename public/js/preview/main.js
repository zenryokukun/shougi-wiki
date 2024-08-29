import Canvas from "../shougi/canvas.js";
import GameHistory from "../shougi/game-history.js";
import initIcon from "../shougi/icon.js";
import initToggle from "../answer.js";

function init() {
    const cvs = document.querySelector("canvas");
    const sprite = document.querySelector("#sprite");
    const mainStr = cvs.dataset["main"];
    const tegomaStr = cvs.dataset["tegoma"];
    const goteTegomaStr = cvs.dataset["gotetegoma"];
    const kihuStr = cvs.dataset["kihu"];
    const main = JSON.parse(mainStr);
    const tegoma = JSON.parse(tegomaStr);
    const game = new GameHistory(main, tegoma);
    const mycvs = new Canvas(cvs, sprite);
    initIcon(game, mycvs);
    initToggle();
    mycvs.render(game);
    // 送信ボタン押下時の処理を生成し、eventをアタッチ。
    // canvasの初期状態（0手目）を画像にするため、mycvs.render後に実施する必要がある
    const fixBtn = document.querySelector(".fix");
    const callback = submitFactory(mainStr, tegomaStr, goteTegomaStr, kihuStr);
    fixBtn.addEventListener("click", callback);
}

function submitFactory(
    mainStr, tegomaStr, goteTegomaStr, kihuStr
) {
    const expNode = document.querySelector(".exp");
    const authorNode = document.querySelector(".author");
    const titleNode = document.querySelector(".title-wrapper h1");
    if (expNode === null || authorNode === null || titleNode === null) {
        alert("予期せぬエラーがありました。このページは閉じ、前のページからやり直してください");
        throw new Error("予期せぬエラー");
    }
    const explanation = expNode.innerHTML;
    const author = authorNode.textContent;
    const title = titleNode.textContent;
    const img = canvasToImage();
    const body = {
        explanation, author, title,
        main: mainStr,
        tegoma: tegomaStr,
        goteTegoma: goteTegomaStr,
        kihu: kihuStr,
        pic: img,
    };

    return function () {
        fetch("/api/insert-work", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(body),
        }).then(res => {
            if (!res.ok) {
                throw new Error("登録できませんでした。後で試してください");
            }
            return res.text();
        }).then(txt => {
            // 正常。メッセージを表示。
            alert(txt)
        })
            .catch(err => alert(err));
        // 2度目は押せなくする
        disable();
    };
}

// canvasから画像を生成する。base64だが、dataURLやMIMEが不要された形式
// data:image/png;base64,が冒頭に付く
function canvasToImage() {
    const cvs = document.querySelector("canvas");
    return cvs.toDataURL();
}

function disable() {
    const btn = document.querySelector(".fix");
    btn.disabled = true;
    console.log(btn);
}

window.addEventListener("load", init);