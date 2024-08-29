import * as ct from "./constant.js";

export default class Sprite {
    constructor(img) {
        this.img = img;
        this.rows = 2;
        this.cols = 14;
        this.spWidth = 43;
        this.spHeight = 48;
        this.renderingWidth = ct.BLOCK;
        this.renderingHeight = ct.BLOCK;
    }
    index(koma) {
        const col = koma % this.cols;
        const row = Math.floor(koma / this.cols);
        return [col, row];
    }
    area(koma) {
        const [col, row] = this.index(koma);
        const { spWidth, spHeight } = this;
        return [col * spWidth, row * spHeight, spWidth, spHeight];
    }
}