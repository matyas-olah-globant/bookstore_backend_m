CREATE DATABASE IF NOT EXISTS bookstore;

USE bookstore;

-- genres
CREATE TABLE IF NOT EXISTS genres (
    id INT NOT NULL PRIMARY KEY,
    genre TEXT NOT NULL
);
INSERT IGNORE INTO genres VALUES (1, 'Adventure'), (2, 'Classics'), (3, 'Fantasy');

-- books
DROP TABLE IF EXISTS books;
CREATE TABLE books (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name TEXT NOT NULL, -- book title
    price DOUBLE NOT NULL,
    genre_id INT NOT NULL,
    amount INT NOT NULL,
    CONSTRAINT fk_genre FOREIGN KEY (genre_id)
        REFERENCES genres (id)
        ON UPDATE CASCADE
        ON DELETE RESTRICT
);
