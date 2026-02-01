-- +goose Up
-- +goose StatementBegin
CREATE TABLE flights (
    flight_number VARCHAR(20) NOT NULL,
    departure_date TIMESTAMP WITH TIME ZONE NOT NULL,
    aircraft_type VARCHAR(50),
    arrival_date TIMESTAMP WITH TIME ZONE,
    passengers_count INTEGER DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (flight_number, departure_date)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE flights;
-- +goose StatementEnd
