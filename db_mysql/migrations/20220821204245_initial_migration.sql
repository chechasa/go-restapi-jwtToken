-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS persons(
    id int NOT NULL PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100),
    age int,
    family VARCHAR(50),
    username VARCHAR(30) UNIQUE,
    password VARCHAR(30),
    role VARCHAR(20) NOT NULL DEFAULT 'normal'
);

INSERT INTO persons VALUES
(1, "Maickel", 20, "Cardoso", "first", "1234", "admin");


-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

DROP TABLE persons;

