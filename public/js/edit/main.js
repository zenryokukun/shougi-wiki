
/**
 * canvasのwidthとheightはhtmlで固定値を記述すること。
 *  【編集モード】width="590" height="710"
 * cssでレスポンシブ化させる。canvas内のレンダリング部分も、自動でスケールしてくれる。
 * ただし、マウス位置から配列のインデックスを計算する際は、（描写はないので）
 * 自分で縮尺を計算する必要があるので注意。
 * @todo 描写、およびマウスの選択位置を計算する際に、同じエリアの計算をそれぞれ別のオブジェクトで実施してしまっている。描写をgameオブジェクトが一括して担ってしまっているため、エリア（将棋盤、手駒版、選択エリア）ごとにobjectを分け、それぞれにのエリア情報のx,y,width,heightを持たせ、描写の機能もそれぞれ持たせる。（と読みやすくなるかも？）
 */

/**@typedef {typeof import("./board.js").board} Board*/

import { board } from "./board.js";

// 躍動するこまた、駒たち。spriteと同じに並びにする必要ある。
const OU = 13, HU = 12, KYO = 11, KEI = 10, GIN = 9, KIN = 8, KAKU = 7, HI = 6, RHU = 5, RKYO = 4, RKEI = 3, RGIN = 2, RKAKU = 1, RHI = 0, E_OU = 27, E_HU = 26, E_KYO = 25, E_KEI = 24, E_GIN = 23, E_KIN = 22, E_KAKU = 21, E_HI = 20, E_RHU = 19, E_RKYO = 18, E_RKEI = 17, E_RGIN = 16, E_RKAKU = 15, E_RHI = 14;
// 空白マス
const EMPTY = -1;

