import * as form from "./form.js";

const BG = "#ffe599";
const BLACK = "#000000";
const BLOCK = 40;
const TEGOMA = BLOCK;
const PADDING = 20;
const ARC = 4;

// 躍動するこまた、駒たち。spriteと同じに並びにする必要ある。
const OU = 13, HU = 12, KYO = 11, KEI = 10, GIN = 9, KIN = 8, KAKU = 7, HI = 6, RHU = 5, RKYO = 4, RKEI = 3, RGIN = 2, RKAKU = 1, RHI = 0, E_OU = 27, E_HU = 26, E_KYO = 25, E_KEI = 24, E_GIN = 23, E_KIN = 22, E_KAKU = 21, E_HI = 20, E_RHU = 19, E_RKYO = 18, E_RKEI = 17, E_RGIN = 16, E_RKAKU = 15, E_RHI = 14;
// 空白マス
const EMPTY = -1;
// 先手、後手
const SENTE = 0;
const GOTE = 1;
// 駒名
const KOMA_NAME = {
    [OU]: "王", [E_OU]: "玉",
    [HU]: "歩", [E_HU]: "歩",
    [KYO]: "香", [E_KYO]: "香",
    [KEI]: "桂", [E_KEI]: "桂",
    [GIN]: "銀", [E_GIN]: "銀",
    [KIN]: "金", [E_KIN]: "金",
    [HI]: "飛", [E_HI]: "飛",
    [KAKU]: "角", [E_KAKU]: "角",
    [RHU]: "と", [E_RHU]: "と",
    [RKYO]: "成香", [E_RKYO]: "成香",
    [RKEI]: "成桂", [E_RKEI]: "成桂",
    [RGIN]: "成銀", [E_RGIN]: "成銀",
    [RHI]: "竜", [E_RHI]: "竜",
    [RKAKU]: "馬", [E_RKAKU]: "馬",
};

// 別のjsファイルからGameHistoryインスタンスを使うため。
// GameHistoryを初期化する際は、historyにもセットしておくこと。
let history;

class Area {
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

    isIn(x, y, m) {
        const sx = this.x * m.scaleX;
        const ex = sx + this.width * m.scaleX;
        const sy = this.y * m.scaleY;
        const ey = sy + this.height * m.scaleY;
        return x > sx && x < ex && y > sy && y < ey;
    }

    get() {
        return [this.x, this.y, this.width, this.height];
    }

    // implement
    click(x, y) { }
    render(ctx) { }
    index(x, y, m) { }
}

class WholeArea extends Area {
    constructor() {
        // 描写上のx,y,width,heightな点に注意。Tegomaの背景は描写しない
        const x = TEGOMA;
        const y = 0;
        const w = PADDING * 2 + BLOCK * 9;
        const h = PADDING * 2 + BLOCK * 9;
        super(x, y, w, h);
    }

    /**@param {CanvasRenderingContext2D} ctx */
    render(ctx) {
        ctx.fillStyle = BG;
        ctx.fillRect(...this.get());
        ctx.strokeStyle = BLACK;
        ctx.font = "12px sans-serif";
        const kanji = {
            0: "一", 1: "二", 2: "三", 3: "四", 4: "五", 5: "六", 6: "七", 7: "八", 8: "九",
        }
        const sx = this.x + PADDING;
        for (let i = 0; i < 9; i++) {
            let ch = 9 - i;
            let m = ctx.measureText(ch);
            ctx.strokeText(ch, sx + i * BLOCK + BLOCK / 2 - m.width / 2, PADDING - 5);
            ch = kanji[i];
            m = ctx.measureText(ch);
            ctx.strokeText(ch, sx + BLOCK * 9 + 5, PADDING + i * BLOCK + BLOCK / 2 + 5);
        }
    }
}

class MainArea extends Area {
    constructor() {
        const x = PADDING + TEGOMA;
        const y = PADDING;
        const w = BLOCK * 9;
        const h = BLOCK * 9;
        super(x, y, w, h);
    }

    index(x, y, m) {
        const blockx = BLOCK * m.scaleX;
        const blocky = BLOCK * m.scaleY;
        const sx = x - this.x * m.scaleX;
        const sy = y - this.y * m.scaleY;
        const col = Math.floor(sx / blockx);
        const row = Math.floor(sy / blocky);
        return [col, row];
    }

    calcRowCol(i) {
        const col = i % 9;
        const row = Math.floor(i / 9);
        return [col, row];
    }

    calcXY(col, row) {
        return [this.x + col * BLOCK, this.y + row * BLOCK];
    }

    /**@param {CanvasRenderingContext2D} ctx */
    render(ctx) {
        // 枠線
        ctx.strokeStyle = BLACK;
        ctx.lineWidth = 2;
        ctx.strokeRect(...this.get());
        // マス
        ctx.lineWidth = 1;
        for (let i = 1; i < 9; i++) {
            let fromx = this.x + i * BLOCK;
            let fromy = this.y;
            let tox = fromx;
            let toy = this.y + this.height;
            ctx.beginPath();
            ctx.moveTo(fromx, fromy);
            ctx.lineTo(tox, toy);
            ctx.stroke();

            fromx = this.x;
            fromy = this.y + i * BLOCK;
            tox = this.x + this.width;
            toy = this.y + i * BLOCK;
            ctx.beginPath();
            ctx.moveTo(fromx, fromy);
            ctx.lineTo(tox, toy);
            ctx.stroke();
        }
        // 〇
        ctx.fillStyle = BLACK;
        ctx.beginPath();
        ctx.arc(PADDING + 3 * BLOCK, PADDING + 3 * BLOCK, ARC, Math.PI * 2, 0);
        ctx.arc(PADDING + 3 * BLOCK, PADDING + 6 * BLOCK, ARC, Math.PI * 2, 0);
        ctx.fill();
        ctx.beginPath();
        ctx.arc(PADDING + 6 * BLOCK, PADDING + 3 * BLOCK, ARC, Math.PI * 2, 0); ctx.arc(PADDING + 6 * BLOCK, PADDING + 6 * BLOCK, ARC, Math.PI * 2, 0);
        ctx.fill();
    }
}

