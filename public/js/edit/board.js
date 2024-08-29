
const EMPTY = -1

export const board = {
    /**@type {number[]} - 将棋盤*/
    main: [],

    /**@type {{[key:string]:number}} */
    tegoma: {},

    init: function () {
        const data = []
        for (let i = 0; i < 81; i++) {
            data.push(EMPTY);
        }
        this.main = data;
        this.tegoma = {};
    },

    /**
     * @param {number[]} newBoard 
     * @param {{[key:string]:number}} tegoma 
     */
    new: function (newBoard, tegoma) {
        this.main = newBoard;
        this.tegoma = tegoma || {};
    },

    getPos: function (i) {
        const row = Math.floor(i / 9);
        const col = i % 9;
        return [col, row];
    },

    /**
     * @param {number} koma 
     * @param {number} col 
     * @param {number} row 
     */
    set: function (koma, col, row) {
        const i = row * 9 + col;
        this.main[i] = koma;
    },

    get: function (col, row) {
        const i = row * 9 + col;
        return this.main[i];
    },

    /**@param {number} koma */
    setTegoma: function (koma) {
        const tegoma = this.tegoma;
        if (tegoma[koma] === undefined) {
            tegoma[koma] = 1;
        } else {
            tegoma[koma] += 1;
        }
    },

    getTegoma: function () {
        return Object.entries(this.tegoma);
    },

    reduceTegoma: function (koma) {
        const tegoma = this.tegoma;
        if (tegoma[koma] === undefined) {
            return;
        }
        tegoma[koma] -= 1;
        if (tegoma[koma] === 0) {
            delete tegoma[koma];
        }
    },

};