import Area from "./area.js";
import * as ct from "./constant.js";

export default class WholeArea extends Area {
    constructor() {
        // 描写上のx,y,width,heightな点に注意。Tegomaの背景は描写しない
        const x = 0;
        const y = 0;
        const w = ct.PADDING * 2 + ct.BLOCK * 9;
        const h = ct.PADDING * 2 + ct.BLOCK * 9;
        super(x, y, w, h);
    }

    /**@param {CanvasRenderingContext2D} ctx */
    render(ctx) {
        ctx.fillStyle = ct.BG;
        ctx.fillRect(...this.get());
        ctx.strokeStyle = ct.BLACK;
        ctx.font = "12px sans-serif";
        const kanji = {
            0: "一", 1: "二", 2: "三", 3: "四", 4: "五", 5: "六", 6: "七", 7: "八", 8: "九",
        }
        const sx = this.x + ct.PADDING;
        for (let i = 0; i < 9; i++) {
            let ch = 9 - i;
            let m = ctx.measureText(ch);
            ctx.strokeText(ch, sx + i * ct.BLOCK + ct.BLOCK / 2 - m.width / 2, ct.PADDING - 5);
            ch = kanji[i];
            m = ctx.measureText(ch);
            ctx.strokeText(ch, sx + ct.BLOCK * 9 + 5, ct.PADDING + i * ct.BLOCK + ct.BLOCK / 2 + 5);
        }
    }
}