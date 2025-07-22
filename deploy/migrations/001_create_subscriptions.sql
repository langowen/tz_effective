CREATE TABLE IF NOT EXISTS subscriptions (
    id SERIAL PRIMARY KEY,
    service_name VARCHAR(255) NOT NULL,
    price BIGINT NOT NULL,
    user_id UUID NOT NULL,
    start_date VARCHAR(7) NOT NULL,
    end_date VARCHAR(7)
);
