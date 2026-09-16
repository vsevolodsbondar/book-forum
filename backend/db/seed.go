package db

import (
	"database/sql"
	"fmt"
)

func SeedDB(db *sql.DB) error {
	query := `
		-- =========================
		-- USERS
		-- =========================

		INSERT INTO user (id, user_name, profile_picture, name, description)
		VALUES
	    (1, 'alice', 'alice.jpg', 'Alice Johnson',
	     'Backend developer and book lover.'),

	    (2, 'bob', 'bob.jpg', 'Bob Smith',
	     'I enjoy science fiction and fantasy.'),

	    (3, 'charlie', NULL, 'Charlie Brown',
	     'Software engineer. Coffee enthusiast.'),

	    (4, 'diana', 'diana.jpg', 'Diana Wilson',
	     'Mostly reading classics and historical fiction.'),

	    (5, 'eve', NULL, 'Eve Davis',
	     'Trying to read at least one book every month.'),

	    (6, 'frank', NULL, 'Frank Miller',
	     'Fantasy, games and everything in between.');


		-- =========================
		-- CATEGORIES
		-- =========================

		INSERT INTO category (id, name)
		VALUES
	    (1, 'Fantasy'),
	    (2, 'Science Fiction'),
	    (3, 'Mystery'),
	    (4, 'Classics'),
	    (5, 'Non-Fiction');


		-- =========================
		-- POSTS
		-- =========================

		INSERT INTO post
	    (id, title, author_id, category_id, init_comment_id)
		VALUES
	    (1,
	     'What fantasy book should I read next?',
	     1,
	     1,
	     NULL),

	    (2,
	     'Best science fiction books of the last decade',
	     2,
	     2,
	     NULL),

	    (3,
	     'Books that completely changed your perspective',
	     3,
	     5,
	     NULL),

	    (4,
	     'What is your favorite classic novel?',
	     4,
	     4,
	     NULL),

	    (5,
	     'Looking for a good mystery novel',
	     5,
	     3,
	     NULL),

	    (6,
	     'The best fantasy world you have ever read about',
	     6,
	     1,
	     NULL);


		-- =========================
		-- COMMENTS
		-- =========================

		-- Post 1
		INSERT INTO comment
	    (id, text, post_id, parent_comment_id, user_id)
		VALUES
	    (1,
	     'I highly recommend The Name of the Wind.',
	     1,
	     NULL,
	     2),

	    (2,
	     'I agree! The world building is fantastic.',
	     1,
	     1,
	     3),

	    (3,
	     'You might also enjoy Mistborn.',
	     1,
	     NULL,
	     4),

	    (4,
	     'Mistborn is a great recommendation, especially if you like magic systems.',
	     1,
	     3,
	     1);


		-- Post 2
		INSERT INTO comment
	    (id, text, post_id, parent_comment_id, user_id)
		VALUES
	    (5,
	     'Project Hail Mary is probably my favorite recent sci-fi book.',
	     2,
	     NULL,
	     1),

	    (6,
	     'Same here. The science was surprisingly interesting.',
	     2,
	     5,
	     5),

	    (7,
	     'I would also recommend Children of Time.',
	     2,
	     NULL,
	     6);


		-- Post 3
		INSERT INTO comment
	    (id, text, post_id, parent_comment_id, user_id)
		VALUES
	    (8,
	     'For me it was Sapiens. It changed how I look at human history.',
	     3,
	     NULL,
	     4),

	    (9,
	     'That book definitely gives you a lot to think about.',
	     3,
	     8,
	     3),

	    (10,
	     'Meditations by Marcus Aurelius had a similar effect on me.',
	     3,
	     NULL,
	     2);


		-- Post 4
		INSERT INTO comment
	    (id, text, post_id, parent_comment_id, user_id)
		VALUES
	    (11,
	     'Definitely The Count of Monte Cristo.',
	     4,
	     NULL,
	     6),

	    (12,
	     'Great choice. It is a huge book but absolutely worth it.',
	     4,
	     11,
	     5),

	    (13,
	     'Mine would be Crime and Punishment.',
	     4,
	     NULL,
	     1);


		-- Post 5
		INSERT INTO comment
	    (id, text, post_id, parent_comment_id, user_id)
		VALUES
	    (14,
	     'Try And Then There Were None by Agatha Christie.',
	     5,
	     NULL,
	     2),

	    (15,
	     'That is one of my favorites!',
	     5,
	     14,
	     4),

	    (16,
	     'The Murder of Roger Ackroyd is also worth reading.',
	     5,
	     NULL,
	     6);


		-- Post 6
		INSERT INTO comment
	    (id, text, post_id, parent_comment_id, user_id)
		VALUES
	    (17,
	     'Middle-earth is still my favorite fantasy world.',
	     6,
	     NULL,
	     1),

	    (18,
	     'For me it is the Cosmere.',
	     6,
	     NULL,
	     3),

	    (19,
	     'The Cosmere is enormous. I still have a lot of books to read.',
	     6,
	     18,
	     5);


		-- =========================
		-- INITIAL COMMENT IDS
		-- =========================

		UPDATE post
		SET init_comment_id = 1
		WHERE id = 1;

		UPDATE post
		SET init_comment_id = 5
		WHERE id = 2;

		UPDATE post
		SET init_comment_id = 8
		WHERE id = 3;

		UPDATE post
		SET init_comment_id = 11
		WHERE id = 4;

		UPDATE post
		SET init_comment_id = 14
		WHERE id = 5;

		UPDATE post
		SET init_comment_id = 17
		WHERE id = 6;


		-- =========================
		-- LIKES
		-- type_of_like:
		-- 1 = like
		-- 0 = dislike
		-- =========================

		INSERT INTO likes (user_id, comment_id, type_of_like)
		VALUES
	    -- Comment 1
	    (1, 1, 1),
	    (3, 1, 1),
	    (4, 1, 0),

	    -- Comment 2
	    (1, 2, 1),
	    (5, 2, 1),

	    -- Comment 3
	    (2, 3, 1),
	    (6, 3, 1),

	    -- Comment 5
	    (3, 5, 1),
	    (4, 5, 1),
	    (6, 5, 0),

	    -- Comment 6
	    (1, 6, 1),

	    -- Comment 8
	    (2, 8, 1),
	    (5, 8, 1),

	    -- Comment 10
	    (3, 10, 1),
	    (6, 10, 0),

	    -- Comment 11
	    (1, 11, 1),
	    (2, 11, 1),
	    (5, 11, 1),

	    -- Comment 14
	    (3, 14, 1),
	    (4, 14, 1),

	    -- Comment 17
	    (2, 17, 1),
	    (3, 17, 1),
	    (4, 17, 1),

	    -- Comment 18
	    (1, 18, 1),
	    (6, 18, 1);
	`
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin seed transaction: %w", err)
	}

	defer tx.Rollback()

	_, err = tx.Exec(query)
	if err != nil {
		return fmt.Errorf("seeding failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit seed transaction: %w", err)
	}

	fmt.Println("Database seeded")
	return nil
}
