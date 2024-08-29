
export default class GameHistory {
    /**@type {number[][]} */
    main;
    /**@type {{[key:string]:number}[]} */
    tegoma
    at = 0;

    constructor(main, tegoma) {
        if (main.length !== tegoma.length) {
            throw new Error(`main.length !== tegoma.length: main.length:${main.length} tegoma.length:${tegoma.length}`);
        }
        this.main = main;
        this.tegoma = tegoma;
    }

    next() {
        const ni = this.at + 1;
        if (ni >= this.main.length) {
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

    /**直近ゲーム版のreferenceを返す */
    get() {
        // return this.data[this.at];
        return {
            main: this.main[this.at],
            tegoma: this.tegoma[this.at],
        };
    }

}