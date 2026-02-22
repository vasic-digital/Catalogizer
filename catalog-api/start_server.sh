#!/bin/bash
export ADMIN_USERNAME=admin
export ADMIN_PASSWORD=admin123
export JWT_SECRET="test-secret-1234567890-abcdefghijklmnop"
export ENABLE_AUTH=true
export DATABASE_TYPE=sqlite
export DB_PATH=./data/catalogizer.db
exec ./test.bin -test-mode
