-- mystack-controller api
-- https://github.com/topfreegames/mystack-controller
--
-- Licensed under the MIT license:
-- http://www.opensource.org/licenses/mit-license
-- Copyright © 2016 Top Free Games <backend@tfgco.com>

CREATE TABLE custom_domains (
    cluster varchar(255) NOT NULL,
    app varchar(255) NOT NULL,
    domains varchar(255)[],
    UNIQUE(cluster, app)
);