const game = {

    /**@type {HTMLCanvasElement | null} */
    cvs: null,

    /**@type {CanvasRenderingContext2D | null} */
    ctx: null,

    /**@type {boolean} - trueなら編集モード */
    edit: false,

    /**
     * @param {HTMLCanvasElement} cvs
     * @param {boolean} isEdit 
     */
    init: function (cvs, isEdit) {
        this.cvs = cvs;
        this.ctx = cvs.getContext("2d");
        this.edit = isEdit;
    },

    /**@param {metrics} m */
    clear(m) {
        this.ctx.clearRect(0, 0, m.canvasWidth, m.canvasHeight);
    },

    /**@param {metrics} m */
    renderBG(m) {

        const { ctx } = this;

        // canvas全体を描写
        ctx.fillStyle = "#ffe599";
        // ctx.fillRect(0, 0, m.padding * 2 + m.boardWidth, m.canvasHeight);
        ctx.fillRect(...m.rectCanvas(this.edit));

        // 外側の枠線
        ctx.strokeStyle = "#000000";
        ctx.lineWidth = 2;
        ctx.strokeRect(...m.rectBoard());

        // 内側の枠線
        ctx.lineWidth = 1;
        ctx.beginPath();
        for (let i = 0; i < 9; i++) {
            ctx.moveTo(m.padding + i * m.block, m.padding);
            ctx.lineTo(m.padding + i * m.block, m.padding + m.boardHeight);
            ctx.moveTo(m.padding, m.padding + i * m.block);
            ctx.lineTo(m.padding + m.boardWidth, m.padding + i * m.block);
        }
        ctx.stroke();

        // ●
        ctx.beginPath();
        ctx.fillStyle = "#000000";
        ctx.arc(m.padding + 3 * m.block, m.padding + 3 * m.block, m.arc, 0, Math.PI * 2);
        ctx.arc(m.padding + 6 * m.block, m.padding + 3 * m.block, m.arc, 0, Math.PI * 2);
        ctx.fill();
        ctx.beginPath();
        ctx.arc(m.padding + 3 * m.block, m.padding + 6 * m.block, m.arc, 0, Math.PI * 2);
        ctx.arc(m.padding + 6 * m.block, m.padding + 6 * m.block, m.arc, 0, Math.PI * 2);
        ctx.fill();

        // 横の数字（1,2,~9）
        // 縦の漢数字（一,二,~九）
        ctx.font = `${m.font}px sans-serif`;
        // 縦
        const chineseMap = {
            0: "一", 1: "二", 2: "三", 3: "四", 4: "五", 5: "六", 6: "七", 7: "八", 8: "九",
        }

        for (let i = 0; i < 9; i++) {
            // 横
            const arabic = 9 - i;
            let measure = ctx.measureText(arabic);
            const arabX = m.padding + i * m.block + m.block / 2 - measure.width / 2;
            ctx.fillText(arabic, arabX, m.padding - m.font / 2);
            // 縦
            const chinese = chineseMap[i];
            measure = ctx.measureText(chinese);
            const chinaX = m.padding + m.boardWidth + measure.width / 2;
            const chinaY = m.padding + i * m.block + m.block / 2 + m.font / 2;
            ctx.fillText(chinese, chinaX, chinaY);
        }
    },

    /**
     * 将棋盤を描写する
     * @param {Board} board
     * @param {sprite} sprite 
     * @param {metrics} m
     */
    renderMain(board, sprite, m) {
        const data = board.main;
        for (let i = 0; i < data.length; i++) {
            const [col, row] = board.getPos(i);
            const koma = data[i];
            this.renderKoma(sprite, m, koma, col, row);
        }
    },

    /**
     * 将棋盤の１マスの駒を描写する
     * @param {sprite} sprite
     * @param {metrics} m 
     * @param {number} koma 
     * @param {number} col 
     * @param {number} row 
     */
    renderKoma(sprite, m, koma, col, row) {
        const [sx, sy] = sprite.calcPosition(koma);
        const [dx, dy] = m.calcXY(col, row);
        this.ctx.drawImage(sprite.img, sx * m.spriteWidth, sy * m.spriteHeight, m.spriteWidth, m.spriteHeight, dx, dy, m.block, m.block);
    },

    /**
     * 手駒版を描写する
     * @param {Board} board 
     * @param {sprite} sprite 
     * @param {metrics} m 
     * @param {number} koma 
     * 
     */
    renderTegoma(board, sprite, m) {
        const tegoma = board.getTegoma();
        const len = tegoma.length;
        for (let i = 0; i < len; i++) {
            const row = len - i;
            const [koma, cnt] = tegoma[i];
            const [x, y] = sprite.calcPosition(koma);
            const sx = x * m.spriteWidth;
            const sy = y * m.spriteHeight;
            const [dx, dy] = m.calcTegomaXY(row);
            let scale = 0.9;
            this.ctx.drawImage(
                sprite.img,
                sx, sy, m.spriteWidth, m.spriteHeight,
                dx, dy, m.block * scale, m.block * scale);
            if (cnt > 1) {
                this.ctx.font = "14px sans-serif";
                const txt = "x" + cnt.toString();
                this.ctx.strokeText(txt, dx + m.block / 2 + 7, dy + 5);
            }
        }

        if (this.edit) {
            this.ctx.strokeStyle = "#000000";
            const x = m.padding * 2 + m.boardWidth;
            const y = m.padding;
            this.ctx.strokeRect(x, y, m.block, m.boardHeight);
        }
    },

    /**
     * 選択エリアを描写する
     * @param {sprite} sprite 
     * @param {select} sel
     * @param {metrics} m 
     */
    renderSelect(sprite, sel, m) {

        const ctx = this.ctx;
        const { sente, gote } = sel;

        let startY = m.canvasHeight - 4 * m.block;
        let startX = m.padding;
        ctx.fillStyle = "#333";
        ctx.fillRect(startX, startY, m.boardWidth, m.block * 4);

        const len = sente.length;
        for (let i = 0; i < len; i++) {
            // 先手
            let koma = sente[i];
            let [sx, sy] = sprite.calcPosition(koma);
            let dx = koma > RHU ? m.padding + i * m.block : m.padding + (RHU - koma) * m.block;
            let dy = koma > RHU ? startY : startY + m.block;
            ctx.drawImage(
                sprite.img,
                sx * m.spriteWidth, sy * m.spriteHeight,
                m.spriteWidth, m.spriteHeight,
                dx, dy,
                m.block, m.block
            );
            // 後手
            let gotestartY = m.canvasHeight - 2 * m.block;
            koma = gote[i];
            [sx, sy] = sprite.calcPosition(koma);
            dx = koma > E_RHU ? m.padding + i * m.block : m.padding + (E_RHU - koma) * m.block;
            dy = koma > E_RHU ? gotestartY : gotestartY + m.block;
            ctx.drawImage(
                sprite.img,
                sx * m.spriteWidth, sy * m.spriteHeight,
                m.spriteWidth, m.spriteHeight,
                dx, dy,
                m.block, m.block
            );
        }
    },

    /**
     * 選択エリアの選択された駒をハイライト
     * @param {{koma:number,col:number,row:number}} selected 
     * @param {metrics} m 
     */
    highlightSelect: function (selected, m) {
        const ctx = this.ctx;
        ctx.globalAlpha = 0.3;
        ctx.fillStyle = "turquoise";
        const dx = m.padding + selected.col * m.block;
        const dy = m.canvasHeight - 4 * m.block + selected.row * m.block;
        ctx.fillRect(dx, dy, m.block, m.block);
        ctx.globalAlpha = 1;
    },


    /**
     * 将棋盤の選択された駒をハイライト
     * @param {{koma:number,col:number,row:number}} selected 
     * @param {metrics} m 
     */
    highlightBoardSelect: function (selected, m) {
        const ctx = this.ctx;
        ctx.globalAlpha = 0.3;
        ctx.fillStyle = "turquoise";
        const dx = m.padding + selected.col * m.block;
        const dy = m.padding + selected.row * m.block;
        ctx.fillRect(dx, dy, m.block, m.block);
        ctx.globalAlpha = 1;
    },

    /**
     * 手駒版の選択された駒をハイライト
     * @param {{koma:number,row:number}} selected 
     * @param {metrics} m 
     */
    highlightTegomaSelect: function (selected, m) {
        const ctx = this.ctx;
        ctx.globalAlpha = 0.3;
        ctx.fillStyle = "turquoise";
        const dx = m.padding * 2 + m.boardWidth;
        const dy = m.padding + selected.row * m.block;
        ctx.fillRect(dx, dy, m.block, m.block);
        ctx.globalAlpha = 1;
    },

    /**
     * mouseoverされたマス（将棋盤or手駒版）をハイライト
     * @param {number[]} rect
     * */
    highlightHover: function (rect) {
        const ctx = this.ctx;
        ctx.fillStyle = "#074f4f";
        ctx.globalAlpha = 0.3;
        ctx.fillRect(...rect);
        ctx.globalAlpha = 1;
    },
}

