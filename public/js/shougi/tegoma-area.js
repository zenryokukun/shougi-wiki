import Area from "./area.js";
import * as ct from "./constant.js";

export default class TegomaArea extends Area {
    constructor() {
        const x = ct.PADDING;
        const y = ct.PADDING * 2 + ct.BLOCK * 9;
        const w = ct.BLOCK * 9;
        const h = ct.TEGOMA;
        super(x, y, w, h);
    }

    calcXY(col) {
        return [this.x + ct.BLOCK * col, this.y];
    }

    render(ctx) {
        // ctx.fillStyle = "turquoise";
        // ctx.fillRect(...this.get());
    }
}