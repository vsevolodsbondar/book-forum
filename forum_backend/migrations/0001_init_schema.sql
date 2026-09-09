
CREATE TABLE IF NOT EXISTS user(
    id INTEGER PRIMARY KEY,
    user_name TEXT NOT NULL UNIQUE,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    profile_picture TEXT,
    last_seen TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name TEXT,
    description TEXT
);
CREATE TABLE IF NOT EXISTS category(
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS post(
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL,
    author_id INTEGER,
    category_id INTEGER,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    init_comment_id INTEGER,
    FOREIGN KEY(author_id) REFERENCES user(id) ON DELETE SET NULL,
    FOREIGN KEY(category_id) REFERENCES category(id) ON DELETE SET NULL
);
CREATE TABLE IF NOT EXISTS comment(
    id INTEGER PRIMARY KEY,
    text TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    post_id INTEGER NOT NULL,
    parent_comment_id INTEGER,
    author_id INTEGER,
    FOREIGN KEY(post_id) REFERENCES post(id) ON DELETE CASCADE,
    FOREIGN KEY(parent_comment_id) REFERENCES comment(id) ON DELETE SET NULL,
    FOREIGN KEY(author_id) REFERENCES user(id) ON DELETE SET NULL,
    CHECK (parent_comment_id != id)
);
CREATE TABLE IF NOT EXISTS likes(
    id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL,
    comment_id INTEGER NOT NULL,
    type INTEGER NOT NULL,
    FOREIGN KEY(user_id) REFERENCES user(id) ON DELETE CASCADE,
    FOREIGN KEY(comment_id) REFERENCES comment(id) ON DELETE CASCADE,
    UNIQUE(user_id, comment_id),
    CHECK(type IN (0,1))
);
CREATE INDEX idx_comment_post_id ON comment(post_id);
CREATE INDEX idx_post_author_id ON post(author_id);
CREATE INDEX idx_post_category_id ON post(category_id);
CREATE INDEX idx_likes_comment_id ON likes(comment_id);