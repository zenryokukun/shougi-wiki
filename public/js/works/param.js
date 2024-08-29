export default function getParamId() {
    const param = new URLSearchParams(window.location.search);
    return param.get("id");
}