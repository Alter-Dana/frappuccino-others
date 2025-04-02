CREATE TYPE status AS ENUM ('open', 'in progress', 'closed');

CREATE TABLE orders (
    id serial primary key,
    customer_name varchar(255) not null,
    order_status status not null,
    created_at timestamp not null default now(),
    customer_preferences jsonb not null default '{}'::jsonb
);
CREATE INDEX idx_orders_customer_name ON orders (customer_name);

CREATE TABLE order_status_history (
    id serial primary key,
    order_id int references orders (id) on delete cascade,
    updated_at timestamp not null,
    old_status status not null,
    new_status status not null
);

CREATE TABLE menu_items (
    id serial primary key,
    name varchar(255) not null unique,
    description varchar(1000) not null,
    tsv tsvector,
    price decimal(10, 2) not null constraint positive_price CHECK (price >= 0)
);
CREATE INDEX idx_menu_items_tsv ON menu_items USING GIN(tsv);

CREATE TABLE price_history (
    id serial primary key,
    menu_item_id int references menu_items (id) on delete cascade,
    old_price decimal(10,2) not null,
    new_price decimal(10,2) not null,
    updated_at timestamp not null
);

CREATE TABLE order_item (
    order_id int references orders (id) on delete cascade,
    menu_item_id int references menu_items (id) on delete cascade,
    quantity int not null constraint positive_quantity CHECK (quantity >= 0)
);

CREATE TYPE unit AS ENUM ('shots', 'ml', 'g', 'units');

CREATE TABLE inventory (
    id serial primary key,
    name varchar(255) not null unique,
    quantity int not null default 0 constraint positive_quantity CHECK (quantity >= 0),
    unit unit not null,
    categories varchar(50)[]
);

CREATE TABLE inventory_transactions (
    id serial primary key,
    inventory_id int references inventory (id) on delete cascade,
    old_quantity int not null,
    new_quantity int not null,
    transaction_date timestamp not null
);

CREATE TABLE menu_item_inventory (
    menu_id int references menu_items (id) on delete cascade,
    inventory_id int references inventory (id) on delete cascade,
    quantity int not null constraint positive_quantity CHECK (quantity >= 0)
);

-- Function for inventory quantity tracking
CREATE OR REPLACE FUNCTION log_inventory_transaction()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.quantity <> NEW.quantity THEN
        INSERT INTO inventory_transactions (inventory_id, old_quantity, new_quantity, transaction_date)
        VALUES (NEW.id, OLD.quantity, NEW.quantity, NOW());
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to track quantity updates in inventory
CREATE TRIGGER after_inventory_update
AFTER UPDATE ON inventory
FOR EACH ROW
EXECUTE FUNCTION log_inventory_transaction();


-- Function for price change tracking
CREATE OR REPLACE FUNCTION log_price_change()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.price <> NEW.price THEN
        INSERT INTO price_history (menu_item_id, old_price, new_price, updated_at)
        VALUES (NEW.id, OLD.price, NEW.price, NOW());
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to track price updates in menu_items
CREATE TRIGGER after_price_update
AFTER UPDATE ON menu_items
FOR EACH ROW
EXECUTE FUNCTION log_price_change();


-- Function for order status tracking
CREATE OR REPLACE FUNCTION log_order_status_change()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.order_status <> NEW.order_status THEN
        INSERT INTO order_status_history (order_id, updated_at, old_status, new_status)
        VALUES (NEW.id, NOW(), OLD.order_status, NEW.order_status);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to track status updates in orders
CREATE TRIGGER after_order_status_update
AFTER UPDATE ON orders
FOR EACH ROW
EXECUTE FUNCTION log_order_status_change();

CREATE OR REPLACE FUNCTION set_menu_items_tsv() 
RETURNS trigger AS $$
BEGIN
  NEW.tsv := setweight(to_tsvector('english', NEW.name), 'A') ||
             setweight(to_tsvector('english', NEW.description), 'B');
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER menu_items_tsv_trigger
BEFORE INSERT OR UPDATE ON menu_items
FOR EACH ROW
EXECUTE FUNCTION set_menu_items_tsv();

