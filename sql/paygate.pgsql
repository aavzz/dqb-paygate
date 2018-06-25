
create table payments (
   id                  bigserial    not null unique,
   payment_subject_id  varchar(20)  not null,
   payment_sum         float        not null,
   payment_direction   varchar(20)  not null check (payment_direction in ('in', 'out')),
   payment_channel     varchar(20)  not null,
   channel_payment_id  varchar(100),
   channel_terminal_id varchar(20),
   tstamp_paygate      timestamp    not null default current_timestamp,
   tstamp_ofd          timestamp,
   tstamp_billing      timestamp,
   unique (channel,channel_payment_id)
);

