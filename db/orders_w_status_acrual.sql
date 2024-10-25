SELECT o.number, o.accrual, s.status_name FROM orders o INNER JOIN order_status s ON o.status_id = s.status_id;
SELECT * from orders_delayed;