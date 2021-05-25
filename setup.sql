CREATE DATABASE IF NOT EXISTS bookstore;

USE bookstore;

CREATE TABLE IF NOT EXISTS genres (
    id INT NOT NULL PRIMARY KEY,
    genre TEXT NOT NULL
);

INSERT IGNORE INTO genres (id, genre) VALUES (1, 'Adventure');
INSERT IGNORE INTO genres (id, genre) VALUES (2, 'Classics');
INSERT IGNORE INTO genres (id, genre) VALUES (3, 'Fantasy');

CREATE TABLE IF NOT EXISTS books (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name TEXT NOT NULL, -- title
    price DOUBLE NOT NULL,
    genre_id INT NOT NULL,
    amount INT NOT NULL,
    CONSTRAINT fk_genre FOREIGN KEY (genre_id)
        REFERENCES genres (id)
        ON UPDATE CASCADE
        ON DELETE RESTRICT
);
