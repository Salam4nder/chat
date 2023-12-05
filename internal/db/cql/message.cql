CREATE TABLE message.message_by_room (
  id uuid,
  data blob,
  type text,
  sender text,
  room_id text,
  time timestamp,
  PRIMARY KEY (room_id, id)
);
