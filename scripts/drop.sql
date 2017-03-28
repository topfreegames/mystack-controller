-- mystack-controller api
-- https://github.com/topfreegames/mystack/mystack-controller
--
-- Licensed under the MIT license:
-- http://www.opensource.org/licenses/mit-license
-- Copyright © 2016 Top Free Games <backend@tfgco.com>

REVOKE ALL ON SCHEMA public FROM mystack_controller;
DROP DATABASE IF EXISTS mystack_controller;

DROP ROLE mystack_controller;

CREATE ROLE mystack_controller LOGIN
  SUPERUSER INHERIT CREATEDB CREATEROLE;

CREATE DATABASE mystack_controller
  WITH OWNER = mystack_controller
       ENCODING = 'UTF8'
       TABLESPACE = pg_default
       TEMPLATE = template0;

GRANT ALL ON SCHEMA public TO mystack_controller;