class TegomaArea extends Area {
    constructor() {
        const x = TEGOMA + PADDING * 2 + BLOCK * 9;
        const y = PADDING;
        const w = TEGOMA;
        const h = BLOCK * 9;
        super(x, y, w, h);
    }

    index(x, y, m) {
        const blocky = BLOCK * m.scaleY;
        const sy = y - this.y * m.scaleY;
        const row = Math.floor(sy / blocky);
        return row;
    }

    calcXY(row) {
        return [this.x, this.y + BLOCK * row];
    }

    render(ctx) {
        // ctx.fillStyle = "turquoise";
        // ctx.fillRect(...this.get());
    }
}

class GoteTegomaArea extends Area {
    constructor() {
        const x = 0;
        const y = PADDING;
        const w = TEGOMA;
        const h = BLOCK * 9;
        super(x, y, w, h);
    }
    index(x, y, m) {
        const blocky = BLOCK * m.scaleY;
        const sy = y - this.y * m.scaleY;
        const row = Math.floor(sy / blocky);
        return row;
    }
    calcXY(row) {
        return [this.x, this.y + BLOCK * row];
    }
}

class Promote extends Area {
    constructor(sx, sy, koma, pkoma, col, row) {
        let offsetY = BLOCK / 2
        if (isGoteKoma(koma)) {
            offsetY *= -1;
        }
        const x = sx - BLOCK / 2;
        const y = sy - offsetY;
        const w = BLOCK * 2;
        const h = BLOCK;
        super(x, y, w, h);
        this.koma = koma;
        this.pkoma = pkoma;
        this.col = col;
        this.row = row;
    }
    index(x, y, m) {
        const blockx = BLOCK * m.scaleX;
        const blocky = BLOCK * m.scaleY;
        const sx = x - this.x * m.scaleX;
        const sy = y - this.y * m.scaleY;
        const col = Math.floor(sx / blockx);
        const row = Math.floor(sy / blocky);
        return [col, row];
    }
    calcXY(col) {
        return [this.x + BLOCK * col, this.y];
    }
    render(ctx) {
        ctx.fillStyle = "#333";
        ctx.globalAlpha = 0.8;
        ctx.fillRect(...this.get());
        ctx.globalAlpha = 1;
    }
}

