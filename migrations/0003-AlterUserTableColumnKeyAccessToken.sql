-- mystack-controller api
-- https://github.com/topfreegames/mystack-controller
--
-- Licensed under the MIT license:
-- http://www.opensource.org/licenses/mit-license
-- Copyright © 2016 Top Free Games <backend@tfgco.com>

ALTER TABLE users ADD COLUMN key_access_token varchar(255) NOT NULL CHECK (key_access_token <> '') DEFAULT 'token';
UPDATE users SET key_access_token = access_token;
ALTER TABLE users ALTER COLUMN key_access_token DROP DEFAULT;
CREATE UNIQUE INDEX ON users (key_access_token);
