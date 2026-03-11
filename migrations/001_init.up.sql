create table errors(
	error_id serial primary key,
	cause varchar(255) not null,
	row_and_column text
);

create table orders(
	id serial primary key,
	start_date timestamp not null default NOW(),
	end_date timestamp,
	total_cost decimal(15, 2) not null default 0 check(total_cost >= 0),
	status varchar(20) not null,
	error_id integer references errors(error_id) on delete set null
);

create table overhead(
	id serial primary key,
	order_id integer not null references orders(id) on delete cascade,
	date timestamp not null,
	prod_type varchar(50) not null,
	amount decimal(15, 2) not null check(amount >= 0)
);

create table boms(
	id serial primary key,	
	order_id integer not null references orders(id) on delete cascade,
	quantity decimal(15, 3) not null check(quantity > 0),
	unit_cost decimal(15, 2) not null check(unit_cost >= 0),
	material_code varchar(10) not null
);

create table labor(
	id serial primary key,
	order_id integer not null references orders(id) on delete cascade,
	rate decimal(15, 2) not null check(rate >= 0),
	hours decimal(10, 2) not null check(hours >= 0)
);