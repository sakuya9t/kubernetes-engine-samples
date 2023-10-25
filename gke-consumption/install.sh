#!/bin/bash

# Create config file from template.
cp ./config/config_template.yml ./config/config.yml

# Require installation of sqlite3 and sqlite-utils
# Initilize local cache
sqlite3 cache.db < ./store/cache.sql
