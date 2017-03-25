-- kubecos api
-- https://github.com/topfreegames/kubecos
--
-- Licensed under the MIT license:
-- http://www.opensource.org/licenses/mit-license
-- Copyright © 2016 Top Free Games <backend@tfgco.com>

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id UUID PRIMARY KEY,
    email varchar(255) NOT NULL,
    created_at timestamp WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at timestamp WITH TIME ZONE NULL
);

