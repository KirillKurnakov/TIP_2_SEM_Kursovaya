--
-- PostgreSQL database dump
--

-- Dumped from database version 16.4
-- Dumped by pg_dump version 16.4

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: category; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.category (
    category_id character varying(2) NOT NULL,
    name character varying(40) NOT NULL
);


ALTER TABLE public.category OWNER TO postgres;

--
-- Name: delivery; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.delivery (
    deliverytype_id character varying(5) NOT NULL,
    name character varying(50) NOT NULL,
    numberkilometer double precision NOT NULL,
    price double precision NOT NULL,
    timedelivery time without time zone NOT NULL
);


ALTER TABLE public.delivery OWNER TO postgres;

--
-- Name: items; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.items (
    item_id character varying(5) NOT NULL,
    name text,
    category_id text,
    description text,
    price bigint,
    quantity_stock integer
);


ALTER TABLE public.items OWNER TO postgres;

--
-- Name: order; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public."order" (
    order_id character varying(5) NOT NULL,
    user_id character varying(5),
    orderdate timestamp without time zone NOT NULL,
    deliverytype_id character varying(5),
    addressdelivery character varying(50) NOT NULL,
    numberkilometer double precision NOT NULL,
    deliveryprice double precision NOT NULL,
    finalprice double precision NOT NULL,
    orderstatus_id character varying(5)
);


ALTER TABLE public."order" OWNER TO postgres;

--
-- Name: order_items; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.order_items (
    orderitem_id character varying(5) NOT NULL,
    order_id text,
    itemid text,
    quantity bigint,
    sum_price_item bigint,
    package_id text,
    quantity_package bigint,
    sum_price_package numeric
);


ALTER TABLE public.order_items OWNER TO postgres;

--
-- Name: orderstatus; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.orderstatus (
    orderstatus_id character varying(5) NOT NULL,
    name character varying(50) NOT NULL
);


ALTER TABLE public.orderstatus OWNER TO postgres;

--
-- Name: package; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.package (
    package_id character varying(5) NOT NULL,
    name character varying(50) NOT NULL,
    description character varying(100) NOT NULL,
    price double precision NOT NULL
);


ALTER TABLE public.package OWNER TO postgres;

--
-- Name: payment; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.payment (
    payment_id character varying(5) NOT NULL,
    order_id character varying(5),
    paymentdate timestamp without time zone NOT NULL,
    paymenttype_id character varying(5),
    finalprice double precision NOT NULL,
    confirmation boolean NOT NULL
);


ALTER TABLE public.payment OWNER TO postgres;

--
-- Name: paymenttype; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.paymenttype (
    paymenttype_id character varying(5) NOT NULL,
    name character varying(50) NOT NULL
);


ALTER TABLE public.paymenttype OWNER TO postgres;

--
-- Name: review; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.review (
    review_id character varying(5) NOT NULL,
    rating double precision NOT NULL,
    comment character varying(100) NOT NULL,
    reviewdate timestamp without time zone NOT NULL,
    anonymous boolean NOT NULL,
    item_id character varying(5)
);


ALTER TABLE public.review OWNER TO postgres;

--
-- Name: user; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public."user" (
    user_id character varying(5) NOT NULL,
    name character varying(50) NOT NULL,
    lastname character varying(50) NOT NULL,
    surname character varying(50) NOT NULL,
    email character varying(50) NOT NULL,
    phone character varying(18) NOT NULL,
    usertype_id character varying(5)
);


ALTER TABLE public."user" OWNER TO postgres;

--
-- Name: usertype; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.usertype (
    usertype_id character varying(5) NOT NULL,
    name character varying(50) NOT NULL
);


ALTER TABLE public.usertype OWNER TO postgres;

--
-- Data for Name: category; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.category (category_id, name) FROM stdin;
C1	Мягкая игрушка
C2	Кружка
C3	Кофта
\.


--
-- Data for Name: delivery; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.delivery (deliverytype_id, name, numberkilometer, price, timedelivery) FROM stdin;
D001	Стандартная доставка	5	300	02:00:00
D002	Экспресс-доставка	10	600	01:00:00
D003	Самовывоз	0	0	00:30:00
\.


