-- default categories, so the forum can create posts without the seed
INSERT OR IGNORE INTO category (id, name) VALUES
(1, 'Fantasy'),
(2, 'Science Fiction'),
(3, 'Mystery'),
(4, 'Classics'),
(5, 'Non-Fiction');