/**canvasに描写する各エリアのサイズを保持*/
const metrics = {

    /**@type {number | null} - canvasの幅 */
    canvasWidth: null,

    /**@type {number | null} - canvasの高さ */
    canvasHeight: null,

    /**@type {number | null} - 手駒盤の高さ。1マスの大きさと同じ想定*/
    tegoma: null,

    /**@type {number | null} - 将棋盤の幅*/
    boardWidth: null,

    /**@type {number | null} - 将棋盤の高さ*/
    boardHeight: null,

    /**@type {number} - マスの高さ＆幅 */
    block: 40,

    /**@type {number} - 将棋盤のまわりの余白 */
    padding: 20,

    /**@type {number} - 将棋盤中央にある4つの黒丸の大きさ */
    arc: 4,

    /**@type {number} - デフォルトの文字の大きさ */
    font: 11,

    /**@type {number} - sprite上の１コマの幅 */
    spriteWidth: 43,

    /**@type {number} - sprite上の１コマの幅*/
    spriteHeight: 48,

    /**@type {number} - canvasは大きさが自動スケールするのでそのX方向の比率 */
    scaleX: 1,
    /**@type {number} - canvasは大きさが自動スケールするのでそのY方向の比率 */
    scaleY: 1,

    /**
     * @param {CanvasRenderingContext2D} cvs 
     */
    init: function (cvs) {
        this.canvasWidth = cvs.width;
        this.canvasHeight = cvs.height;
        this.boardWidth = this.block * 9;
        this.boardHeight = this.boardWidth;
        this.tegoma = this.block;
        this.setScale(cvs);
    },

    /**
     * 指定したcol,rowに応じて、Canvas上の将棋盤の描写位置x,yを返す
     * @param {number} col - 将棋盤の横のindex 
     * @param {number} row - 将棋盤の縦のindex
     * @returns {[number,number]}
     */
    calcXY: function (col, row) {
        const x = col * this.block + this.padding;
        const y = row * this.block + this.padding;
        return [x, y];
    },

    /**
     * 指定したcolに応じて、Canvas上の手駒版の描写位置x,yを返す
     * 高さは固定のためrowはない。
     * @param {number} col 
     * @returns {[number,number]}
     */
    calcTegomaXY: function (row) {
        // const x = this.canvasWidth - this.padding - this.block * col;
        // const y = this.padding + this.boardHeight + this.padding / 1.5;
        const x = this.padding * 2 + this.boardWidth;
        const bottom = this.padding + this.boardHeight;
        const y = bottom - this.block * row;
        return [x, y];
    },

    /**
     * canvasの大きさを取得する。編集モード、通常モードだと
     * サイズが異なるので注意
     * @param {boolean} edit - 編集モードかいなか 
     * @returns 
     */
    rectCanvas: function (edit = false) {
        if (edit) {
            return [0, 0, this.canvasWidth, this.canvasHeight];
        } else {
            return [0, 0, this.padding * 2 + this.boardWidth, this.canvasHeight];

        }

    },

    /**
     * 将棋盤の大きさを取得する関数
     */
    rectBoard: function () {
        return [this.padding, this.padding, this.boardWidth, this.boardHeight];
    },

    /**
     * マウスが将棋盤内にあるか判定
     * @param {number} x - canvas上のoffsetX 
     * @param {number} y ^ canvas上のoffsetY
     * @returns {boolean}
     */
    isInBoard: function (x, y) {
        if (x <= this.padding * this.scaleX) return false;
        if (x >= this.padding * this.scaleX + this.boardWidth * this.scaleX) return false;
        if (y <= this.padding * this.scaleY) return false;
        if (y >= this.padding * this.scaleY + this.boardHeight * this.scaleY) return false;
        return true;
    },

    /**
     * マウスが手駒版内にあるか判定
     * @param {number} x - canvas上のoffsetX 
     * @param {number} y ^ canvas上のoffsetY
     * @returns {boolean}
     */
    isInTegoma: function (x, y) {
        const sx = (this.padding * 2 + this.boardWidth) * this.scaleX;
        const ex = sx + (this.block) * this.scaleX;
        const sy = this.padding * this.scaleY;
        const ey = sy + this.boardHeight * this.scaleY;
        return x > sx && x < ex && y > sy && y < ey;
    },

    /**
     * マウスが選択エリア内にあるか判定
     * @param {number} x - canvas上のoffsetX 
     * @param {number} y ^ canvas上のoffsetY
     * @returns {boolean}
     */
    isInSelect: function (x, y) {
        if (x <= this.padding * this.scaleX) return false;
        if (x >= (this.padding + this.boardWidth) * this.scaleX) return false;
        if (y < (this.canvasHeight - 4 * this.block) * this.scaleY) return false;
        if (y > (this.canvasHeight) * this.scaleY) return false;
        return true;
    },

    /**
     * 選択エリアの選択された駒をcol,rowのindexで返す
     * @param {number} x - canvas上のoffsetX 
     * @param {number} y ^ canvas上のoffsetY
     * @returns {[number,number]} - [col,row]
     */
    calcSelectPosition: function (x, y) {
        // エリア内の相対x,yを計算する
        const rx = x - (this.padding * this.scaleX);
        const col = Math.floor(rx / (this.block * this.scaleX));
        const ry = y - (this.canvasHeight - 4 * this.block) * this.scaleY;
        const row = Math.floor(ry / (this.block * this.scaleY));
        return [col, row];
    },

    /**
     * 将棋盤の選択された駒をcol,rowのindexで返す
     * @param {number} x - canvas上のoffsetX 
     * @param {number} y ^ canvas上のoffsetY
     * @returns {[number,number]} - [col,row]
     */
    calcBoardPosition: function (x, y) {
        const rx = x - (this.padding * this.scaleX);
        const ry = y - (this.padding * this.scaleY);
        const col = Math.floor(rx / (this.block * this.scaleX));
        const row = Math.floor(ry / (this.block * this.scaleY));
        return [col, row];
    },

    /**
     * 手駒版の選択された駒をrowのindexで返す
     * @param {number} x - canvas上のoffsetX 
     * @param {number} y ^ canvas上のoffsetY
     * @returns {[number,number]} - [col,row]
     */
    calcTegomaPosition: function (x, y) {
        const ry = y - this.padding * this.scaleY;
        const row = Math.floor(ry / (this.block * this.scaleY));
        return row;
    },

    /**
     * レスポンシブ？対応。cssで600px以下は
     * canvasの大きさが可変になる。
     * canvas内の描写は自動でスケールしてくれるが、
     * クリック位置からindexを計算する部分は、手動で
     * スケールを計算する必要がある。
     * @param {HTMLCanvasElement} cvs 
     * */
    setScale: function (cvs) {
        // clientWidth,clientHeightは実際にレンダリングされたサイズ
        this.scaleX = cvs.clientWidth / cvs.width;
        this.scaleY = cvs.clientHeight / cvs.height;
    },
};

