-- Sample queries, grouped by pattern. scripts/run-samples.sh formats each one.
-- A "-- pattern: <name>" line starts a group. A blank line ends a query,
-- so one query can span several lines. Other "--" lines are ignored.

-- pattern: select_basics
select id, first_name as fname, last_name lname from users

select distinct country from users

select COUNT(*) as total, SUM(amount) as revenue, avg(amount) from orders where status = 'paid'

-- pattern: where
select * from orders where status='paid' and amount>100 and created_at >= '2026-01-01'

select * from t where a = 1 and not ((b = 2 and c = 3) or d = 4)

select * from users where country not in ( 'US', 'CA' ) and amount between 10 and 20

-- pattern: group_having_order
select country, city, count(*) as n from users where active = 1 group by country, city

select user_id, sum(amount) from orders group by user_id having count(*) > 10 and sum(amount) < 1000

select id, amount from orders order by amount desc, id asc limit 10 offset 20

-- pattern: subquery
select * from users where id in (select user_id from orders where amount > 100)

select u.id, (select count(*) from orders o where o.user_id = u.id) as order_count from users u

select t.user_id, t.total from (select user_id, sum(amount) as total from orders group by user_id) t where t.total > 1000

-- pattern: cte
with paid_orders as (select user_id, amount from orders where status = 'paid') select user_id, sum(amount) as total from paid_orders group by user_id

with daily as (select created_at, sum(amount) as total from orders group by created_at), ranked as (select created_at, total from daily where total > 100) select * from ranked order by total desc limit 10

-- pattern: joins
select u.id, p.amount from users u join orders o on o.user_id = u.id left join payments p on p.order_id = o.id

select * from users u join orders o on o.user_id = u.id and (o.status = 'paid' or o.status = 'shipped')

SELECT u.id, o.amount
FROM users u
LEFT JOIN orders o ON o.user_id = u.id
WHERE u.active = 1

-- pattern: insert_queries
insert   into shop.users(id,name,email,score) values(1,'Ann',NULL,-3)

insert into users (id, name) values (1, 'ann'), (2, 'bob'), (3, upper('cy'))

insert into archive (id, amount) select o.id, o.amount from orders o join users u on u.id = o.user_id where o.status = 'done'

-- pattern: update_queries
UPDATE users SET name='Ann', email = 'a@x.io',age=30 WHERE id = 7

update orders set status = 'vip' from users u where orders.user_id = u.id and u.tier = 'gold'

with vip as (select user_id from orders group by user_id having sum(amount) > 1000) update users set tier = 'vip' where id in (select user_id from vip)
