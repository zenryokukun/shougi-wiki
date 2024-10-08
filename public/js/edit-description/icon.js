export function init(game, cvs) {
    // 「＜」「＞」のボタンを押したときの処理
    const nextButton = document.querySelector(".next");
    const prevButton = document.querySelector(".prev");
    nextButton.addEventListener("click", e => nextClicked(game, cvs));
    prevButton.addEventListener("click", e => prevClicked(game, cvs));
}

function nextClicked(game, cvs) {
    if (game.at >= game.data.length) {
        return;
    }
    game.next();
    updateIcon(game);
    updateTe(game);
    highlight(game.at);
    cvs.select.reset();
    cvs.render(game);
    // renderKihu(game);
}

function prevClicked(game, cvs) {
    if (game.at === 0) {
        return;
    }
    game.prev();
    updateIcon(game);
    updateTe(game);
    highlight(game.at);
    cvs.select.reset();
    cvs.render(game);
    // renderKihu(game);
}

export function updateIcon(game) {
    const nextSVG = document.querySelector(".next svg");
    const prevSVG = document.querySelector(".prev svg");
    if (game.at === 0) {
        prevSVG.setAttribute("class", "disable");
    } else if (game.at === game.data.length - 1) {
        // 手順の最終位置にいる場合。
        nextSVG.setAttribute("class", "disable");
    }
    if (game.at > 0) {
        prevSVG.setAttribute("class", "enable");
    }
    if (game.at < game.data.length - 1) {
        nextSVG.setAttribute("class", "enable");
    }
}

export function updateTe(game) {
    const te = document.querySelector(".te");
    let no = "";
    if (game.at === 0) {
        no = "初期配置";
    } else {
        no = game.at.toString() + "手目";
    }
    te.textContent = no;
}

/**
 * 棋譜を出力する関数
 * icon.jsにあるのも違和感あり、別ファイルにしようとしたが、
 * highlight関数がnextClicked等にも共通して使われているので、とりあえずこのまま
 */
export function renderKihu(game) {
    const kihus = game.extractKihu();
    const parent = document.querySelector("#kihu");
    while (parent.firstChild) {
        parent.removeChild(parent.firstChild);
    }

    for (let i = 0; i < kihus.length; i++) {
        const div = document.createElement("div");
        div.className = "kihu-row";
        const lbl = document.createElement("label");
        lbl.className = "kihu-number"
        const input = document.createElement("input");
        const step = "step" + (i + 1).toString();
        input.value = kihus[i];
        input.id = step;
        input.name = "kihu";
        lbl.htmlFor = step;
        lbl.textContent = (i + 1).toString() + "手目";
        div.append(lbl, input);
        parent.append(div);

    }
    parent.scrollTop = parent.scrollHeight;
    highlight(kihus.length);
}

function highlight(i) {
    // 既存のハイライトを消す
    const hl = document.querySelector(".highlight");
    if (hl !== null) {
        hl.classList.remove("highlight");
    }
    const targ = document.querySelector(`#step${i}`);
    if (targ) {
        targ.setAttribute("class", "highlight");
    }
    // scroll調整
    const parent = document.querySelector("#kihu");
    parent.scrollTop = 57 * (i - 1);
}