const sprite = {

    /**@type {HTMLImageElement} */
    img: null,

    /**@type {number} - spriteは2行  */
    rows: 2,

    /**@type {number} - spriteは14列  */
    cols: 14,

    /**@param {HTMLImageElement} img */
    init: function (img) {
        this.img = img;
    },

    /**
     * 
     * @param {number} koma 
     * @returns {[number,number]} - [col,row]
     */
    calcPosition: function (koma) {
        const row = Math.floor(koma / this.cols);
        const col = koma % this.cols;
        return [col, row];
    },
};

const select = {
    cols: 8, //1行あたり８コマ
    sente: [HU, KYO, KEI, GIN, KIN, KAKU, HI, OU, RHU, RKYO, RKEI, RGIN, RKAKU, RHI],
    gote: [E_HU, E_KYO, E_KEI, E_GIN, E_KIN, E_KAKU, E_HI, E_OU, E_RHU, E_RKYO, E_RKEI, E_RGIN, E_RKAKU, E_RHI],
    _select: null,

    set: function (koma, col, row) {
        this._select = { koma, col, row };
    },

    reset: function () {
        this._select = null;
    },

    click: function (col, row) {
        const koma = this.get(col, row);
        this.set(koma, col, row);
    },

    get: function (col, row) {
        // 先手、後手、２列ずつ描写している前提。
        // col,rowに応じた駒を取得し、選択駒として設定する
        let convertedRow = row;
        const isGote = row >= 2;
        if (isGote) convertedRow -= 2;
        const i = this.cols * convertedRow + col;
        const koma = !isGote ? this.sente[i] : this.gote[i];
        return koma;
    },

    isSelected: function () {
        return this._select !== null ? true : false;
    },

    value: function () {
        return this._select !== null ? this._select.koma : null;
    },

};

