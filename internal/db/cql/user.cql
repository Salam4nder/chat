CREATE TABLE chat.user_in_room (
    user_id uuid,
    room_id uuid,
    PRIMARY KEY (user_id, room_id)
);