INSERT INTO inventory (name, quantity, unit, categories) VALUES
('Espresso Shot', 500, 'shots', ARRAY['Beverage']),
('Milk', 5000, 'ml', ARRAY['Dairy']),
('Flour', 10000, 'g', ARRAY['Baking']),
('Blueberries', 2000, 'g', ARRAY['Fruit']),
('Raspberry', 2000, 'g', ARRAY['Fruit']),
('Sugar', 5000, 'g', ARRAY['Baking', 'Sweetener']),
('Coffee Beans', 5000, 'g', ARRAY['Beverage', 'Raw Material']),
('Ground Coffee', 3000, 'g', ARRAY['Beverage']),
('Vanilla Syrup', 2000, 'ml', ARRAY['Flavoring']),
('Caramel Syrup', 2000, 'ml', ARRAY['Flavoring']),
('Chocolate Syrup', 2500, 'ml', ARRAY['Flavoring']),
('Whipped Cream', 1000, 'ml', ARRAY['Dairy', 'Topping']),
('Tea Leaves', 1500, 'g', ARRAY['Beverage', 'Raw Material']),
('Honey', 1000, 'ml', ARRAY['Sweetener', 'Flavoring']),
('Pastry Dough', 5000, 'g', ARRAY['Baking']),
('Butter', 2000, 'g', ARRAY['Dairy']),
('Eggs', 300, 'units', ARRAY['Baking', 'Dairy']),
('Cinnamon', 1500, 'g', ARRAY['Spice']),
('Nutmeg', 1000, 'g', ARRAY['Spice']),
('Matcha Powder', 800, 'g', ARRAY['Tea', 'Flavoring']),
('Ice Cubes', 3000, 'units', ARRAY['Cooling']),
('Hazelnut Syrup', 1000, 'ml', ARRAY['Flavoring']);


INSERT INTO menu_items (name, description, price) VALUES
('Blueberry Muffin', 'Freshly baked muffin with blueberries', 2.00),
('Raspberry Muffin', 'Muffin with fresh raspberries', 2.00),
('Strawberry Muffin', 'Freshly baked muffin with strawberries', 2.00),
('Caffe Latte', 'Espresso with steamed milk', 3.50),
('Espresso', 'A strong shot of coffee', 2.00),
('Vanilla Cappuccino', 'Espresso with vanilla syrup and foam', 3.80),
('Caramel Macchiato', 'Espresso with caramel syrup and steamed milk', 4.20),
('Chocolate Frappe', 'Blended chocolate drink with whipped cream', 4.50),
('Matcha Latte', 'Green tea with steamed milk', 3.60),
('Chai Tea Latte', 'Spiced tea with milk', 3.70),
('Barista Special', 'Rich espresso with hazelnut syrup and cream', 4.60),
('Ice Latte', 'Chilled espresso with milk and ice cubes', 4.10),
('Double Espresso', 'Two strong espresso shots', 3.20);


-- Blueberry Muffin
INSERT INTO menu_item_inventory (menu_id, inventory_id, quantity) VALUES
(1, 3, 100),  -- Flour
(1, 4, 50),   -- Blueberries
(1, 6, 10),   -- Sugar
(1, 15, 100), -- Pastry Dough
(1, 16, 20);  -- Butter

-- Raspberry Muffin
INSERT INTO menu_item_inventory (menu_id, inventory_id, quantity) VALUES
(2, 3, 100),
(2, 5, 50),
(2, 6, 10),
(2, 15, 100),
(2, 16, 20);

-- Strawberry Muffin
INSERT INTO menu_item_inventory (menu_id, inventory_id, quantity) VALUES
(3, 3, 100),
(3, 1, 30),
(3, 10, 20),
(3, 2, 100),
(3, 16, 20);

-- Caffe Latte
INSERT INTO menu_item_inventory (menu_id, inventory_id, quantity) VALUES
(4, 1, 1),    -- Espresso Shot
(4, 2, 200);  -- Milk

-- Espresso
INSERT INTO menu_item_inventory (menu_id, inventory_id, quantity) VALUES
(5, 1, 1);

-- Vanilla Cappuccino
INSERT INTO menu_item_inventory (menu_id, inventory_id, quantity) VALUES
(6, 1, 1),
(6, 2, 150),
(6, 9, 30); -- Vanilla Syrup

-- Caramel Macchiato
INSERT INTO menu_item_inventory (menu_id, inventory_id, quantity) VALUES
(7, 1, 1),
(7, 2, 200),
(7, 10, 30); -- Caramel Syrup

