-- kubecos api
-- https://github.com/topfreegames/kubecos
--
-- Licensed under the MIT license:
-- http://www.opensource.org/licenses/mit-license
-- Copyright © 2016 Top Free Games <backend@tfgco.com>

REVOKE ALL ON SCHEMA public FROM kubecos;
DROP DATABASE IF EXISTS kubecos;

DROP ROLE kubecos;

CREATE ROLE kubecos LOGIN
  SUPERUSER INHERIT CREATEDB CREATEROLE;

CREATE DATABASE kubecos
  WITH OWNER = kubecos
       ENCODING = 'UTF8'
       TABLESPACE = pg_default
       TEMPLATE = template0;

GRANT ALL ON SCHEMA public TO kubecos;
