export default class Area {
    // canvas上のスタート位置
    x = 0;
    y = 0;
    // canvas上の幅と高さ
    width = 0;
    height = 0;

    constructor(x, y, width, height) {
        this.x = x;
        this.y = y;
        this.width = width;
        this.height = height;
    }

    get() {
        return [this.x, this.y, this.width, this.height];
    }

    // implement
    render(ctx) { }
}