const boardSelect = {
    _select: null,
    set: function (koma, col, row) {
        this._select = { koma, col, row };
    },
    reset: function () {
        this._select = null;
    },
    isSelected: function () {
        return this._select !== null ? true : false;
    }
};

const tegomaSelect = {
    _select: null,
    set: function (koma, row) {
        this._select = { koma, row };
    },
    reset: function () {
        this._select = null;
    },
    isSelected: function () {
        return this._select !== null ? true : false;
    },
    value: function () {
        return this._select !== null ? this._select.koma : null;
    },
}

const checker = {

    default: { ok: true, msg: null },

    do: function (cnts, koma) {

        if (koma === OU || koma === E_OU) {
            return this.ou(cnts);
        }

        if (koma === HI || koma === RHI || koma === E_HI || koma === E_RHI) {
            return this.hi(cnts);
        }

        if (koma === KAKU || koma === RKAKU || koma === E_KAKU || koma === E_RKAKU) {
            return this.kaku(cnts);
        }

        if (koma === KIN || koma === E_KIN) {
            return this.kin(cnts);
        }

        if (koma === GIN || koma === E_GIN || koma === RGIN || koma === E_RGIN) {
            return this.gin(cnts);
        }

        if (koma === KEI || koma === E_KEI || koma === RKEI || koma === E_RKEI) {
            return this.kei(cnts);
        }

        if (koma === KYO || koma === E_KYO || koma === RKYO || koma === E_RKYO) {
            return this.kyo(cnts);
        }

        if (koma === HU || koma === RHU || koma === E_HU || koma === E_RHU) {
            return this.hu(cnts);
        }

        return { ok: false, msg: "予期せぬエラー：存在しない駒？" };
    },

    ou(cnts) {
        if (cnts[OU] >= 2 || cnts[E_OU] >= 2) {
            return { ok: false, msg: "王と玉は１つずつしか置けません" };
        }
        return this.default;
    },

    hi(cnts) {
        let n = (cnts[HI] || 0) + (cnts[RHI] || 0) + (cnts[E_HI] || 0) + (cnts[E_RHI] || 0);
        if (n >= 3) {
            return { ok: false, msg: "飛と竜は2つまでしか置けません" };
        }
        return this.default;
    },

    kaku(cnts) {
        let n = (cnts[KAKU] || 0) + (cnts[RKAKU] || 0) + (cnts[E_KAKU] || 0) + (cnts[E_RKAKU] || 0);
        if (n >= 3) {
            return { ok: false, msg: "角と馬は２つまでしか置けません" };
        }
        return this.default;
    },

    kin(cnts) {
        let n = (cnts[KIN] || 0) + (cnts[E_KIN] || 0);
        if (n >= 5) {
            return { ok: false, msg: "金は４つまでしか置けません" };
        }
        return this.default;
    },

    gin(cnts) {
        let n = (cnts[GIN] || 0) + (cnts[E_GIN] || 0) + (cnts[RGIN] || 0) + (cnts[E_RGIN] || 0)
        if (n >= 5) {
            return { ok: false, msg: "銀・成銀は4つまでしか置けません" };
        }
        return this.default;
    },

    kei(cnts) {
        let n = (cnts[KEI] || 0) + (cnts[RKEI] || 0) + (cnts[E_KEI] || 0) + (cnts[E_RKEI] || 0);
        if (n >= 5) {
            return { ok: false, msg: "桂・成桂は4つまでしか置けません" };
        }
        return this.default;
    },

    kyo(cnts) {
        let n = (cnts[KYO] || 0) + (cnts[RKYO] || 0) + (cnts[E_KYO] || 0) + (cnts[E_RKYO] || 0);
        if (n >= 5) {
            return { ok: false, msg: "香・成香は４つまでしか置けません" };
        }
        return this.default;
    },

    hu(cnts) {
        let n = (cnts[HU] || 0) + (cnts[RHU] || 0) + (cnts[E_HU] || 0) + (cnts[E_RHU] || 0);
        if (n >= 19) {
            return { ok: false, msg: "歩・と金は18つまでしか置けません" };
        }
        return this.default;
    },
}

