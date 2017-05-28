-- mystack-controller api
-- https://github.com/topfreegames/mystack-controller
--
-- Licensed under the MIT license:
-- http://www.opensource.org/licenses/mit-license
-- Copyright © 2016 Top Free Games <backend@tfgco.com>

CREATE TABLE users (
    access_token varchar(255) UNIQUE NOT NULL CHECK (access_token <> ''),
    key_access_token varchar(255) UNIQUE NOT NULL CHECK (key_access_token <> ''),
    refresh_token varchar(255) UNIQUE NOT NULL CHECK (refresh_token <> ''),
    expiry timestamp NOT NULL,
    token_type varchar(255) NOT NULL,
    email varchar(255) UNIQUE NOT NULL CHECK (email <> ''),
    created_at timestamp WITH TIME ZONE NOT NULL DEFAULT NOW()
);

