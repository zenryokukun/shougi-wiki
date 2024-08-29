import Area from "./area.js";
import * as ct from "./constant.js";

export default class MainArea extends Area {
    constructor() {
        const x = ct.PADDING;
        const y = ct.PADDING;
        const w = ct.BLOCK * 9;
        const h = ct.BLOCK * 9;
        super(x, y, w, h);
    }

    calcRowCol(i) {
        const col = i % 9;
        const row = Math.floor(i / 9);
        return [col, row];
    }

    calcXY(col, row) {
        return [this.x + col * ct.BLOCK, this.y + row * ct.BLOCK];
    }

    /**@param {CanvasRenderingContext2D} ctx */
    render(ctx) {
        // 枠線
        ctx.strokeStyle = ct.BLACK;
        ctx.lineWidth = 2;
        ctx.strokeRect(...this.get());
        // マス
        ctx.lineWidth = 1;
        for (let i = 1; i < 9; i++) {
            let fromx = this.x + i * ct.BLOCK;
            let fromy = this.y;
            let tox = fromx;
            let toy = this.y + this.height;
            ctx.beginPath();
            ctx.moveTo(fromx, fromy);
            ctx.lineTo(tox, toy);
            ctx.stroke();

            fromx = this.x;
            fromy = this.y + i * ct.BLOCK;
            tox = this.x + this.width;
            toy = this.y + i * ct.BLOCK;
            ctx.beginPath();
            ctx.moveTo(fromx, fromy);
            ctx.lineTo(tox, toy);
            ctx.stroke();
        }
        // 〇
        ctx.fillStyle = ct.BLACK;
        ctx.beginPath();
        ctx.arc(ct.PADDING + 3 * ct.BLOCK, ct.PADDING + 3 * ct.BLOCK, ct.ARC, Math.PI * 2, 0);
        ctx.arc(ct.PADDING + 3 * ct.BLOCK, ct.PADDING + 6 * ct.BLOCK, ct.ARC, Math.PI * 2, 0);
        ctx.fill();
        ctx.beginPath();
        ctx.arc(ct.PADDING + 6 * ct.BLOCK, ct.PADDING + 3 * ct.BLOCK, ct.ARC, Math.PI * 2, 0); ctx.arc(ct.PADDING + 6 * ct.BLOCK, ct.PADDING + 6 * ct.BLOCK, ct.ARC, Math.PI * 2, 0);
        ctx.fill();
    }
}