function checkSet(koma, board) {
    const main = board.main;
    // {HU:2}のように、駒：数のオブジェクトで管理。
    const cnts = { ...board.tegoma }; // 手駒版と合わせてチェックするため、手駒版をコピー
    // 選択した駒をあらかじめ加算しておく
    // cnts[koma] = 1;
    if (cnts[koma] === undefined) {
        cnts[koma] = 1;
    } else {
        cnts[koma] += 1;
    }
    // 将棋盤をcntsに展開
    for (let i = 0; i < main.length; i++) {
        let current = main[i];
        if (current === undefined || current === null || current === EMPTY) continue;
        if (cnts[current] === undefined) {
            cnts[current] = 1;
        } else {
            cnts[current] += 1;
        }
    }

    return checker.do(cnts, koma);
}

function render() {
    game.clear(metrics);
    game.renderBG(metrics);
    game.renderMain(board, sprite, metrics);
    game.renderTegoma(board, sprite, metrics);
    if (game.edit) {
        game.renderSelect(sprite, select, metrics);
    }
    if (select.isSelected()) {
        game.highlightSelect(select._select, metrics);
    }
    if (boardSelect.isSelected()) {
        game.highlightBoardSelect(boardSelect._select, metrics);
    }
    if (tegomaSelect.isSelected()) {
        game.highlightTegomaSelect(tegomaSelect._select, metrics);
    }
    if (hoverHandler.rect.length === 4) {
        game.highlightHover(hoverHandler.rect);
    }
}

