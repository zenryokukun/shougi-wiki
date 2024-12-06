/** 
 * edit-descriptionと共通している部分が多いので、スクリプトを拝借している。
 * そのままでは動かず、手を入れている箇所があるので注意。
 * とりあえず、別ページとロジックを共有していることは念頭にいれて修正すること。
 */
import { Canvas, GameHistory, click } from "../edit-description/logic.js";
import { init as initIcon } from "../edit-description/icon.js";
import { renderKihu, updateIcon } from "../edit-description/icon.js";
import { init as initForm } from "../edit-description/form.js";

/**
 * 棋譜、タイトル、解説が全て同じかチェックする
 * 盤面は比較不要（棋譜でチェックしているため）
 * game.extractKihuのオーバーライドの後に呼び出す必要がある 
 * */
const checker = {
    init: function (game) {
        const expNode = document.querySelector("#explanation");
        const titleNode = document.querySelector("#title")
        this.kihus = game.extractKihu();
        this.iniExp = expNode.value;
        this.iniTitle = titleNode.value;
    },
    isSame: function (game) {
        const expNode = document.querySelector("#explanation");
        const titleNode = document.querySelector("#title");
        const exp = expNode.value;
        const title = titleNode.value;
        const kihus = game.extractKihu();
        let isSameKihu = true;
        if (this.kihus.length != kihus.length) {
            isSameKihu = false;
        } else {
            for (let i = 0; i < this.kihus.length; i++) {
                if (this.kihus[i] !== kihus[i]) {
                    isSameKihu = false;
                    break;
                }
            }
        }

        return isSameKihu && this.iniExp === exp && this.iniTitle === title;
    }
}

function initHistory() {
    const recs = document.querySelectorAll("tbody tr");

    function loadHistory(id, seq) {
        // tableタグ内にaタグは入れられない仕様とのこと。
        // そのためtrタグクリックで画面遷移するようにする。
        const url = `/undo/?id=${id}&seq=${seq}`;
        // 別のwindowで開きたい場合はwindow.openを使えばいける
        window.location.href = url;
    }

    for (let i = 0; i < recs.length; i++) {
        const rec = recs[i];
        rec.addEventListener("click", () => {
            // 履歴レコード（data-restored="true"の属性を持つ場合）は戻しの対象外
            const isRestored = rec.dataset.restored === "true";
            if (isRestored) {
                alert("この行は復元できません。他の行を選択してください。")
                return;
            }
            // 履歴レコードでない場合
            const id = rec.dataset.id;
            const seq = rec.dataset.seq;
            const isRestore = confirm(`過去の内容に戻します。よろしいですか？\nOKを押すとプレビュー画面が開きます。問題なければプレビュー画面で確定を押してください。`);
            if (isRestore) {
                loadHistory(id, seq);
            }
        });
    }
}

function init() {
    const cvs = document.querySelector("canvas");
    const sprite = document.querySelector("#sprite");
    const mainStr = cvs.dataset["main"];
    const tegomaStr = cvs.dataset["tegoma"];
    const goteTegomaStr = cvs.dataset["gotetegoma"];
    const kihuStr = cvs.dataset["kihu"];
    const main = JSON.parse(mainStr);
    const tegoma = JSON.parse(tegomaStr);
    const goteTegoma = JSON.parse(goteTegomaStr);
    const kihu = JSON.parse(kihuStr);
    // GameHistoryを手動で初期化していく
    const game = new GameHistory(main[0], tegoma[0]);
    for (let i = 1; i < main.length; i++) {
        game.add(main[i], tegoma[i], goteTegoma[i], kihu[i]);
    }

    game.at = 0; // 最初に戻しておく
    const mycvs = new Canvas(cvs, sprite);
    mycvs.render(game);

    initForm(game, true, checker);

    // renderKihu内でgame.extractKihuを呼び出しているが、
    // game.atの位置に基づいて棋譜を出力している。上でgame.atを0に戻しているので、
    // うまく出力されない。「revise」ではgame.atに基づいて描写する必要がないため、
    // extractKihuをオーバーライドする。
    game.extractKihu = () => {
        const kihus = [];
        for (const d of game.data) {
            if (!d.kihu || d.kihu.length === 0) {
                continue;
            }
            kihus.push(d.kihu);
        }
        return kihus;
    };

    // 入力内容の初期値をセット。
    // game.extractKihuのオーバーライド後に呼ぶ必要がある
    checker.init(game);

    renderKihu(game);
    initIcon(game, mycvs);
    // 初期表示時に「＞」アイコンを活性化させるために必要
    updateIcon(game);

    // 編集履歴を引き込むための処理を初期化
    initHistory();

    cvs.addEventListener("click", e => {
        const x = e.offsetX;
        const y = e.offsetY;
        click(mycvs, game, x, y);
    })

    // 初期表示では１手目がハイライトされてしまうため、ハイライトを手動で削除する。
    // renderKihu内で呼び出しているhightlight関数が原因。
    // ２度手間だが、edit-descriptionと共通して使うのでやむなし。
    const node = document.querySelector(".highlight");
    node.classList.remove("highlight");

}
window.addEventListener("load", init);