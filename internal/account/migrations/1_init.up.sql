create table accounts (
    id INTEGER constraint accounts_pk primary key autoincrement,
    balance INTEGER default 0 not null,
    currency TEXT default 'SBP' not null
);
INSERT INTO accounts (balance)
VALUES (0);