/**
 * 
 * @param {MouseEvent} e
 * @param {HTMLCanvasElement} cvs 
 * @param {board} board 
 * @param {select} select 
 * @param {metrics} m 
 */
function click(e, cvs, board, trash, m) {
    // canvasは自動でスケールするため、処理前に現在のディスプレイサイズに応じて
    // 比率を再計算する。
    m.setScale(cvs)
    const x = e.offsetX;
    const y = e.offsetY;

    // selectエリアを選択した場合
    if (m.isInSelect(x, y)) {
        const [col, row] = m.calcSelectPosition(x, y);
        // 空白のマスは無効扱いにする。。２列に折り返して描写している都合。
        // 既に選択済の駒がある場合リセットする。
        if (col === 8) {
            select.reset();
            render();
            return
        };
        if (row === 1 || row === 3) {
            if (col === 6 || col === 7) {
                select.reset();
                render();
                return;
            };
        }
        // 同じ駒をクリックした場合、リセットする
        if (select.get(col, row) === select.value()) {
            select.reset();
            render();
            return;
        }
        tegomaSelect.reset(); // 手駒の選択は解除しておく
        trash.off();
        // boardエリアをクリック済みの場合はリセットする。
        boardSelect.reset();
        select.click(col, row);
        render();
        return;
    }

    // boardエリアを選択した場合
    if (m.isInBoard(x, y)) {
        const [col, row] = m.calcBoardPosition(x, y);
        if (select.isSelected()) {
            const koma = select.value();
            const stat = checkSet(koma, board);
            if (stat.ok) {
                board.set(koma, col, row);
                select.reset();
                hoverHandler.reset();
                // render();
            } else {
                alert(stat.msg);
            }
        } else {
            tegomaSelect.reset(); // 手駒の選択は解除しておく
            if (boardSelect.isSelected()) {
                const { koma: selectedKoma, col: selectedCol, row: selectedRow } = boardSelect._select;
                board.set(EMPTY, selectedCol, selectedRow);
                board.set(selectedKoma, col, row);
                boardSelect.reset();
                trash.off();
            } else {
                let koma = board.get(col, row);
                if (koma !== EMPTY) {
                    boardSelect.set(koma, col, row);
                    trash.on();
                } else {
                    trash.off();
                }
            }
        }
        render();
        return;
    }

    // 手駒エリア
    if (m.isInTegoma(x, y)) {
        // 選択エリアで駒が選択されていない場合
        if (!select.isSelected()) {
            const row = m.calcTegomaPosition(x, y);
            // 手駒版の上が0で計算されるため、下が0に換算して計算。
            const i = 8 - row;
            // 空白マスの場合何もしない。
            const tegoma = board.getTegoma();
            if (i >= tegoma.length) {
                return;
            }
            // tegomma配列の後ろから、手駒版の下から描写している。
            // 描写順に対応するようにindexを計算する。
            const j = tegoma.length - 1 - i;

            const [koma, _] = tegoma[j];

            // 同じ駒を選択した場合リセット
            if (tegomaSelect.value() === koma) {
                tegomaSelect.reset();
                trash.off();
                render();
                return;
            }

            // 手駒エリアの駒が選択された状態にする。
            tegomaSelect.set(koma, row);
            // hoverHandler.reset();
            select.reset();
            boardSelect.reset();
            trash.on();
            render();
            return;
        }
        // 選択エリアで駒が選択されている場合
        const koma = select.value();
        if (koma > HU || koma < HI) {
            alert("王・成駒・後手駒は手駒にできません");
            return;
        }
        const stat = checkSet(koma, board);
        if (!stat.ok) {
            alert(stat.msg);
            return;
        }
        board.setTegoma(koma);
        select.reset();
        hoverHandler.reset();
        render();
        return;
    }
}

