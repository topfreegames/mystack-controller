-- mystack-controller api
-- https://github.com/topfreegames/mystack-controller
--
-- Licensed under the MIT license:
-- http://www.opensource.org/licenses/mit-license
-- Copyright © 2016 Top Free Games <backend@tfgco.com>

REVOKE ALL ON SCHEMA public FROM mystack_controller_test;
DROP DATABASE IF EXISTS mystack_controller_test;

DROP ROLE mystack_controller_test;

CREATE ROLE mystack_controller_test LOGIN
  SUPERUSER INHERIT CREATEDB CREATEROLE;

CREATE DATABASE mystack_controller_test
  WITH OWNER = mystack_controller_test
       ENCODING = 'UTF8'
       TABLESPACE = pg_default
       TEMPLATE = template0;

GRANT ALL ON SCHEMA public TO mystack_controller_test;
