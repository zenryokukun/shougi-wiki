
function validate() {
    const kihus = document.querySelectorAll(".kihu-row");
    const cnt = kihus.length;
    const ret = { ok: true, msg: "" };
    if (cnt === 0) {
        ret.ok = false;
        ret.msg = "棋譜が設定されていません。盤面を操作してください。"
    }
    return ret;
}

function confirm(game) {
    const res = validate();
    if (!res.ok) {
        alert(res.msg);
        return;
    }
    const form = document.querySelector("form");

    // htmlのrequiredのチェックをする。
    if (!form.reportValidity()) return;

    const dataInput = document.querySelector("input[name='data']");
    const main = [];
    const tegoma = [];
    const goteTegoma = [];
    const kihu = [];
    for (const d of game.data) {
        main.push(d.main);
        tegoma.push(d.tegoma);
        goteTegoma.push(d.goteTegoma);
        kihu.push(d.kihu);
    }
    const boardStr = JSON.stringify({ main, tegoma, goteTegoma, kihu });
    dataInput.value = boardStr;
    form.submit();
}

export function init(game, isReviseMode, checker) {
    // 確定ボタンを押下時
    const confButton = document.querySelector(".submit");
    confButton.addEventListener("click", e => {
        if (isReviseMode) {
            // reviseモード時は変更されているかチェックする。
            // 変更されていない場合、エラーを表示して登録には進まない
            const isSame = checker.isSame(game);
            if (isSame) {
                alert("何も変更されていません");
                return;
            }
        }
        confirm(game);
    });

}







