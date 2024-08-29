export default function init(game, cvs) {
    const nextBtn = document.querySelector(".next");
    const prevBtn = document.querySelector(".prev");
    nextBtn.addEventListener("click", e => nextClicked(game, cvs));
    prevBtn.addEventListener("click", e => prevClicked(game, cvs));
}

function nextClicked(game, cvs) {
    if (game.at >= game.main.length) {
        return;
    }
    game.next();
    update(game);
    renderKihu(game);
    cvs.render(game);
}

function prevClicked(game, cvs) {
    if (game.at === 0) {
        return;
    }
    game.prev();
    update(game);
    renderKihu(game);
    cvs.render(game);
}

function update(game) {
    const nextSVG = document.querySelector(".next svg");
    const prevSVG = document.querySelector(".prev svg");
    if (game.at === 0) {
        prevSVG.setAttribute("class", "disable");
    } else if (game.at == game.main.length - 1) {
        nextSVG.setAttribute("class", "disable");
    }
    if (game.at > 0) {
        prevSVG.setAttribute("class", "enable");
    }
    if (game.at < game.main.length - 1) {
        nextSVG.setAttribute("class", "enable");
    }
}

function renderKihu(game) {
    const { at } = game;
    const targ = document.querySelector(".current-kihu")
    if (at === 0) {
        targ.textContent = "";
        return;
    };
    const node = document.querySelector(`[data-kihu-i="${at - 1}"]`);
    targ.textContent = node.textContent;
}