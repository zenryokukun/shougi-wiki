import WholeArea from "./whole-area.js";
import MainArea from "./main-area.js";
import TegomaArea from "./tegoma-area.js";
import Sprite from "./sprite.js";
import * as ct from "./constant.js";

export default class Canvas {
    cvs = null;
    ctx = null;
    sprite = null;
    constructor(cvs, sprite) {
        this.cvs = cvs;
        this.ctx = cvs.getContext("2d");
        this.wholeArea = new WholeArea();
        this.mainArea = new MainArea();
        this.tegomaArea = new TegomaArea();
        this.sprite = new Sprite(sprite);
    }

    scale() {
        const cvs = this.cvs;
        return {
            scaleX: cvs.clientWidth / cvs.width,
            scaleY: cvs.clientHeight / cvs.height,
        }
    }

    clear() {
        this.ctx.clearRect(0, 0, this.cvs.width, this.cvs.height);
    }

    render(game) {
        const current = game.get();
        const board = current.main;
        const tegoma = current.tegoma;
        this.clear();
        this.renderBG();
        this.renderBoard(board);
        this.renderTegoma(tegoma);
    }

    renderTegoma(tegoma) {
        // 盤上の駒より小さく描写したいので。
        const scale = 0.9;
        const tegomaArea = this.tegomaArea
        const sprite = this.sprite;
        const ctx = this.ctx;
        const entries = Object.entries(tegoma);
        const len = entries.length;

        for (let i = 0; i < len; i++) {
            const [komaStr, cnt] = entries[i];
            const koma = parseInt(komaStr);
            // row換算。先頭が0。
            let col = len - i;
            // tegomaエリアの下に描写したいので、0→8に換算
            col = 9 - col;
            // canvas上のx,yを取得
            const [x, y] = tegomaArea.calcXY(col);
            // spriteのエリア
            const area = sprite.area(koma);
            ctx.drawImage(
                sprite.img,
                ...area,
                x, y, ct.BLOCK * scale, ct.BLOCK * scale
            )
            if (cnt > 1) {
                const txt = "x" + cnt.toString();
                ctx.strokeText(txt, x + ct.BLOCK / 2 + 7, y + 5);
            }
        }
    }

    renderBoard(board) {
        const mainArea = this.mainArea
        const sprite = this.sprite;
        const ctx = this.ctx;
        for (let i = 0; i < 81; i++) {
            if (board[i] === ct.EMPTY) continue;
            const [col, row] = mainArea.calcRowCol(i);
            const [x, y] = mainArea.calcXY(col, row);
            const koma = board[i];
            const area = sprite.area(koma);
            ctx.drawImage(sprite.img, ...area, x, y, ct.BLOCK, ct.BLOCK);
        }
    }

    renderBG() {
        const ctx = this.ctx;
        this.wholeArea.render(ctx);
        this.mainArea.render(ctx);
        this.tegomaArea.render(ctx);

    }
}