class Sprite {
    constructor(img) {
        this.img = img;
        this.rows = 2;
        this.cols = 14;
        this.spWidth = 43;
        this.spHeight = 48;
        this.renderingWidth = BLOCK;
        this.renderingHeight = BLOCK;
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

class Canvas {
    cvs = null;
    ctx = null;
    sprite = null;
    constructor(cvs, sprite) {
        this.cvs = cvs;
        this.ctx = cvs.getContext("2d");
        this.wholeArea = new WholeArea();
        this.mainArea = new MainArea();
        this.tegomaArea = new TegomaArea();
        this.goteTegomaArea = new GoteTegomaArea();
        this.sprite = new Sprite(sprite);
        this.select = new Select();
        this.promote = null;
    }

    setPromote(koma, col, row) {
        const [x, y] = this.mainArea.calcXY(col, row);
        const pkoma = this.select.getPromotedKoma(koma);
        this.promote = new Promote(x, y, koma, pkoma, col, row);
    }

    resetPromote() {
        this.promote = null;
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
        const goteTegoma = current.goteTegoma;
        this.clear();
        this.renderBG();
        this.renderBoard(board);
        this.renderTegoma(tegoma);
        this.renderGoteTegoma(goteTegoma);
        this.renderSelect();
        this.renderPromote();
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
            let row = len - i;
            // tegomaエリアの下に描写したいので、0→8に換算
            row = 9 - row;
            // canvas上のx,yを取得
            const [x, y] = tegomaArea.calcXY(row);
            // spriteのエリア
            const area = sprite.area(koma);
            ctx.drawImage(
                sprite.img,
                ...area,
                x, y, BLOCK * scale, BLOCK * scale
            )
            if (cnt > 1) {
                const txt = "x" + cnt.toString();
                ctx.strokeText(txt, x + BLOCK / 2 + 7, y + 5);
            }
        }
    }

    renderGoteTegoma(goteTegoma) {
        // 盤上の駒より小さく描写したいので。
        const scale = 0.9;
        const goteTegomaArea = this.goteTegomaArea;
        const sprite = this.sprite;
        const ctx = this.ctx;
        const entries = Object.entries(goteTegoma);
        const len = entries.length;
        for (let i = 0; i < len; i++) {
            const [komaStr, cnt] = entries[i];
            const koma = parseInt(komaStr);
            const [x, y] = goteTegomaArea.calcXY(i);
            const area = sprite.area(koma);
            ctx.drawImage(
                sprite.img,
                ...area,
                x, y, BLOCK * scale, BLOCK * scale
            )
            if (cnt > 1) {
                const txt = "x" + cnt.toString();
                ctx.strokeText(txt, x + BLOCK / 2 + 4, y + BLOCK - 7);
            }
        }
    }

    renderSelect() {
        if (!this.select.isSelected()) return;

        const ctx = this.ctx;
        const { col, row, which } = this.select;

        let x, y;
        if (which === "main") {
            [x, y] = this.mainArea.calcXY(col, row);
        } else if (which === "tegoma") {
            [x, y] = this.tegomaArea.calcXY(row);
        } else {
            [x, y] = this.goteTegomaArea.calcXY(row);
        }
        ctx.globalAlpha = 0.3;
        ctx.fillStyle = "turquoise";
        ctx.fillRect(x, y, BLOCK, BLOCK);
        ctx.globalAlpha = 1;

        this.renderPaths();
    }

    renderPaths() {
        const ctx = this.ctx;
        const paths = this.select.paths;
        const mainArea = this.mainArea;
        ctx.globalAlpha = 0.2;
        ctx.fillStyle = "turquoise";
        for (const path of paths) {
            const [col, row] = path;
            const [x, y] = mainArea.calcXY(col, row);
            ctx.fillRect(x, y, BLOCK, BLOCK);
        }
        ctx.globalAlpha = 1;
    }

    renderBoard(board) {
        const mainArea = this.mainArea
        const sprite = this.sprite;
        const ctx = this.ctx;
        for (let i = 0; i < 81; i++) {
            if (board[i] === EMPTY) continue;
            const [col, row] = mainArea.calcRowCol(i);
            const [x, y] = mainArea.calcXY(col, row);
            const koma = board[i];
            const area = sprite.area(koma);
            ctx.drawImage(sprite.img, ...area, x, y, BLOCK, BLOCK);
        }
    }

    renderPromote() {
        const promote = this.promote;
        const sprite = this.sprite;
        const ctx = this.ctx;

        if (promote === null) {
            return;
        }

        promote.render(ctx);
        const { koma, pkoma } = promote;
        const [komaX, komaY] = promote.calcXY(0);
        const komaArea = sprite.area(koma);
        const [pkomaX, pkomaY] = promote.calcXY(1);
        const pkomaArea = sprite.area(pkoma);
        ctx.drawImage(sprite.img, ...komaArea, komaX, komaY, BLOCK, BLOCK);
        ctx.drawImage(sprite.img, ...pkomaArea, pkomaX, pkomaY, BLOCK, BLOCK);
    }

    renderBG() {
        const ctx = this.ctx;
        this.wholeArea.render(ctx);
        this.mainArea.render(ctx);
        this.tegomaArea.render(ctx);

    }
}

class Select {
    koma = null;
    col = null;
    row = null;
    which = null;
    paths = null;
    isPromoted = false;
    // 成れるときに成らない場合
    notPromoting = false;

    constructor() { }

    main(koma, col, row, paths) {
        this.koma = koma;
        this.col = col;
        this.row = row;
        this.paths = paths;
        this.which = "main";
    }

    tegoma(koma, row, paths) {
        this.koma = koma;
        this.row = row;
        this.col = null;
        this.paths = paths;
        this.which = "tegoma";
    }

    goteTegoma(koma, row, paths) {
        this.koma = koma;
        this.row = row;
        this.col = null;
        this.paths = paths;
        this.which = "gotetegoma";
    }

    isSelected() {
        return this.koma !== null ? true : false;
    }

    isPlaceable(col, row) {
        for (const path of this.paths) {
            const [x, y] = path;
            if (col === x && row === y) {
                return true;
            }
        }
        return false;
    }

    reset() {
        this.koma = null;
        this.col = null;
        this.row = null;
        this.which = null;
        this.paths = null;
        this.isPromoted = false;
        this.notPromoting = false;
    }

    /**成る場合の処理 */
    promote() {
        this.koma = this.getPromotedKoma(this.koma);
        this.isPromoted = true;
    }

    /**成れるけど成らない場合のフラグ設定 */
    stay() {
        this.notPromoting = true;
    }

    getPromotedKoma(koma) {
        let subtr = 7;
        if (koma === HI || koma === E_HI || koma === KAKU || koma === E_KAKU) {
            subtr = 6;
        }
        return koma - subtr;
    }

    unpromoteKoma(koma) {
        let add = 7;
        if (koma === RHI || koma === E_RHI || koma === RKAKU || koma === E_RKAKU) {
            add = 6;
        }
        return koma + add;
    }
}

class GameHistory {
    /**@type {{main:number[],tegoma:{[key:string]:number},goteTegoma:{[key:string]:number},kihu:string}[]} */
    data = [];
    at = 0;

    constructor(main, tegoma) {
        const goteTegoma = this.calcGoteTegoma(main, tegoma);
        this.data.push({ main, tegoma, goteTegoma, kihu: "" });
    }

    next() {
        const ni = this.at + 1;
        if (ni >= this.data.length) {
            return this.at;
        }
        return this.at++;
    }

    prev() {
        const pi = this.at - 1;
        if (pi < 0) {
            return this.at;
        }
        return this.at--;
    }

    add(main, tegoma, goteTegoma, kihu) {
        // 現在位置が最後のデータの時。atはindexなので-1必要
        if (this.at === this.data.length - 1) {
            this.data.push({ main, tegoma, goteTegoma, kihu });
            this.at++;
            return;
        }
        // 盤面を戻して修正する場合：
        // 現在位置を含めてスライスし、新データをスライスにpush。
        const sliced = this.data.slice(0, this.at + 1);
        sliced.push({ main, tegoma, goteTegoma, kihu });
        // スライスでdataを更新
        this.data = sliced;
        // atも1進める
        this.at++;
    }

