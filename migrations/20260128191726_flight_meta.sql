-- +goose Up
-- +goose StatementBegin
CREATE TABLE flight_meta (
    id SERIAL PRIMARY KEY,
    flight_number VARCHAR(20) NOT NULL,
    departure_date TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(20) DEFAULT 'scheduled',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP WITH TIME ZONE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE flight_meta;
-- +goose StatementEnd
