-- 詰将棋の作品を登録するDB。全て直近の内容
CREATE TABLE WORKS(
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    -- 手数
    TESU INTEGER,
    TITLE TEXT,
    KIHU TEXT,
    EXPLANATION TEXT,
    AUTHOR TEXT,
    EDITOR TEXT,
    MAIN TEXT,
    TEGOMA TEXT,
    GOTETEGOMA TEXT,
    PUBLISH_DATE INTEGER,
    EDIT_DATE INTEGER,
    GOOD INTEGER,
    BAD INTEGER,
    -- 修正要望数
    DEMAND INTEGER,
    -- バックアップから復元した時に設定される
    BACKUP_SEQ INTEGER,
    -- 編集した時に記録される修正内容
    COMMENT TEXT,
    -- 削除した時に設定される
    DEL_FLG INTEGER,
    DEL_REASON TEXT,
    DEL_BY TEXT,
    DEL_DATE INTEGER
);