    /**直近ゲームのshallow-copyを返す */
    copy() {
        const current = this.get();
        const { main, tegoma, goteTegoma, kihu } = current;
        return {
            main: [...main],
            tegoma: { ...tegoma },
            goteTegoma: { ...goteTegoma },
            kihu,
        }
    }

    /**直近ゲーム版のreferenceを返す */
    get() {
        return this.data[this.at];
    }

    set(select, col, row, kihuStr) {
        const { main, tegoma, goteTegoma } = this.copy();
        const i = row * 9 + col;
        const bef = main[i];
        main[i] = select.koma;
        if (select.which === "main") {
            // 手駒に加える
            if (bef !== EMPTY) {
                let captive = bef;
                if (isPromotedKoma(captive)) {
                    captive = getUnpromotedKoma(captive);
                }
                captive = getOppositeKoma(captive);

                let refTegoma;
                if (isSenteKoma(captive)) {
                    refTegoma = tegoma;
                } else {
                    refTegoma = goteTegoma;
                }
                if (refTegoma[captive] === undefined) {
                    refTegoma[captive] = 1;
                } else {
                    refTegoma[captive] += 1;
                }
            }
            // 移動元のマスを空にする。
            main[select.row * 9 + select.col] = EMPTY;
        } else if (select.which === "tegoma") {
            tegoma[select.koma] -= 1;
            if (tegoma[select.koma] === 0) {
                delete tegoma[select.koma];
            }
        } else if (select.which === "gotetegoma") {
            goteTegoma[select.koma] -= 1;
            if (goteTegoma[select.koma] === 0) {
                delete goteTegoma[select.koma];
            }
        }
        select.reset();
        this.add(main, tegoma, goteTegoma, kihuStr);
    }

    getKoma(col, row) {
        const i = row * 9 + col;
        const current = this.get();
        return current.main[i];
    }

    getTegoma(row) {
        const current = this.get();
        const tegoma = current.tegoma;
        const entries = Object.entries(tegoma);
        let i = 8 - row;
        if (i >= entries.length) return;
        i = entries.length - 1 - i;
        const [koma, _] = entries[i];
        return parseInt(koma);
    }

    getGoteTegoma(row) {
        const current = this.get();
        const goteTegoma = current.goteTegoma;
        const entries = Object.entries(goteTegoma);
        if (row >= entries.length) return;
        const [koma, _] = entries[row];
        return parseInt(koma);
    }

    calcGoteTegoma(main, tegoma) {
        // 後手番は盤上にない駒を全て手駒にできる
        const total = { ...tegoma };
        for (let i = 0; i < main.length; i++) {
            const koma = main[i];
            if (koma === EMPTY) {
                continue;
            }
            if (total[koma] === undefined) {
                total[koma] = 1;
                continue;
            }
            total[koma] += 1;
        }
        // 玉は手駒に置けないので無視
        const tHu = (total[HU] || 0) + (total[E_HU] || 0) + (total[RHU] || 0) + (total[E_RHU] || 0);
        const tKyo = (total[KYO] || 0) + (total[E_KYO] || 0) + (total[RKYO] || 0) + (total[E_RKYO] || 0);
        const tKei = (total[KEI] || 0) + (total[E_KEI] || 0) + (total[RKEI] || 0) + (total[E_RKEI] || 0);
        const tGin = (total[GIN] || 0) + (total[E_GIN] || 0) + (total[RGIN] || 0) + (total[E_RGIN] || 0);
        const tKin = (total[KIN] || 0) + (total[E_KIN] || 0);
        const tKaku = (total[KAKU] || 0) + (total[E_KAKU] || 0) + (total[RKAKU] || 0) + (total[E_RKAKU] || 0);
        const tHi = (total[HI] || 0) + (total[E_HI] || 0) + (total[RHI] || 0) + (total[E_RHI] || 0);
        const goteTegoma = {
            [E_HU]: 18 - tHu,
            [E_KYO]: 4 - tKyo,
            [E_KEI]: 4 - tKei,
            [E_GIN]: 4 - tGin,
            [E_KIN]: 4 - tKin,
            [E_KAKU]: 2 - tKaku,
            [E_HI]: 2 - tHi,
        };
        // 枚数が０の駒を除外する
        for (const key of Object.keys(goteTegoma)) {
            if (goteTegoma[key] === 0) {
                delete goteTegoma[key]
            }
        }
        return goteTegoma;
    }

    isSameKihu(kihuStr) {
        // trueなら「同」
        // 一文字目は▲or△な点に注意
        let i = this.at;
        let current = this.data[i];
        let latest = current.kihu;
        while (latest.slice(1, 2) === "同") {
            i -= 1;
            current = this.data[i];
            latest = current.kihu;
        }
        const prevPos = latest.slice(1, 3);
        const nextPos = kihuStr.slice(1, 3);
        return prevPos === nextPos;
    }

