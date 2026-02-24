-- +goose Up

ALTER TABLE kanban_board_tickets RENAME TO board_tickets;

-- +goose Down

ALTER TABLE board_tickets RENAME TO kanban_board_tickets;
