function example() {
    const cand = [
        "同歩", "5六飛打", "2二成香", "7一桂成", "3四銀不成", "8七金右", "7五銀上"
    ];
    const rnd = Math.random() * cand.length;
    const i = Math.floor(rnd);
    return "例）" + cand[i];
}

function formEdit(e) {
    const parent = document.querySelector(".answers-wrapper");
    const te = document.querySelector("select[name='te']");
    const long = document.querySelector("input[name='long']");
    let steps = te.value;

    te.addEventListener("change", function (e) {
        if (steps === te.value) return;
        if (te.value === "other") {
            long.disabled = false;
            return;
        }

        long.value = "";
        long.disabled = true;

        steps = this.value;
        while (parent.firstChild) {
            parent.removeChild(parent.firstChild);
        }
        for (let i = 0; i < steps; i++) {
            const div = document.createElement("div");
            const label = document.createElement("label");
            label.htmlFor = `step${i + 1}`;
            const input = document.createElement("input");
            input.id = `step${i + 1}`;
            input.type = "text";
            input.className = "answers-template";
            input.name = "answers";
            input.placeholder = example();
            let txt = i % 2 === 0 ? "☗" : "☖";
            txt += (i + 1).toString() + "手目";

            label.textContent = txt;
            div.appendChild(label);
            div.appendChild(input);
            parent.appendChild(div);
        }
    });
}

window.addEventListener("load", formEdit);