#!/bin/bash

# Start PHP server in the background
php -S localhost:8000 server.php &
PHP_PID=$!

# Start Vite dev server
npm run dev

# When Vite server stops, kill PHP server
kill $PHP_PID 