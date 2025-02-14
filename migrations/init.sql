CREATE TABLE users (
                       id SERIAL PRIMARY KEY,
                       username VARCHAR(255) UNIQUE NOT NULL,
                       password_hash TEXT NOT NULL,
                       coins INT DEFAULT 0
);

CREATE TABLE inventory (
                           id SERIAL PRIMARY KEY,
                           user_id INT NOT NULL,
                           item_type VARCHAR(255) NOT NULL,
                           quantity INT DEFAULT 0,
                           FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE coin_transactions (
                                   id SERIAL PRIMARY KEY,
                                   sender_id INT,
                                   receiver_id INT,
                                   amount INT NOT NULL,
                                   timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                   FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE SET NULL,
                                   FOREIGN KEY (receiver_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE TABLE items (
                       id SERIAL PRIMARY KEY,
                       name VARCHAR(255) UNIQUE NOT NULL,
                       price INT NOT NULL
);

INSERT INTO items(name, price) VALUES
('t-shirt', 80), ('cup', 20),
('book', 50), ('pen', 10),
('powerbank', 200), ('hoody', 300),
('umbrella', 200), ('socks', 10),
('wallet', 50), ('pink-hoody', 500);

CREATE INDEX idx_inventory_user_item ON inventory(user_id, item_type);
CREATE INDEX idx_coin_transactions_sender ON coin_transactions(sender_id);
CREATE INDEX idx_coin_transactions_receiver ON coin_transactions(receiver_id);