    extractKihu() {
        const kihus = [];
        const data = this.data.slice(0, this.at + 1);
        for (const d of data) {
            if (!d.kihu || d.kihu.length === 0) {
                continue;
            }
            kihus.push(d.kihu);
        }
        return kihus;
    }
}


function click(mycvs, game, x, y) {
    const mainArea = mycvs.mainArea;
    const tegomaArea = mycvs.tegomaArea;
    const goteTegomaArea = mycvs.goteTegomaArea;
    const scale = mycvs.scale();
    const select = mycvs.select;
    const promote = mycvs.promote;

    // 成る成らずの選択がでている場合、そこしかクリックできない
    if (promote !== null) {
        if (!promote.isIn(x, y, scale)) {
            return;
        }
        const [col, _] = promote.index(x, y, scale);
        promoteClick(game, mycvs, col);
        mycvs.render(game);
        // ここでreturnしないとmainAreaのクリック処理流れるので必須
        return;
    }

    if (mainArea.isIn(x, y, scale)) {
        const [col, row] = mainArea.index(x, y, scale);
        mainClick(game, mycvs, col, row);
    }

    else if (tegomaArea.isIn(x, y, scale)) {
        const row = tegomaArea.index(null, y, scale);
        tegomaClick(game, select, row);
    }

    else if (goteTegomaArea.isIn(x, y, scale)) {
        const row = goteTegomaArea.index(null, y, scale);
        goteTegomaClick(game, select, row);
    }

    mycvs.render(game);
}

function mainClick(game, mycvs, col, row) {
    const koma = game.getKoma(col, row);
    const turn = getTurn(game);
    const select = mycvs.select;
    const current = game.get();
    // 選択した駒が空白ではない場合
    if (koma !== EMPTY) {
        // 既に駒が選択されており、動けるマスが敵駒の位置の場合
        if (select.isSelected() && select.isPlaceable(col, row)) {
            const promote = isPromote(select, row, turn);
            if (promote.force) {
                select.promote();
            } else if (promote.can) {
                mycvs.setPromote(select.koma, col, row);
                // 成・不成を選ばせるため、まだ盤面確定はできない。なのでreturn。
                return;
            }
            setGame(game, mycvs, col, row);
            return;
        }
        // 駒が選択されていない場合
        // 先手・後手に対応した駒を選択していなかったらreturn
        if (turn === SENTE && !isSenteKoma(koma)) {
            return;
        }
        if (turn === GOTE && !isGoteKoma(koma)) {
            return;
        }
        const cand = move(current.main, col, row);
        select.main(koma, col, row, cand);
        return;
    }
    // 空白マスをクリックした場合
    if (!select.isSelected()) {
        // 何も選択されていなければreturn
        return;
    }
    if (!select.isPlaceable(col, row)) {
        // 配置できないマスならreturn
        return;
    }

    // 配置
    if (select.which === 'main') {
        const promote = isPromote(select, row, turn);
        if (promote.force) {
            select.promote();
        } else if (promote.can) {
            mycvs.setPromote(select.koma, col, row);
            // 成・不成を選ばせるため、まだ盤面確定はできない。なのでreturn。
            return;
        }
    }
    setGame(game, mycvs, col, row);
}

/**先手の手駒版のクリック処理 */
function tegomaClick(game, select, row) {
    const turn = getTurn(game);
    if (turn !== SENTE) {
        return;
    }
    const koma = game.getTegoma(row);
    const current = game.get();
    if (!koma) return;
    const cand = searchUtsu(current.main, koma)
    select.tegoma(koma, row, cand);
}

/**後手の手駒版のクリック処理 */
function goteTegomaClick(game, select, row) {
    const turn = getTurn(game);
    if (turn !== GOTE) {
        return;
    }
    const koma = game.getGoteTegoma(row);
    const current = game.get();
    if (!koma) return;
    const cand = searchUtsu(current.main, koma);
    select.goteTegoma(koma, row, cand);
}

function promoteClick(game, mycvs, col) {
    const promote = mycvs.promote;
    if (col === 1) {
        mycvs.select.promote();
    } else {
        mycvs.select.stay();
    }
    setGame(game, mycvs, promote.col, promote.row);
}

function setGame(game, mycvs, col, row) {
    const select = mycvs.select;
    if (getTurn(game) === SENTE && !isOute(game, select, col, row)) {
        alert("王手ではありません")
        select.reset();
        mycvs.resetPromote();
        return;
    }
    // 棋譜を計算するには盤面を確定前にする必要がある。
    // moveでは自駒が配置されているマスは候補として抽出しないため。
    const kihuStr = kihu(game, select, col, row);
    game.set(select, col, row, kihuStr);
    mycvs.resetPromote();
    form.updateIcon(game); // icon更新
    form.updateTe(game);
    form.renderKihu(game);
}

/**寄付用：同じとこに動かせる、同じ駒を返す。 */
function getSameKoma(game, select, cx, cy) {
    let koma = select.koma;
    if (select.isPromoted) {
        koma = select.unpromoteKoma(koma);
    }
    const current = game.get();
    const main = current.main;
    const same = [];
    for (let i = 0; i < main.length; i++) {
        // 異なる駒の場合はスキップ
        if (koma !== main[i]) {
            continue;
        }
        const x = i % 9;
        const y = Math.floor(i / 9);
        // 元の駒（動かす元の位置）はスキップ
        if (x === select.col && y === select.row) {
            continue;
        }
        const cand = move(main, x, y);
        for (const [c, r] of cand) {
            if (c === cx && r === cy) {
                same.push([x, y]);
            }
        }
    }
    return same;
}

/**
 * 棋譜のルールは複雑。。。
 * https://www.shogi.or.jp/faq/kihuhyouki.html
 */
function kihu(game, select, cx, cy) {
    const same = getSameKoma(game, select, cx, cy);
    const bx = select.col;
    const by = select.row;
    const koma = select.koma;
    let unpromoteKomaName;
    let nari = "";
    if (select.isPromoted) {
        const unpromoteKoma = getUnpromotedKoma(koma);
        unpromoteKomaName = KOMA_NAME[unpromoteKoma];
        nari = "成";
    }
    if (select.notPromoting) {
        nari = "不成";
    }
    const komaName = unpromoteKomaName || KOMA_NAME[koma];
    const mark = isSenteKoma(koma) ? "▲" : "△";
    // 右左直
    let pos = "";
    // 上引寄
    let mv = "";
    // 棋譜
    let kihuStr = mark;
    let suuji = colToNumber(cx);
    let kanji = rowToKanji(cy);
    kihuStr = mark + suuji + kanji;
    if (game.isSameKihu(kihuStr)) {
        kihuStr = mark + "同";
    }

    if (select.which === "main") {
        // 左右直：同じ高さのマスをセット
        const sayu = [];
        for (const [x, y] of same) {
            if (y === by) {
                sayu.push([x, y]);
            }
        }

        let left = false;
        let right = false;
        let sente = isSenteKoma(koma);
        for (const [x, y] of sayu) {
            if (x < bx) {
                right = true;
            }
            if (x > bx) {
                left = true;
            }
        }
        if (left && right) {
            pos = "直";
        } else if (left && !right) {
            pos = sente ? "左" : "右";
        } else if (!left && right) {
            pos = sente ? "右" : "左";
        }

        // 真っすぐ動いた場合の直（例外）
        // 2枚横並びしている状態で、上に動いた場合で、駒が金（成含む）銀の場合、直
        if ((left && !right) || (!left && right)) {
            if (cx - bx === 0) {
                if (koma === KIN || koma === E_KIN
                    || koma === GIN || koma === E_GIN
                    || koma === RHU || koma === E_RHU
                    || koma === RKYO || koma === E_RKYO
                    || koma === RKEI || koma === E_RKEI
                    || koma === RGIN || koma === E_RGIN
                ) {
                    pos = "直"
                }
            }
        }

        // 上寄引
        const uehiku = [];
        for (const [x, y] of same) {
            if (y !== by) {
                uehiku.push([x, y]);
            }
        }

        if (uehiku.length > 0 && pos !== "直") {
            let hasSameRow = false;
            for (const [x, y] of uehiku) {
                if (x === bx) {
                    hasSameRow = true;
                }
            }
            if (cy - by === 0) {
                mv = "寄";
            } else if (cy - by > 0) {
                if (pos === "" || hasSameRow) {
                    mv = sente ? "引" : "上";
                }
            } else if (cy - by < 0) {
                if (pos === "" || hasSameRow) {
                    mv = sente ? "上" : "引";
                }
            }

            if (koma === RHI || koma === E_RHI
                || koma === RKAKU || koma === E_RKAKU
                || koma === KAKU || koma === E_KAKU
            ) {
                // 竜例外（連盟HPの竜のパターンE対応）
                // 左右は同じ高さの駒のみに算出しているが、竜・馬（角）の場合は
                // 異なる場合でも必要なケースあり。斜めに動き（上・引）、もう１つの竜と同じ列に来た場合、
                // どちらもの竜も「上（引）」となってしまうため、左右で表す。「寄」の場合は考慮不要。
                const [ox, oy] = uehiku[0]; //竜は２駒しかないので。
                // 動かした駒の向き
                const moveDirY = cy - by;
                // もう片方の駒の向き
                const otherDirY = cy - oy;
                // 同じ向きか
                const isSameDir = (moveDirY < 0 && otherDirY < 0) || (moveDirY > 0 && otherDirY > 0);

                if ((bx - ox > 0) && isSameDir) {
                    mv = sente ? "右" : "左";
                } else if ((bx - ox < 0) && isSameDir) {
                    mv = sente ? "左" : "右";
                }
            }

        }
        kihuStr += komaName + pos + mv + nari;
    } else {
        kihuStr += komaName
        if (same.length !== 0) {
            kihuStr += "打";
        }
    }
    return kihuStr;
}

function isSameSide(koma, targ) {
    if (koma >= RHI && koma <= OU && targ >= RHI && targ <= OU) {
        return true;
    }
    if (koma >= E_RHI && koma <= E_OU && targ >= E_RHI && targ <= E_OU) {
        return true;
    }
    return false;
}

function isSenteKoma(koma) {
    return koma >= RHI && koma <= OU;
}

function isGoteKoma(koma) {
    return koma >= E_RHI && koma <= E_OU;
}

function isPromotedKoma(koma) {
    return (koma >= RHI && koma <= RHU) || (koma >= E_RHI && koma <= E_RHU);
}

function getUnpromotedKoma(koma) {
    let add = 7;
    if (koma === RHI || koma === E_RHI || koma === RKAKU || koma === E_RKAKU) {
        add = 6;
    }
    return koma + add;
}

function getOppositeKoma(koma) {
    let opposite;
    if (isSenteKoma(koma)) {
        opposite = koma + 14;
    } else {
        opposite = koma - 14;
    }
    return opposite;
}

function getTurn(game) {
    return game.at % 2 === 0 ? SENTE : GOTE;
}

function isIn(col, row) {
    return col >= 0 && col < 9 && row >= 0 && row < 9;
}

function colToNumber(x) {
    return (9 - x).toString();
}

function rowToKanji(y) {
    const kanji = {
        0: "一", 1: "二", 2: "三", 3: "四", 4: "五", 5: "六", 6: "七", 7: "八", 8: "九",
    };
    return kanji[y];
}

function isPromote(select, row) {
    const koma = select.koma;
    const ret = {
        promote: false,
        force: false,
        can: false,
    }
    if (isSenteKoma(koma)) {
        if (koma === KIN || koma === OU || (koma >= RHI && koma <= RHU)) {
            return ret;
        }

        if ((koma === HU && row === 0) ||
            (koma === KYO && row === 0) ||
            (koma === KEI && row <= 1)
        ) {
            ret.promote = true;
            ret.force = true;
            return ret;
        }

        if (row <= 2 || select.row <= 2) {
            ret.promote = true;
            ret.can = true;
            return ret;
        }

    } else if (isGoteKoma(koma)) {
        if (koma === E_KIN || koma === E_OU || (koma >= E_RHI && koma <= E_RHU)) {
            return ret;
        }
        if ((koma === E_HU && row === 8) ||
            (koma === E_KYO && row === 8) ||
            (koma === E_KEI && row >= 7)
        ) {
            ret.promote = true;
            ret.force = true;
            return ret;
        }
        if (row >= 6 || select.row >= 6) {
            ret.promote = true;
            ret.can = true;
            return ret;
        }
    }
    return ret;
}

function isOute(game, select, cx, cy) {
    // 盤面をコピーし、選択した位置で仮確定
    const current = game.get();
    const board = [...current.main];
    board[cy * 9 + cx] = select.koma;
    if (select.which === "main") {
        board[select.row * 9 + select.col] = EMPTY;
    }
    // 玉の位置のindexを取得
    let gindex = -1;
    for (let i = 0; i < board.length; i++) {
        if (board[i] === E_OU) {
            gindex = i;
            break;
        }
    }
    // 王手がかかっているかチェック
    let oute = false;
    for (let i = 0; i < board.length; i++) {
        const koma = board[i];
        if (!isSenteKoma(koma)) {
            continue;
        }
        const c = i % 9;
        const r = Math.floor(i / 9);
        const cands = move(board, c, r);
        for (const cand of cands) {
            const [x, y] = cand;
            if (gindex === y * 9 + x) {
                oute = true;
                break;
            }
        }
        if (oute) {
            break;
        }
    }
    return oute;
}

const dir = {
    left: function (isSente) {
        const x = isSente ? -1 : 1;
        const y = 0;
        return [x, y];
    },
    leftup: function (isSente) {
        const x = isSente ? -1 : 1;
        const y = isSente ? -1 : 1;
        return [x, y];
    },
    up: function (isSente) {
        const x = 0;
        const y = isSente ? -1 : 1;
        return [x, y];
    },
    rightup: function (isSente) {
        const x = isSente ? 1 : -1;
        const y = isSente ? -1 : 1;
        return [x, y];
    },
    right: function (isSente) {
        const x = isSente ? 1 : -1;
        const y = 0;
        return [x, y];
    },
    rightdown: function (isSente) {
        const x = isSente ? 1 : -1;
        const y = isSente ? 1 : -1;
        return [x, y];

    },
    down: function (isSente) {
        const x = 0
        const y = isSente ? 1 : -1;
        return [x, y];
    },
    leftdown: function (isSente) {
        const x = isSente ? -1 : 1;
        const y = isSente ? 1 : -1;
        return [x, y];
    },
    // 桂例外:
    keileft: function (isSente) {
        const x = -1;
        const y = isSente ? -2 : 2;
        return [x, y];
    },
    keiright: function (isSente) {
        const x = 1;
        const y = isSente ? -2 : 2;
        return [x, y];
    },
};

/**
 * 飛車・角・香車以外の駒の選択可能マスを検索
 * @param {GameHistory.main} board 
 * @param {[[x:number,y:number]]} dirs 
 * @param {number} c 
 * @param {number} r 
 * @returns 
 */
function search(board, dirs, c, r) {
    const koma = board[r * 9 + c];
    const cand = [];
    for (const dir of dirs) {
        const [x, y] = dir;
        const nx = x + c;
        const ny = y + r;
        if (!isIn(nx, ny)) {
            continue;
        }
        const targ = board[ny * 9 + nx];
        if (targ === EMPTY) {
            cand.push([nx, ny]);
        } else if (!isSameSide(koma, targ)) {
            cand.push([nx, ny]);
        }
    }
    return cand;
}

/**飛角香のおけるマスを計算 */
function searchHikakuKyo(board, dirs, c, r) {
    const cand = [];
    for (const dir of dirs) {
        searchLongPath(board, dir, c, r, cand);
    }
    return cand;
}

/**竜馬のおけるマスを計算 */
function searchRyuUma(board, dirs, c, r) {
    const koma = board[r * 9 + c];
    const cand = [];
    for (const dir of dirs) {
        const [x, y] = dir;
        const nx = x + c;
        const ny = y + r;
        if (!isIn(nx, ny)) {
            continue;
        }

        if (koma === RHI || koma === E_RHI) {
            // 竜の飛び
            if ((x === 0 && y !== 0) || (x !== 0 && y === 0)) {
                searchLongPath(board, dir, c, r, cand);
                continue;
            }

        } else if (koma === RKAKU || koma === E_RKAKU) {
            // 角の飛び
            if (x !== 0 && y !== 0) {
                searchLongPath(board, dir, c, r, cand);
                continue;
            }
        }

        const targ = board[ny * 9 + nx];
        if (targ === EMPTY) {
            cand.push([nx, ny]);
        } else if (!isSameSide(koma, targ)) {
            cand.push([nx, ny]);
        }
    }
    return cand;
}

/**飛角のような伸びる駒の、１つの方角の置けるマスを計算 */
function searchLongPath(board, dir, c, r, cand) {
    const koma = board[r * 9 + c];
    const [x, y] = dir;
    let nx = x + c;
    let ny = y + r;
    while (true) {
        if (!isIn(nx, ny)) {
            break;
        }
        const targ = board[ny * 9 + nx];
        if (isSameSide(koma, targ)) {
            break;
        }
        cand.push([nx, ny]);
        // 敵ゴマは取れるので、push後にbreak
        if (targ !== EMPTY && !isSameSide(koma, targ)) {
            break;
        }
        nx += x;
        ny += y;
    }
    return cand;
}

function searchUtsu(board, koma) {
    let cand = [];
    for (let i = 0; i < 81; i++) {
        if (board[i] !== EMPTY) {
            continue;
        }

        const x = i % 9;
        const y = Math.floor(i / 9);

        if ((koma === HU && y === 0) ||
            (koma === E_HU && y === 8) ||
            (koma === KYO && y === 0) ||
            (koma === E_KYO && y === 8) ||
            (koma === KEI && y <= 1) ||
            (koma === E_KEI && y >= 7)) {
            continue;
        }
        cand.push([x, y]);
    }
    // 2歩除外
    if (koma === HU || koma === E_HU) {
        for (let j = 0; j < 9; j++) {
            for (let k = 0; k < 9; k++) {
                let idx = k * 9 + j;
                if (board[idx] === koma) {
                    cand = cand.filter(v => v[0] !== j);
                    break;
                }
            }
        }
    }
    return cand;
}

// 動かせるマスの一覧を返す
function move(board, c, r) {
    const koma = board[r * 9 + c];
    const isSente = isSenteKoma(koma);
    const dirs = [];
    switch (koma) {
        case HU:
        case E_HU:
            dirs.push(dir.up(isSente));
            break;
        case KYO:
        case E_KYO:
            dirs.push(dir.up(isSente));
            // 飛角香例外
            return searchHikakuKyo(board, dirs, c, r);
        case KEI:
        case E_KEI:
            dirs.push(dir.keileft(isSente));
            dirs.push(dir.keiright(isSente));
            break;
        case GIN:
        case E_GIN:
            dirs.push(dir.leftup(isSente));
            dirs.push(dir.up(isSente));
            dirs.push(dir.rightup(isSente));
            dirs.push(dir.rightdown(isSente));
            dirs.push(dir.leftdown(isSente));
            break;
        case KIN:
        case E_KIN:
        case RHU:
        case E_RHU:
        case RKYO:
        case E_RKYO:
        case RKEI:
        case E_RKEI:
        case RGIN:
        case E_RGIN:
            dirs.push(dir.left(isSente));
            dirs.push(dir.leftup(isSente));
            dirs.push(dir.up(isSente));
            dirs.push(dir.rightup(isSente))
            dirs.push(dir.right(isSente));
            dirs.push(dir.down(isSente));
            break;
        case HI:
        case E_HI:
            dirs.push(dir.left(isSente));
            dirs.push(dir.up(isSente));
            dirs.push(dir.right(isSente));
            dirs.push(dir.down(isSente));
            // 飛角香例外
            return searchHikakuKyo(board, dirs, c, r);
        case KAKU:
        case E_KAKU:
            dirs.push(dir.leftup(isSente));
            dirs.push(dir.leftdown(isSente));
            dirs.push(dir.rightup(isSente));
            dirs.push(dir.rightdown(isSente));
            // 飛角香例外
            return searchHikakuKyo(board, dirs, c, r);
        case RHI:
        case E_RHI:
            dirs.push(dir.left(isSente));
            dirs.push(dir.leftup(isSente));
            dirs.push(dir.up(isSente));
            dirs.push(dir.rightup(isSente))
            dirs.push(dir.right(isSente));
            dirs.push(dir.rightdown(isSente));
            dirs.push(dir.down(isSente));
            dirs.push(dir.leftdown(isSente));
            // 竜馬例外
            return searchRyuUma(board, dirs, c, r);
        case RKAKU:
        case E_RKAKU:
            dirs.push(dir.left(isSente));
            dirs.push(dir.leftup(isSente));
            dirs.push(dir.up(isSente));
            dirs.push(dir.rightup(isSente))
            dirs.push(dir.right(isSente));
            dirs.push(dir.rightdown(isSente));
            dirs.push(dir.down(isSente));
            dirs.push(dir.leftdown(isSente));
            // 竜馬例外
            return searchRyuUma(board, dirs, c, r);
        case OU:
        case E_OU:
            dirs.push(dir.left(isSente));
            dirs.push(dir.leftup(isSente));
            dirs.push(dir.up(isSente));
            dirs.push(dir.rightup(isSente))
            dirs.push(dir.right(isSente));
            dirs.push(dir.rightdown(isSente));
            dirs.push(dir.down(isSente));
            dirs.push(dir.leftdown(isSente));
            break;
    }
    return search(board, dirs, c, r);
}



function init() {
    const cvs = document.querySelector("canvas");
    const sprite = document.querySelector("#sprite");
    const mainStr = cvs.dataset["main"];
    const tegomaStr = cvs.dataset["tegoma"];
    const main = JSON.parse(mainStr);
    const tegoma = JSON.parse(tegomaStr);
    const game = new GameHistory(main, tegoma);
    const mycvs = new Canvas(cvs, sprite);
    // const current = game.get();
    mycvs.render(game);

    cvs.addEventListener("click", e => {
        const x = e.offsetX;
        const y = e.offsetY;
        click(mycvs, game, x, y);
    });

    form.init(game, mycvs);
}

window.addEventListener("load", init);