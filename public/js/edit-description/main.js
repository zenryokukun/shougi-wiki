import * as form from "./form.js";
import { Canvas, GameHistory, click } from "./logic.js";
import { init as iconInit } from "./icon.js";

function init() {
    const cvs = document.querySelector("canvas");
    const sprite = document.querySelector("#sprite");
    const mainStr = cvs.dataset["main"];
    const tegomaStr = cvs.dataset["tegoma"];
    const main = JSON.parse(mainStr);
    const tegoma = JSON.parse(tegomaStr);
    const game = new GameHistory(main, tegoma);
    const mycvs = new Canvas(cvs, sprite);
    mycvs.render(game);

    cvs.addEventListener("click", e => {
        const x = e.offsetX;
        const y = e.offsetY;
        click(mycvs, game, x, y);
    });

    form.init(game);
    iconInit(game, mycvs);
}

window.addEventListener("load", init);