const hoverHandler = {
    // hoverで色をつけるマス
    rect: [],
    reset: function () {
        this.rect = [];
    },
    set: function (x, y, m) {
        if (!select.isSelected()) {
            return;
        }
        if (m.isInBoard(x, y)) {
            const [col, row] = m.calcBoardPosition(x, y);
            this.rect = [col * m.block + m.padding, row * m.block + m.padding, m.block, m.block];
        } else if (m.isInTegoma(x, y)) {
            this.rect = [m.padding * 2 + m.boardWidth, m.padding, m.block, m.boardHeight];
        } else {
            this.rect = [];
        }
        render();
        return;
    },
};

function hover(e, cvs, board, trash, m) {
    const x = e.offsetX;
    const y = e.offsetY;
    hoverHandler.set(x, y, m);
}

function hasGyoku(main) {
    for (const koma of main) {
        if (koma === E_OU) {
            return true;
        }
    }
    return false;
}

const trash = {
    icon: null,
    disable: true,
    init: function (node) {
        this.icon = node;
        node.addEventListener("click", this.click.bind(this));
    },

    on: function () {
        this.disable = false;
        this.icon.setAttribute("class", "enable");
    },
    off: function () {
        this.disable = true;
        this.icon.setAttribute("class", "disable");
    },
    click: function () {
        // 活性化している場合、選択されたマスを削除
        if (!this.disable) {
            if (boardSelect.isSelected()) {
                const { _, col, row } = boardSelect._select;
                boardSelect.reset();
                board.set(EMPTY, col, row);
                this.off();
            } else if (tegomaSelect.isSelected()) {
                const { koma, _ } = tegomaSelect._select;
                board.reduceTegoma(koma);
                tegomaSelect.reset();
                this.off();
            }
            render();
        }
    },
};

function deletePressed(e) {
    if (e.key !== "Delete") {
        return;
    }
    trash.click();
}

function main() {
    const cvs = document.querySelector("canvas");
    const img = document.querySelector("#sprite");
    const isEdit = true;
    game.init(cvs, isEdit);
    board.init();
    metrics.init(cvs);
    sprite.init(img);
    if (isEdit) {
        const icon = document.querySelector(".disable");
        const redo = document.querySelector(".redo");
        const done = document.querySelector(".done");
        trash.init(icon);
        cvs.addEventListener("click", e => click(e, cvs, board, trash, metrics));
        cvs.addEventListener("mousemove", e => hover(e, cvs, board, select, metrics))
        window.addEventListener("keyup", e => deletePressed(e));
        redo.addEventListener("click", e => {
            const isClear = window.confirm("全てクリアします。よろしいですか？")
            if (!isClear) return;
            game.init(cvs, isEdit);
            board.init();
            metrics.init(cvs);
            tegomaSelect.reset();
            select.reset();
            boardSelect.reset();
            hoverHandler.reset();
            trash.off();
            render();
        });
        done.addEventListener("click", e => {
            const ok = confirm("盤面を確定し、手順解説作成に進みます。よろしいですか？");
            if (!ok) {
                return;
            }
            if (board.main.length === 0) return;
            if (!hasGyoku(board.main)) {
                alert("玉が配置されていません");
                return;
            }

            const form = document.querySelector(".hidden-form");
            const mainInput = document.querySelector("input[name='main']");
            const tegomaInput = document.querySelector("input[name='tegoma']");
            const mainStr = JSON.stringify(board.main);
            const tegomaStr = JSON.stringify(board.tegoma);
            mainInput.value = mainStr;
            tegomaInput.value = tegomaStr;
            form.submit();
            // location.href = `/edit/description?main=${main}&tegoma=${tegoma}`
        })
    }
    // board.set(13, 1, 1);
    // board.set(27, 5, 1);
    // board.setTegoma(12);
    // board.setTegoma(11);
    // board.setTegoma(11);

    render();
}

window.addEventListener("load", main);