--
-- Data for Name: items; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.items (item_id, name, category_id, description, price, quantity_stock) FROM stdin;
I003	Кошка	C1	Мягкая игрушка в виде кошки	850	15
I004	Птичка	C1	Мягкая игрушка в виде птички	850	15
I005	Капибара	C1	Мягкая игрушка в виде капибары	850	15
I006	Кружка с мишкой	C2	Прикольная кружка с мишкой	450	10
I007	Кружка с собачкой	C2	Прикольная кружка с собачкой	450	10
I008	Кружка с кошкой	C2	Прикольная кружка с кошкой	450	10
I009	Кружка с птичкой	C2	Прикольная кружка с птичкой	450	10
I0010	Кружка с капибарой	C2	Прикольная кружка с капибарой	450	10
I0012	Кофта с собачкой	C3	Крутая кофта с собачкой	850	30
I0013	Кофта с кошкой	C3	Крутая кофта с кошкой	850	30
I0014	Кофта с птичкой	C3	Крутая кофта с птичкой	850	30
I0015	Кофта с капибарой	C3	Крутая кофта с капибарой	850	30
I0016	Мишка	C1	Мягкая игрушка в виде мишки	850	15
I001	Мишка	C1	Красивая игрушка	850	15
\.


--
-- Data for Name: order; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public."order" (order_id, user_id, orderdate, deliverytype_id, addressdelivery, numberkilometer, deliveryprice, finalprice, orderstatus_id) FROM stdin;
O001	U001	2024-11-05 10:00:00	D001	ул. Ленина, д. 1	5	300	55300	OS02
O002	U002	2024-11-05 12:00:00	D002	ул. Советская, д. 15	10	600	45000	OS02
O003	U003	2024-11-05 14:00:00	D003	ул. Мира, д. 8	0	0	14000	OS03
O004	U004	2024-11-05 16:00:00	D002	ул. Победы, д. 23	7	400	300	OS03
O005	U005	2024-11-05 18:00:00	D001	ул. Гагарина, д. 42	20	1500	9500	OS01
\.


--
-- Data for Name: order_items; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.order_items (orderitem_id, order_id, itemid, quantity, sum_price_item, package_id, quantity_package, sum_price_package) FROM stdin;
OI002	O002	I006	1	450	PK005	1	450
OI003	O003	I008	1	850	PK001	1	950
OI004	O004	I0012	2	1700	PK003	1	1720
OI005	O005	I0013	1	850	PK005	1	850
OI006	O001	I006	3	850	PK003	1	15
\.


--
-- Data for Name: orderstatus; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.orderstatus (orderstatus_id, name) FROM stdin;
OS01	Ожидает оплаты
OS02	Подтвержден
OS03	Доставлен
OS04	Отменен
OS05	Возврат средств
\.


--
-- Data for Name: package; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.package (package_id, name, description, price) FROM stdin;
PK001	Коробка стандартная	Картонная коробка для транспортировки	100
PK002	Упаковка премиум	Премиум-упаковка с защитой	500
PK003	Полиэтиленовый пакет	Простая упаковка	20
PK004	Подарочная упаковка	Красочная подарочная упаковка	700
PK005	Без упаковки	Доставка без упаковки	0
\.


--
-- Data for Name: payment; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.payment (payment_id, order_id, paymentdate, paymenttype_id, finalprice, confirmation) FROM stdin;
\.


--
-- Data for Name: paymenttype; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.paymenttype (paymenttype_id, name) FROM stdin;
PT01	Банковская карта
PT02	Наличные
PT03	СБП
\.


--
-- Data for Name: review; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.review (review_id, rating, comment, reviewdate, anonymous, item_id) FROM stdin;
R001	5	Крутая игрушка!	2024-11-01 10:00:00	f	I001
R002	4.5	Крутая крушка в подарок!	2024-11-02 12:00:00	t	I007
R003	5	Мягкая и удобная кофта	2024-11-03 15:30:00	f	I0014
R004	4	Ребенок рад!	2024-11-04 18:45:00	t	I004
R005	5	Кружка топ!	2024-11-05 20:00:00	f	I007
\.


