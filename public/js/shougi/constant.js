export const BG = "#ffe599";
export const BLACK = "#000000";
export const BLOCK = 50;
export const TEGOMA = BLOCK;
export const PADDING = 20;
export const ARC = 4;

// 躍動するこまた、駒たち。spriteと同じに並びにする必要ある。
export const OU = 13, HU = 12, KYO = 11, KEI = 10, GIN = 9, KIN = 8, KAKU = 7, HI = 6, RHU = 5, RKYO = 4, RKEI = 3, RGIN = 2, RKAKU = 1, RHI = 0, E_OU = 27, E_HU = 26, E_KYO = 25, E_KEI = 24, E_GIN = 23, E_KIN = 22, E_KAKU = 21, E_HI = 20, E_RHU = 19, E_RKYO = 18, E_RKEI = 17, E_RGIN = 16, E_RKAKU = 15, E_RHI = 14;
// 空白マス
export const EMPTY = -1;
// 先手、後手
export const SENTE = 0;
export const GOTE = 1;
// 駒名
export const KOMA_NAME = {
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