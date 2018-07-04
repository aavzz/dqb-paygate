
create table payments (
   id                   bigserial    not null unique,
   rcpt_id              bigint       unique references receipts(id),
   payment_id           uuid         not null unique,
   payment_subject_id   varchar(20)  not null,
   payment_sum          money        not null check (payment_sum > 0.00::money),
   payment_vat          varchar(10)  not null default '',
   payment_direction    varchar(10)  not null check (payment_direction in ('in', 'out')),
   payment_channel      varchar(20)  not null,
   notification_sent_to varchar(100),
   channel_payment_id   varchar(100) not null,
   channel_terminal_id  varchar(20),
   tstamp_paygate       timestamp    not null default current_timestamp,
   --tstamp_ofd           timestamp,
   tstamp_billing       timestamp,
   tstamp_notification  timestamp,
   unique (payment_channel,channel_payment_id)
);



create table receipts (
   id         bigserial    not null unique,
   payment_id bigint       not null unique references payments(id),
   rcpt_id    uuid         not null unique,
   tstamp     timestamp    not null default current_timestamp,
   status     varchar(20)  not null,
   error      varchar(200)
);

