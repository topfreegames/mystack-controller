-- mystack-controller api
-- https://github.com/topfreegames/mystack-controller
--
-- Licensed under the MIT license:
-- http://www.opensource.org/licenses/mit-license
-- Copyright © 2016 Top Free Games <backend@tfgco.com>

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE clusters (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name varchar(255) UNIQUE NOT NULL,
    yaml TEXT NOT NULL,
    created_at timestamp WITH TIME ZONE NOT NULL DEFAULT NOW()
);
