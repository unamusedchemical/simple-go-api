CREATE TABLE IF NOT EXISTS User
(
    Id       INTEGER PRIMARY KEY AUTO_INCREMENT NOT NULL,
    Username VARCHAR(60)                        NOT NULL,
    Email    VARCHAR(60)                        NOT NULL,
    Password VARCHAR(72)                        NOT NULL
);


CREATE TABLE IF NOT EXISTS ActivityGroup
(
    Id     INTEGER PRIMARY KEY AUTO_INCREMENT NOT NULL,
    Name   VARCHAR(25)                        NOT NULL,
    UserId INTEGER                            NOT NULL,

    FOREIGN KEY (UserId) REFERENCES User (Id)
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS Activity
(
    Id       INTEGER     NOT NULL PRIMARY KEY AUTO_INCREMENT,
    Title    VARCHAR(60) NOT NULL,
    Body     TEXT        NOT NULL,
    ClosedOn DATETIME,
    OpenedOn DATETIME    NOT NULL,
    Due      DATETIME,
    UserId   INTEGER     NOT NULL,
    GroupId  INTEGER,

    FOREIGN KEY (UserId) REFERENCES User (Id)
        ON DELETE CASCADE,

    FOREIGN KEY (GroupId) REFERENCES ActivityGroup (Id)
        ON DELETE CASCADE
);

CREATE INDEX Opened
ON Activity (OpenedOn);

ALTER TABLE Activity
ADD FULLTEXT(Title, Body);