#!/bin/bash
export DB_HOST
sed -i 's/DB_HOST=localhost//g' .env
make test
