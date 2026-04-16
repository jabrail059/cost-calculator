create table upload_log(
    id serial primary key,
    order_id integer not null references orders(id) on delete cascade,
    file_type varchar(10) not null check (file_type in ('boms', 'labor', 'overhead')),
    uploaded_at timestamp with time zone default NOW(),
    changed_by text not null default 'unknown'
);