-- Chocolate Frappe
INSERT INTO menu_item_inventory (menu_id, inventory_id, quantity) VALUES
(8, 11, 100), -- Chocolate Syrup
(8, 2, 100),
(8, 12, 50);  -- Whipped Cream

-- Matcha Latte
INSERT INTO menu_item_inventory (menu_id, inventory_id, quantity) VALUES
(9, 2, 150),     -- Milk
(9, 18, 10);     -- Matcha Powder

-- Chai Tea Latte
INSERT INTO menu_item_inventory (menu_id, inventory_id, quantity) VALUES
(10, 2, 150),    -- Milk
(10, 17, 5),     -- Cinnamon
(10, 18, 3);     -- Nutmeg

-- Barista
INSERT INTO menu_item_inventory (menu_id, inventory_id, quantity) VALUES
(11, 1, 2),    -- Double Espresso Shot
(11, 8, 20),   -- Ground Coffee
(11, 2, 100),  -- Milk
(11, 22, 20);  -- Hazelnut Syrup

--Ice latte
INSERT INTO menu_item_inventory (menu_id, inventory_id, quantity) VALUES
(12, 1, 1),    -- Espresso Shot
(12, 2, 150),  -- Milk
(12, 21, 10);  -- Ice Cubes

-- Double espresso
INSERT INTO menu_item_inventory (menu_id, inventory_id, quantity) VALUES
(13, 1, 2);


INSERT INTO orders (customer_name, order_status, created_at) VALUES
('Alice Johnson', 'open', '2025-02-05 09:52:00'),
('George Martin', 'open', '2025-02-26 08:16:00'),
('George Jackson', 'closed', '2025-01-16 07:59:00'),
('Fiona Taylor', 'open', '2025-01-07 16:25:00'),
('Hannah Martin', 'open', '2025-01-05 14:58:00'),
('Charlie Jackson', 'open', '2025-02-18 10:10:00'),
('Jane Martin', 'closed', '2025-02-13 11:51:00'),
('George Smith', 'open', '2025-01-02 10:25:00'),
('Diana Martin', 'open', '2025-01-13 17:00:00'),
('Bob Martin', 'open', '2025-01-28 06:49:00'),
('Ian Harris', 'open', '2025-01-14 20:13:00'),
('Charlie Brown', 'closed', '2025-01-04 13:59:00'),
('Hannah Taylor', 'open', '2025-01-07 07:15:00'),
('Alice Taylor', 'open', '2025-02-01 14:26:00'),
('Hannah White', 'open', '2025-02-28 08:34:00'),
('Bob Harris', 'open', '2025-01-04 14:13:00'),
('Jane Harris', 'open', '2025-01-16 17:39:00'),
('George Brown', 'closed', '2025-02-09 14:10:00'),
('Fiona Harris', 'closed', '2025-01-26 08:11:00'),
('Fiona Harris', 'closed', '2025-02-27 20:37:00'),
('Jane Thomas', 'closed', '2025-01-01 14:09:00'),
('Ian Jackson', 'closed', '2025-02-20 15:20:00'),
('Bob Brown', 'closed', '2025-01-26 18:04:00'),
('Ian Harris', 'closed', '2025-02-17 13:12:00'),
('Edward Smith', 'open', '2025-01-08 15:42:00'),
('Charlie Jackson', 'open', '2025-01-30 08:53:00'),
('Alice Harris', 'closed', '2025-01-27 17:50:00'),
('George Brown', 'open', '2025-02-02 13:28:00'),
('Alice Jackson', 'open', '2025-01-11 14:14:00'),
('Ian Smith', 'open', '2025-01-09 12:51:00'),
('Alice Brown', 'closed', '2025-01-09 07:19:00'),
('Ian Martin', 'open', '2025-02-08 13:54:00'),
('Bob Harris', 'closed', '2025-01-02 20:08:00'),
('George Brown', 'closed', '2025-01-03 18:47:00'),
('Edward Martin', 'closed', '2025-01-30 15:03:00'),
('Diana Jackson', 'open', '2025-02-23 15:11:00'),
('Fiona White', 'open', '2025-01-12 10:40:00'),
('George Jackson', 'open', '2025-02-10 07:39:00'),
('Charlie Anderson', 'closed', '2025-01-22 09:20:00'),
('Hannah Johnson', 'open', '2025-01-06 11:36:00'),
('George Thomas', 'open', '2025-02-12 16:18:00'),
('Jane Anderson', 'closed', '2025-01-31 08:00:00'),
('Edward Anderson', 'closed', '2025-02-15 14:30:00'),
('Bob Taylor', 'closed', '2025-02-14 19:09:00'),
('Diana Brown', 'open', '2025-01-25 13:17:00'),
('Fiona White', 'closed', '2025-02-04 10:23:00'),
('George Jackson', 'open', '2025-01-20 09:44:00'),
('Charlie White', 'closed', '2025-01-18 08:57:00'),
('Alice Anderson', 'closed', '2025-01-21 15:41:00'),
('Ian Johnson', 'closed', '2025-02-24 17:22:00'),
('Jane White', 'open', '2025-01-10 16:46:00'),
('Bob Anderson', 'open', '2025-01-17 12:07:00'),
('George Harris', 'closed', '2025-01-15 19:35:00'),
('Fiona Taylor', 'open', '2025-02-07 07:23:00'),
('Alice Martin', 'closed', '2025-02-06 11:10:00'),
('Ian Taylor', 'closed', '2025-02-03 14:18:00'),
('Charlie Martin', 'open', '2025-01-19 16:31:00'),
('George Thomas', 'open', '2025-01-23 08:42:00'),
('Alice Johnson', 'closed', '2025-01-24 10:56:00'),
('Hannah Jackson', 'closed', '2025-02-11 17:45:00');