--
-- Data for Name: user; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public."user" (user_id, name, lastname, surname, email, phone, usertype_id) FROM stdin;
U001	Иван	Иванов	Сергеевич	ivan.ivanov@example.com	89001234567	UT01
U002	Ольга	Смирнова	Викторовна	olga.smirnova@example.com	89007654321	UT01
U003	Анна	Кузнецова	Алексеевна	anna.kuznetsova@example.com	89006543210	UT02
U004	Дмитрий	Петров	Иванович	dmitriy.petrov@example.com	89005432109	UT03
U005	Мария	Васильева	Олеговна	maria.vasileva@example.com	89004321098	UT03
\.


--
-- Data for Name: usertype; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.usertype (usertype_id, name) FROM stdin;
UT01	Клиент
UT02	Администратор
UT03	Менеджер
\.


--
-- Name: category category_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.category
    ADD CONSTRAINT category_pkey PRIMARY KEY (category_id);


--
-- Name: delivery delivery_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.delivery
    ADD CONSTRAINT delivery_pkey PRIMARY KEY (deliverytype_id);


--
-- Name: items item_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.items
    ADD CONSTRAINT item_pkey PRIMARY KEY (item_id);


--
-- Name: order order_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."order"
    ADD CONSTRAINT order_pkey PRIMARY KEY (order_id);


--
-- Name: order_items orderitem_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT orderitem_pkey PRIMARY KEY (orderitem_id);


--
-- Name: orderstatus orderstatus_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.orderstatus
    ADD CONSTRAINT orderstatus_pkey PRIMARY KEY (orderstatus_id);


--
-- Name: package package_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.package
    ADD CONSTRAINT package_pkey PRIMARY KEY (package_id);


--
-- Name: payment payment_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.payment
    ADD CONSTRAINT payment_pkey PRIMARY KEY (payment_id);


--
-- Name: paymenttype paymenttype_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.paymenttype
    ADD CONSTRAINT paymenttype_pkey PRIMARY KEY (paymenttype_id);


--
-- Name: review review_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.review
    ADD CONSTRAINT review_pkey PRIMARY KEY (review_id);


--
-- Name: user user_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."user"
    ADD CONSTRAINT user_pkey PRIMARY KEY (user_id);


--
-- Name: usertype usertype_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.usertype
    ADD CONSTRAINT usertype_pkey PRIMARY KEY (usertype_id);


--
-- Name: items item_category_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.items
    ADD CONSTRAINT item_category_id_fkey FOREIGN KEY (category_id) REFERENCES public.category(category_id);


--
-- Name: order order_deliverytype_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."order"
    ADD CONSTRAINT order_deliverytype_id_fkey FOREIGN KEY (deliverytype_id) REFERENCES public.delivery(deliverytype_id);


--
-- Name: order order_orderstatus_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."order"
    ADD CONSTRAINT order_orderstatus_id_fkey FOREIGN KEY (orderstatus_id) REFERENCES public.orderstatus(orderstatus_id);


--
-- Name: order order_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."order"
    ADD CONSTRAINT order_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(user_id);


--
-- Name: order_items orderitem_item_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT orderitem_item_id_fkey FOREIGN KEY (itemid) REFERENCES public.items(item_id);


--
-- Name: order_items orderitem_order_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT orderitem_order_id_fkey FOREIGN KEY (order_id) REFERENCES public."order"(order_id);


--
-- Name: order_items orderitem_package_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT orderitem_package_id_fkey FOREIGN KEY (package_id) REFERENCES public.package(package_id);


--
-- Name: payment payment_order_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.payment
    ADD CONSTRAINT payment_order_id_fkey FOREIGN KEY (order_id) REFERENCES public."order"(order_id);


--
-- Name: payment payment_paymenttype_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.payment
    ADD CONSTRAINT payment_paymenttype_id_fkey FOREIGN KEY (paymenttype_id) REFERENCES public.paymenttype(paymenttype_id);


--
-- Name: review review_item_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.review
    ADD CONSTRAINT review_item_id_fkey FOREIGN KEY (item_id) REFERENCES public.items(item_id);


--
-- Name: user user_usertype_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."user"
    ADD CONSTRAINT user_usertype_id_fkey FOREIGN KEY (usertype_id) REFERENCES public.usertype(usertype_id);


--
-- PostgreSQL database dump complete
--

