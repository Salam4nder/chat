CREATE TYPE message.message (
  id uuid,
  type text,
  data text,
  sender text,
  recipient text,
  time timestamp
);

CREATE TABLE message.message_by_recipient (
  recipient text,
  id uuid,
  message frozen<message>,
  sender text,
  time timestamp,
  PRIMARY KEY (recipient, id)
);

CREATE TABLE message.message_by_sender (
  sender text,
  id uuid,
  message frozen<message>,
  recipient text,
  time timestamp,
  PRIMARY KEY (sender, id)
);

CREATE TABLE message.message_by_sender_recipient (
  sender text,
  recipient text,
  id uuid,
  message frozen<message>,
  time timestamp,
  PRIMARY KEY ((sender, recipient), id)
);

CREATE TABLE message.message_by_time (
  time timestamp,
  id uuid,
  message frozen<message>,
  sender text,
  recipient text,
  PRIMARY KEY (time, id)
);

CREATE TABLE message.message_by_type (
  type text,
  id uuid,
  message frozen<message>,
  sender text,
  recipient text,
  time timestamp,
  PRIMARY KEY (type, id)
);

CREATE TABLE message.message_by_sender_type (
  sender text,
  type text,
  id uuid,
  message frozen<message>,
  recipient text,
  time timestamp,
  PRIMARY KEY ((sender, type), id)
);

CREATE TABLE message.message_by_recipient_type (
  recipient text,
  type text,
  id uuid,
  message frozen<message>,
  sender text,
  time timestamp,
  PRIMARY KEY ((recipient, type), id)
);