INSERT INTO order_item (order_id, menu_item_id, quantity) VALUES
(1, 13, 2),
(1, 4, 3),
(1, 1, 2),
(2, 9, 4),
(2, 8, 3),
(3, 13, 5),
(3, 1, 4),
(3, 6, 1),
(4, 10, 5),
(4, 4, 3),
(5, 12, 3),
(5, 3, 1),
(5, 1, 3),
(6, 7, 5),
(6, 12, 2),
(7, 3, 2),
(7, 7, 5),
(8, 12, 5),
(9, 12, 5),
(9, 3, 3),
(10, 7, 1),
(11, 8, 4),
(11, 1, 2),
(12, 1, 4),
(13, 5, 4),
(14, 4, 3),
(15, 4, 5),
(16, 5, 1),
(16, 3, 4),
(17, 5, 1),
(17, 13, 2),
(18, 12, 1),
(19, 9, 2),
(20, 2, 4),
(20, 10, 1),
(21, 7, 2),
(22, 8, 4),
(22, 7, 2),
(22, 2, 5),
(23, 4, 4),
(24, 6, 4),
(25, 13, 1),
(25, 7, 2),
(26, 13, 1),
(26, 5, 4),
(27, 12, 3),
(28, 6, 5),
(28, 8, 4),
(28, 9, 4),
(29, 4, 5),
(29, 13, 2),
(29, 1, 3),
(30, 4, 2),
(30, 12, 3),
(31, 12, 5),
(31, 9, 3),
(31, 11, 3),
(32, 11, 5),
(33, 11, 1),
(33, 4, 4),
(34, 7, 4),
(34, 9, 5),
(35, 2, 2),
(35, 6, 2),
(35, 9, 5),
(36, 12, 4),
(37, 13, 3),
(37, 7, 2),
(37, 1, 3),
(38, 13, 5),
(39, 5, 4),
(39, 4, 1),
(40, 11, 3),
(40, 6, 2),
(40, 2, 4),
(41, 13, 2),
(42, 1, 5),
(43, 11, 5),
(43, 9, 2),
(43, 10, 2),
(44, 5, 1),
(44, 13, 1),
(45, 1, 3),
(46, 10, 5),
(47, 4, 4),
(47, 3, 4),
(48, 7, 2),
(49, 9, 5),
(50, 9, 4),
(50, 3, 1),
(51, 5, 1),
(51, 1, 4),
(52, 13, 4),
(52, 12, 2),
(53, 8, 1),
(53, 11, 4),
(54, 7, 4),
(55, 10, 1),
(55, 13, 1),
(55, 8, 4),
(56, 9, 3),
(57, 4, 2),
(57, 2, 4),
(58, 2, 5),
(58, 5, 2),
(58, 8, 5),
(59, 6, 3),
(59, 4, 5),
(60, 7, 3),
(60, 12, 2);

